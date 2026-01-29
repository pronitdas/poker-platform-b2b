package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"poker-platform/internal/game"
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
	tables   map[string]*game.Table
	rng      *rng.System
	upgrader websocket.Upgrader
}

func NewGameServer() (*GameServer, error) {
	rngSystem, err := rng.NewSystem(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RNG: %w", err)
	}

	return &GameServer{
		tables:   make(map[string]*game.Table),
		rng:      rngSystem,
		upgrader: upgrader,
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
		config := game.TableConfig{
			TableID:       tableID,
			MaxPlayers:    9,
			MinPlayers:    2,
			SmallBlind:    5,
			BigBlind:      10,
			BuyInMin:      100,
			BuyInMax:      10000,
			ActionTimeout: 30 * time.Second,
		}
		table = game.NewTable(config)
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
			"type": "joined",
			"state": table.GetState(),
		})

	case "action":
		playerID := msg["player_id"].(string)
		actionType := msg["action"].(string)
		amount := int64(msg["amount"].(float64))

		action := game.PlayerActionRequest{
			PlayerID: playerID,
			Action:   parseAction(actionType),
			Amount:   amount,
		}

		if err := table.SubmitAction(context.Background(), action); err != nil {
			s.sendError(conn, err.Error())
			return
		}

	case "leave":
		playerID := msg["player_id"].(string)
		if err := table.PlayerLeaves(playerID); err != nil {
			s.sendError(conn, err.Error())
		}
	}
}

func parseAction(action string) game.PlayerAction {
	switch action {
	case "fold":
		return game.ActionFold
	case "check":
		return game.ActionCheck
	case "call":
		return game.ActionCall
	case "bet":
		return game.ActionBet
	case "raise":
		return game.ActionRaise
	case "all_in":
		return game.ActionAllIn
	default:
		return game.ActionFold
	}
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

		config := game.TableConfig{
			TableID:       req.TableID,
			MaxPlayers:    9,
			MinPlayers:    2,
			SmallBlind:    5,
			BigBlind:      10,
			BuyInMin:      100,
			BuyInMax:      10000,
			ActionTimeout: 30 * time.Second,
		}
		table := game.NewTable(config)
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

import "encoding/json"
