package rules

import (
	"fmt"
	"sync"
)

// EngineRegistry manages registered rules engines
type EngineRegistry struct {
	engines map[GameType]RulesEngine
	mu      sync.RWMutex
}

var (
	registry *EngineRegistry
	once     sync.Once
)

// GetRegistry returns the singleton engine registry
func GetRegistry() *EngineRegistry {
	once.Do(func() {
		registry = &EngineRegistry{
			engines: make(map[GameType]RulesEngine),
		}
		// Register default engines
		registry.Register(NewTexasHoldem())
		registry.Register(NewOmaha())
	})
	return registry
}

// Register registers a rules engine
func (r *EngineRegistry) Register(engine RulesEngine) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if engine == nil {
		return fmt.Errorf("cannot register nil engine")
	}

	gameType := engine.GameType()
	if _, exists := r.engines[gameType]; exists {
		return fmt.Errorf("engine for game type %s already registered", gameType)
	}

	r.engines[gameType] = engine
	return nil
}

// Unregister removes a rules engine
func (r *EngineRegistry) Unregister(gameType GameType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.engines, gameType)
}

// Get retrieves a rules engine for a game type
func (r *EngineRegistry) Get(gameType GameType) (RulesEngine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	engine, exists := r.engines[gameType]
	if !exists {
		return nil, fmt.Errorf("no engine registered for game type %s", gameType)
	}

	return engine, nil
}

// GetByName retrieves a rules engine by name (case-insensitive)
func (r *EngineRegistry) GetByName(name string) (RulesEngine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, engine := range r.engines {
		if engine.Name() == name {
			return engine, nil
		}
	}

	return nil, fmt.Errorf("no engine found with name %s", name)
}

// List returns all registered game types
func (r *EngineRegistry) List() []GameType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]GameType, 0, len(r.engines))
	for gameType := range r.engines {
		types = append(types, gameType)
	}
	return types
}

// ListEngines returns all registered engines
func (r *EngineRegistry) ListEngines() []RulesEngine {
	r.mu.RLock()
	defer r.mu.RUnlock()

	engines := make([]RulesEngine, 0, len(r.engines))
	for _, engine := range r.engines {
		engines = append(engines, engine)
	}
	return engines
}

// CreateEngine creates a new rules engine instance for a game type
func (r *EngineRegistry) CreateEngine(gameType GameType) (RulesEngine, error) {
	// Return a new instance based on game type
	switch gameType {
	case GameTypeTexasHoldem:
		return NewTexasHoldem(), nil
	case GameTypeOmaha:
		return NewOmaha(), nil
	case GameTypeOmahaHiLo:
		return NewOmahaHiLo(), nil
	case GameTypeFiveCardDraw:
		return NewFiveCardDraw(), nil
	case GameTypeSevenCardStud:
		return NewSevenCardStud(), nil
	default:
		return nil, fmt.Errorf("unsupported game type %s", gameType)
	}
}

// EngineManager provides additional management functionality
type EngineManager struct {
	registry     *EngineRegistry
	activeEngine RulesEngine
	config       TableConfig
	mu           sync.RWMutex
}

// NewEngineManager creates a new engine manager
func NewEngineManager(gameType GameType, config TableConfig) (*EngineManager, error) {
	registry := GetRegistry()

	engine, err := registry.CreateEngine(gameType)
	if err != nil {
		return nil, err
	}

	// Validate and apply config
	if err := engine.ValidateConfig(config); err != nil {
		return nil, err
	}

	if config.GameType == "" {
		config.GameType = gameType
	}

	return &EngineManager{
		registry:     registry,
		activeEngine: engine,
		config:       config,
	}, nil
}

// SetEngine switches to a different rules engine
func (m *EngineManager) SetEngine(gameType GameType, config TableConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	engine, err := m.registry.CreateEngine(gameType)
	if err != nil {
		return err
	}

	if err := engine.ValidateConfig(config); err != nil {
		return err
	}

	m.activeEngine = engine
	m.config = config
	return nil
}

// GetEngine returns the currently active rules engine
func (m *EngineManager) GetEngine() RulesEngine {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.activeEngine
}

// GetConfig returns the current table configuration
func (m *EngineManager) GetConfig() TableConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config
}

// UpdateConfig updates the table configuration
func (m *EngineManager) UpdateConfig(config TableConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.activeEngine.ValidateConfig(config); err != nil {
		return err
	}

	m.config = config
	return nil
}

// GameType returns the current game type
func (m *EngineManager) GameType() GameType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.activeEngine.GameType()
}

// Name returns the current game name
func (m *EngineManager) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.activeEngine.Name()
}

// SupportedGames returns all supported game types
func SupportedGames() []string {
	return []string{
		"texas holdem",
		"texas_hold'em",
		"omaha",
		"omaha hi lo",
		"omaha_hi_lo",
		"five card draw",
		"five_card_draw",
		"seven card stud",
		"seven_card_stud",
	}
}

// ParseGameType parses a string to GameType
func ParseGameType(s string) (GameType, error) {
	switch s {
	case "texas holdem", "texas_hold'em", "holdem":
		return GameTypeTexasHoldem, nil
	case "omaha":
		return GameTypeOmaha, nil
	case "omaha hi lo", "omaha_hi_lo", "omaha8":
		return GameTypeOmahaHiLo, nil
	case "five card draw", "five_card_draw", "fivedraw":
		return GameTypeFiveCardDraw, nil
	case "seven card stud", "seven_card_stud", "7stud":
		return GameTypeSevenCardStud, nil
	default:
		return "", fmt.Errorf("unknown game type: %s", s)
	}
}
