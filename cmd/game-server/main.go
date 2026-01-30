package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"poker-platform/internal/fraud"
	"poker-platform/internal/game"
	"poker-platform/internal/game/rules"
	"poker-platform/pkg/rng"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// GameServer manages WebSocket connections for poker tables
type GameServer struct {
	tables       map[string]*game.Table
	rng          *rng.System
	upgrader     websocket.Upgrader
	fraudService *fraud.FraudService
	mu           sync.RWMutex
}

func NewGameServer() (*GameServer, error) {
	rngSystem, err := rng.NewSystem(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RNG: %w", err)
	}

	// Initialize fraud detection components
	botDetector := fraud.NewBotDetector(nil)
	collusionDetector := fraud.NewCollusionDetector(fraud.DefaultCollusionDetectionConfig())
	multiAccountDetector := fraud.NewMultiAccountDetector(
		fraud.DefaultMultiAccountConfig(),
		nil, // fingerprintDB
		nil, // ipTracker
		nil, // sessionStore
	)
	ruleEngine := fraud.NewRuleEngine(fraud.NewRuleBasedDetector(), nil)

	alertService := fraud.NewAlertService(nil, nil, nil)
	riskScorer := fraud.NewRiskScorer(
		fraud.DefaultRiskScoringConfig(),
		botDetector,
		collusionDetector,
		multiAccountDetector,
		ruleEngine,
		nil, // alertStorage
	)

	fraudConfig := fraud.DefaultFraudServiceConfig()
	fraudService := fraud.NewFraudService(
		fraudConfig,
		botDetector,
		collusionDetector,
		multiAccountDetector,
		ruleEngine,
		riskScorer,
		alertService,
	)

	return &GameServer{
		tables:       make(map[string]*game.Table),
		rng:          rngSystem,
		upgrader:     upgrader,
		fraudService: fraudService,
	}, nil
}

func (s *GameServer) handleWebSocket(c *gin.Context) {
	tableID := c.Param("tableId")
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Player connected to table: %s", tableID)

	// Get or create table
	table, exists := s.tables[tableID]
	if !exists {
		config := rules.TableConfig{
			TableID:       tableID,
			GameType:      rules.GameTypeTexasHoldem,
			BettingType:   rules.BettingTypeNoLimit,
			MaxPlayers:    9,
			MinPlayers:    2,
			SmallBlind:    5,
			BigBlind:      10,
			BuyInMin:      100,
			BuyInMax:      10000,
			ActionTimeout: 30 * time.Second,
		}
		var err error
		table, err = game.NewTable(config)
		if err != nil {
			log.Printf("Failed to create table: %v", err)
			return
		}
		table.Start(context.Background())
		s.tables[tableID] = table
	}

	// Handle messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		s.handleMessage(conn, table, msg)
	}
}

func (s *GameServer) handleMessage(conn *websocket.Conn, table *game.Table, msg map[string]interface{}) {
	switch msg["type"] {
	case "join":
		playerID := msg["player_id"].(string)
		playerName := msg["player_name"].(string)
		chips := int64(msg["chips"].(float64))

		if err := table.PlayerJoins(playerID, playerName, chips); err != nil {
			s.sendError(conn, err.Error())
			return
		}
		s.sendMessage(conn, map[string]interface{}{
			"type":  "joined",
			"state": table.GetState(),
		})

	case "action":
		playerID := msg["player_id"].(string)
		actionType := msg["action"].(string)
		amount := int64(msg["amount"].(float64))

		action := rules.PlayerActionRequest{
			PlayerID: playerID,
			Action:   parseAction(actionType),
			Amount:   amount,
		}

		if err := table.SubmitAction(context.Background(), action); err != nil {
			s.sendError(conn, err.Error())
			return
		}

		// Get table ID for fraud detection
		tableID := table.GetState().TableID

		// Send to fraud detection service (non-blocking)
		go func() {
			fraudAction := s.createFraudAction(playerID, tableID, action, msg)
			if fraudAction != nil {
				result, err := s.fraudService.ProcessPlayerAction(context.Background(), fraudAction)
				if err != nil {
					log.Printf("Fraud detection error for player %s: %v", playerID, err)
					return
				}
				if result != nil && result.RequiresAction {
					log.Printf("Fraud alert for player %s: %v", playerID, result.RecommendedActions)
					// Send fraud alert to client
					s.sendMessage(conn, map[string]interface{}{
						"type":    "fraud_alert",
						"message": "Suspicious activity detected",
						"actions": result.RecommendedActions,
					})
				}
			}
		}()

	case "leave":
		playerID := msg["player_id"].(string)
		if err := table.PlayerLeaves(playerID); err != nil {
			s.sendError(conn, err.Error())
		}
	}
}

func parseAction(action string) rules.PlayerAction {
	switch action {
	case "fold":
		return rules.ActionFold
	case "check":
		return rules.ActionCheck
	case "call":
		return rules.ActionCall
	case "bet":
		return rules.ActionBet
	case "raise":
		return rules.ActionRaise
	case "all_in":
		return rules.ActionAllIn
	default:
		return rules.ActionFold
	}
}

// createFraudAction converts a game action to a fraud detection action
func (s *GameServer) createFraudAction(playerID, tableID string, action rules.PlayerActionRequest, msg map[string]interface{}) *fraud.PlayerAction {
	decisionTime := 0
	if dt, ok := msg["decision_time_ms"].(float64); ok {
		decisionTime = int(dt)
	}

	var potSize, stackSize int64
	if ps, ok := msg["pot_size"].(float64); ok {
		potSize = int64(ps)
	}
	if ss, ok := msg["stack_size"].(float64); ok {
		stackSize = int64(ss)
	}

	position := 0
	if pos, ok := msg["position"].(float64); ok {
		position = int(pos)
	}

	return &fraud.PlayerAction{
		ID:           fmt.Sprintf("action_%d", time.Now().UnixNano()),
		PlayerID:     playerID,
		TableID:      tableID,
		AgentID:      "", // Would be set from connection context
		ClubID:       "", // Would be set from connection context
		ActionType:   action.Action.String(),
		Amount:       action.Amount,
		Position:     position,
		Timestamp:    time.Now(),
		DecisionTime: decisionTime,
		HandPhase:    getHandPhase(action.Action),
		PotSize:      potSize,
		StackSize:    stackSize,
		IPAddress:    "", // Would be extracted from connection
		DeviceID:     "", // Would be extracted from connection
	}
}

// getHandPhase maps game phase to fraud detection hand phase
func getHandPhase(action rules.PlayerAction) string {
	// In production, this would come from the table state
	return "unknown"
}

func (s *GameServer) sendMessage(conn *websocket.Conn, data interface{}) {
	if err := conn.WriteJSON(data); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (s *GameServer) sendError(conn *websocket.Conn, message string) {
	s.sendMessage(conn, map[string]interface{}{
		"type":    "error",
		"message": message,
	})
}

func main() {
	router := gin.Default()

	server, err := NewGameServer()
	if err != nil {
		log.Fatalf("Failed to create game server: %v", err)
	}

	// WebSocket endpoint for game tables
	router.GET("/ws/:tableId", server.handleWebSocket)

	// REST API for table management
	router.GET("/api/tables/:tableId", func(c *gin.Context) {
		tableID := c.Param("tableId")
		table, exists := server.tables[tableID]
		if !exists {
			c.JSON(404, gin.H{"error": "Table not found"})
			return
		}
		c.JSON(200, table.GetState())
	})

	router.POST("/api/tables", func(c *gin.Context) {
		var req struct {
			TableID string `json:"tableId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		config := rules.TableConfig{
			TableID:       req.TableID,
			GameType:      rules.GameTypeTexasHoldem,
			BettingType:   rules.BettingTypeNoLimit,
			MaxPlayers:    9,
			MinPlayers:    2,
			SmallBlind:    5,
			BigBlind:      10,
			BuyInMin:      100,
			BuyInMax:      10000,
			ActionTimeout: 30 * time.Second,
		}
		table, err := game.NewTable(config)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create table: %v", err)})
			return
		}
		table.Start(context.Background())
		server.tables[req.TableID] = table
		c.JSON(201, gin.H{"tableId": req.TableID})
	})

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		for _, table := range server.tables {
			table.Stop()
		}
		os.Exit(0)
	}()

	port := os.Getenv("GAME_SERVER_PORT")
	if port == "" {
		port = "3002"
	}

	log.Printf("Game server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
