# Section 2: Core Modules Breakdown

## 2.1 Player Mobile Application (Cocos Creator)

The player mobile application serves as the primary client interface for end-users, built with Cocos Creator 3.8+ to deliver a lightweight, responsive poker experience across iOS and Android devices.

### Module Overview Table

| Module | Effort | Complexity | Key Features | Dependencies |
|--------|--------|------------|--------------|-------------|
| **2.1.1 Game Client Core** | 6 weeks | Very High | WebSocket connection, state synchronization, event handling | Socket.IO client, Game Engine API |
| **2.1.2 UI/UX Components** | 8 weeks | Medium-High | Tables, cards, avatars, animations, responsive layouts | Cocos Creator UI system, Asset pipeline |
| **2.1.3 Real-Time Communication** | 5 weeks | High | Auto-reconnection, fallback handling, room management | Socket.IO v4, Network adapter |
| **2.1.4 Audio & Visual Effects** | 4 weeks | Medium | Card sounds, dealer animations, celebration effects | Cocos Creator audio system, Particle effects |
| **2.1.5 Cross-Platform Build System** | 3 weeks | Medium | iOS/Android builds, code signing, app store submission | Cocos Creator build tools, Xcode, Android Studio |

### 2.1.1 Game Client Core (6 weeks, Very High Complexity)

**Description**: Core client-side logic handling game state synchronization, player actions, and real-time updates from the game server.

**Key Features**:
- Server-authoritative state management (all game logic validated on server)
- Optimistic UI updates for instant feedback
- State reconciliation when server response differs from client prediction
- Action queue for handling network-latency scenarios

**Implementation Notes**:

```typescript
// Cocos Creator TypeScript - State synchronization pattern
@ccclass('GameClientCore')
export class GameClientCore extends Component {
    private socket: Socket | null = null;
    private localState: TableState | null = null;
    private serverState: TableState | null = null;
    private actionQueue: PlayerAction[] = [];

    connectToTable(tableId: string, authToken: string) {
        this.socket = io(`${GAME_SERVER_URL}/${tableId}`, {
            auth: { token: authToken },
            transports: ['websocket', 'polling']
        });

        // Subscribe to table events
        this.socket.on('gameStateUpdate', (state: TableState) => {
            this.serverState = state;
            this.reconcileState();
        });

        this.socket.on('connect', () => {
            console.log('Connected to table:', tableId);
            this.syncInitial();
        });

        this.socket.on('reconnect', () => {
            console.log('Reconnected - syncing state');
            this.syncInitial();
        });
    }

    sendAction(action: PlayerAction) {
        // Optimistic update for immediate UI feedback
        this.actionQueue.push(action);
        this.applyActionLocally(action);

        // Send to server for validation
        this.socket?.emit('playerAction', action);
    }

    reconcileState() {
        // Apply server-authoritative state
        if (this.serverState) {
            this.localState = JSON.parse(JSON.stringify(this.serverState));
            this.renderTable();
        }
    }
}
```

**Performance Targets**:
- WebSocket connection establishment: <500ms (P95)
- Action round-trip latency (client → server → broadcast): <100ms (P99)
- State reconciliation: <16ms (60 FPS maintenance)

---

### 2.1.2 UI/UX Components (8 weeks, Medium-High Complexity)

**Description**: Comprehensive UI system for poker tables, player interfaces, and game elements, optimized for mobile devices.

**Key Features**:
- Responsive table layouts (adapting to 2-9 players)
- Card rendering with animations (deal, flip, reveal)
- Avatar system with emotional states
- Betting slider with preset buttons
- Chat system with emoji support
- Multi-language support (i18n)

**Implementation Notes**:

```typescript
// Cocos Creator Component pattern for UI elements
@ccclass('PokerTable')
export class PokerTable extends Component {
    @property({type: Prefab})
    private cardPrefab: Prefab | null = null;

    @property({type: Node})
    private playerSeats: Node[] = [];

    private readonly MAX_PLAYERS: number = 9;
    private readonly CARD_ANIMATION_DURATION: number = 0.3; // seconds

    renderTable(state: TableState) {
        // Clear existing cards
        this.clearTable();

        // Render player hands (only visible cards)
        state.players.forEach((player, index) => {
            if (player.hand.visible) {
                player.hand.cards.forEach(card => {
                    this.createCard(card, this.playerSeats[index]);
                });
            }
        });

        // Render community cards
        state.communityCards.forEach(card => {
            this.createCard(card, this.communityNode);
        });

        // Render pot and dealer button
        this.updatePot(state.pot);
        this.updateDealerButton(state.dealerPosition);
    }

    createCard(card: Card, parentNode: Node) {
        const cardNode = instantiate(this.cardPrefab!);
        const cardComponent = cardNode.getComponent(CardComponent);
        cardComponent.setCard(card);
        parentNode.addChild(cardNode);

        // Animate card entry
        tween(cardNode)
            .to(this.CARD_ANIMATION_DURATION, { scale: new Vec3(1, 1, 1) })
            .call(() => {
                // Play sound effect
                this.audioManager.playCardDeal();
            })
            .start();
    }
}
```

**Performance Optimizations**:
- Object pooling for card prefabs (reduce garbage collection)
- Sprite batching for multiple cards
- Lazy loading of assets for table backgrounds
- Texture compression for mobile (ASTC/ETC2 formats)

---

### 2.1.3 Real-Time Communication (5 weeks, High Complexity)

**Description**: Robust WebSocket communication layer with automatic reconnection, fallback mechanisms, and network resilience.

**Key Features**:
- Socket.IO v4 integration with auto-reconnection
- Exponential backoff on connection failures
- Fallback to HTTP long-polling if WebSocket fails
- Connection quality monitoring and adaptive behavior
- Message queuing for offline scenarios

**Implementation Notes**:

```typescript
// Socket.IO connection management
class SocketManager {
    private socket: Socket | null = null;
    private reconnectAttempts: number = 0;
    private readonly MAX_RECONNECT_ATTEMPTS = 10;
    private readonly BASE_RECONNECT_DELAY = 1000; // ms

    connect(serverUrl: string, tableId: string, token: string) {
        this.socket = io(serverUrl, {
            path: '/socket.io/',
            transports: ['websocket', 'polling'], // Fallback to polling
            auth: { token, tableId },
            reconnection: true,
            reconnectionDelay: this.calculateReconnectDelay(),
            reconnectionAttempts: this.MAX_RECONNECT_ATTEMPTS,
            timeout: 10000 // 10 seconds
        });

        this.setupEventHandlers();
    }

    private calculateReconnectDelay(): number {
        // Exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s, 60s max
        const delay = Math.min(
            this.BASE_RECONNECT_DELAY * Math.pow(2, this.reconnectAttempts),
            60000
        );
        this.reconnectAttempts++;
        return delay;
    }

    private setupEventHandlers() {
        this.socket?.on('connect', () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0; // Reset on successful connection
        });

        this.socket?.on('connect_error', (error) => {
            console.error('Connection error:', error);
            this.showReconnectingIndicator();
        });

        this.socket?.on('disconnect', (reason) => {
            console.log('Disconnected:', reason);
            if (reason === 'io server disconnect') {
                // Server-initiated disconnect - manual reconnect required
                this.socket?.connect();
            }
        });
    }
}
```

**Network Resilience Features**:
- Ping/pong heartbeat mechanism (30-second intervals)
- Connection quality scoring based on latency and packet loss
- Adaptive data compression based on network conditions
- Graceful degradation (reduce animations, disable effects on poor connections)

---

### 2.1.4 Audio & Visual Effects (4 weeks, Medium Complexity)

**Description**: Polished audio and visual effects system for immersive gameplay experience.

**Key Features**:
- Card sounds (deal, flip, shuffle)
- Dealer voice announcements (multilingual)
- Chip animations and sound effects
- Celebration animations (winning hand visual effects)
- Ambient sounds (casino background, table ambience)
- Volume controls and mute options

**Performance Considerations**:
- Audio compression (MP3 for compatibility, AAC for iOS)
- Preload critical sounds during app startup
- Lazy load non-critical effects
- Use Web Audio API for low-latency playback

---

### 2.1.5 Cross-Platform Build System (3 weeks, Medium Complexity)

**Description**: Automated build pipeline for iOS and Android app store submissions.

**Key Features**:
- Cocos Creator 3.8+ build configuration
- iOS code signing and provisioning profiles
- Android keystore management
- App store screenshot generation
- Version management and release notes

**Build Configuration**:

| Platform | Build Tool | Output Size | Build Time |
|-----------|------------|-------------|------------|
| **iOS** | Xcode 15+ | ~25 MB | 3-5 minutes |
| **Android** | Android Studio / Gradle | ~20 MB | 2-4 minutes |

**App Store Requirements**:
- iOS: App Store Connect API, TestFlight for beta testing
- Android: Google Play Console, internal/alpha/beta tracks

---

## 2.2 Poker Game Engine (Server-Side)

The game engine handles all core poker logic, state management, and real-time game orchestration. Built in Go for optimal concurrency and performance.

### Module Overview Table

| Module | Effort | Complexity | Key Features | Dependencies |
|--------|--------|------------|--------------|-------------|
| **2.2.1 Hand Evaluation System** | 6 weeks | Very High | Ultra-fast hand ranking, equity calculation, Monte Carlo | Rust evaluator (FFI), Go bindings |
| **2.2.2 Table Management Engine** | 7 weeks | Very High | State machine, player actions, betting rounds, side pots | Redis, PostgreSQL, Kafka |
| **2.2.3 Game Rules & Validation** | 5 weeks | High | Rule enforcement, pot calculation, showdown resolution | Hand Evaluator, Table Engine |
| **2.2.4 RNG & Certification System** | 4 weeks | Medium-High | Hardware RNG, PRNG implementation, audit trails | Hardware RNG, AES-CTR, PostgreSQL |

### 2.2.1 Hand Evaluation System (6 weeks, Very High Complexity)

**Description**: High-performance poker hand evaluation system capable of processing millions of evaluations per second for real-time equity calculations and game logic.

**Research-Based Implementation Strategy**:

Hand evaluation is the most performance-critical component. Based on extensive research, we recommend integrating a Rust-based evaluator via Foreign Function Interface (FFI) for maximum throughput.

**Benchmarked Hand Evaluator Performance**:

| Evaluator | Language | Sequential | Random 5-Card | Random 7-Card | Memory Usage |
|------------|-----------|------------|---------------|---------------|--------------|
| **OMPEval** | C++ | 775M eval/sec | - | 272M eval/sec | 200KB lookup tables |
| **DoubleTap Evaluator** | C++ | - | 161M eval/sec | 133M eval/sec | Precomputed tables |
| **holdem-hand-evaluator** | Rust | **1.2B eval/sec** | - | - | ~212KB lookup tables |
| **PHEvaluator** | C++ | - | 50K eval/sec (Python) | 28K eval/sec (Python) | - |

**Recommendation**: Use **holdem-hand-evaluator (Rust)** for the following reasons:
- **1.2 Billion evaluations/second** on Ryzen 9 5950X (single-threaded)
- Small memory footprint (~212KB lookup tables)
- No external dependencies
- Safe memory management (Rust ownership model)
- Easy FFI integration with Go

**Implementation Architecture**:

```go
// Go-Rust FFI integration for hand evaluation
// hand_evaluator.go
/*
#cgo CFLAGS: -I./rust/target/include
#cgo LDFLAGS: -L./rust/target/release -lpoker_eval -lm
#include <stdlib.h>
#include <stdint.h>
#include "poker_eval.h"
*/
import "C"
import (
    "encoding/binary"
    "unsafe"
)

// Hand represents a 5-7 card poker hand
type Hand struct {
    Cards []byte // Card IDs: 0-51 (2c-As)
}

// Evaluate returns the hand rank (higher is better)
func (h *Hand) Evaluate() uint32 {
    if len(h.Cards) < 5 || len(h.Cards) > 7 {
        return 0
    }

    // Prepare input for Rust FFI
    cardCount := C.uint8_t(len(h.Cards))
    cardsPtr := (*C.uint8_t)(unsafe.Pointer(&h.Cards[0]))

    // Call Rust evaluator
    rank := C.evaluate_hand(cardsPtr, cardCount)

    return uint32(rank)
}

// CalculateEquity uses Monte Carlo simulation
func (h *Hand) CalculateEquity(opponents []Hand, iterations int) float32 {
    // Monte Carlo simulation via Rust FFI
    totalWins := C.uint32_t(0)

    for i := 0; i < iterations; i++ {
        // Simulate deck and deal remaining cards
        // Compare hands, tally wins
        // This is done in Rust for performance
        wins := C.simulate_hand(cardsPtr, opponentCardsPtr, numOpponents)
        totalWins += wins
    }

    return float32(totalWins) / float32(iterations)
}
```

**Rust Integration** (FFI layer):

```rust
// rust/src/lib.rs
use std::ffi::{c_uint8_t, c_uint32_t};

#[repr(C)]
pub struct CardArray {
    cards: *const c_uint8_t,
    len: usize,
}

#[no_mangle]
pub extern "C" fn evaluate_hand(cards: *const c_uint8_t, count: c_uint8_t) -> c_uint32_t {
    let card_slice = unsafe {
        std::slice::from_raw_parts(cards, count as usize)
    };

    // Use holdem-hand-evaluator crate
    let mut hand = holdem_hand_evaluator::Hand::new();
    for &card in card_slice {
        hand = hand.add_card(card as u8);
    }

    hand.evaluate() as c_uint32_t
}

#[no_mangle]
pub extern "C" fn simulate_hand(
    hero_cards: *const c_uint8_t,
    opponent_cards: *const c_uint8_t,
    num_opponents: c_uint8_t,
    iterations: c_uint32_t
) -> c_uint32_t {
    // Monte Carlo simulation logic
    // Returns number of wins out of iterations
    // ... implementation ...
    0
}
```

**Performance Targets**:
- Single hand evaluation: <1 microsecond
- Equity calculation (10K iterations): <10 milliseconds
- Support 100+ simultaneous evaluations per game server

---

### 2.2.2 Table Management Engine (7 weeks, Very High Complexity)

**Description**: Core game engine managing table state, player actions, betting rounds, and real-time game orchestration.

**Key Features**:
- State machine for Texas Hold'em phases (preflop, flop, turn, river, showdown)
- Player action validation (check, bet, call, raise, fold)
- Side pot calculation for all-in scenarios
- Multi-table tournament support (MTT)
- Sit-and-go (SNG) tournament logic

**Architecture Pattern (One Goroutine per Table)**:

```go
// game_table.go - Table goroutine implementation
type GameTable struct {
    id            string
    state         TableState
    players       map[string]*PlayerState
    actionChan    chan PlayerAction // Buffered channel for player actions
    phase         GamePhase        // Preflop, Flop, Turn, River, Showdown
    pot           int64
    sidePots      []SidePot
    dealerPos     int
    currentPos    int
    minBet        int64
    currentBet    int64
    ctx           context.Context
    cancel        context.CancelFunc
}

type PlayerAction struct {
    PlayerID  string
    ActionType string // fold, check, call, bet, raise
    Amount    int64  // Amount for bet/raise
    Timestamp time.Time
}

func NewGameTable(tableID string, config TableConfig) *GameTable {
    ctx, cancel := context.WithCancel(context.Background())

    return &GameTable{
        id:          tableID,
        state:       TableState{},
        players:     make(map[string]*PlayerState),
        actionChan:  make(chan PlayerAction, 100), // Buffered
        minBet:      config.SmallBlind * 2,
        ctx:         ctx,
        cancel:      cancel,
    }
}

func (t *GameTable) Run() {
    // Main game loop - one goroutine per table
    ticker := time.NewTicker(50 * time.Millisecond) // 20 FPS update rate
    defer ticker.Stop()

    for {
        select {
        case action := <-t.actionChan:
            // Process player action
            t.processAction(action)

        case <-ticker.C:
            // Periodic state updates (timers, auto-fold)
            t.updateState()

        case <-t.ctx.Done():
            // Table shutdown
            return
        }
    }
}

func (t *GameTable) processAction(action PlayerAction) {
    player := t.players[action.PlayerID]
    if player == nil {
        log.Printf("Unknown player: %s", action.PlayerID)
        return
    }

    // Validate action based on current game state
    if !t.validateAction(player, action) {
        log.Printf("Invalid action from %s: %s", action.PlayerID, action.ActionType)
        return
    }

    switch action.ActionType {
    case "fold":
        t.handleFold(player)
    case "check":
        t.handleCheck(player)
    case "call":
        t.handleCall(player)
    case "bet":
        t.handleBet(player, action.Amount)
    case "raise":
        t.handleRaise(player, action.Amount)
    }

    // Check if betting round complete
    if t.isBettingRoundComplete() {
        t.nextPhase()
    }

    // Broadcast updated state
    t.broadcastState()
}

func (t *GameTable) nextPhase() {
    switch t.phase {
    case Preflop:
        t.phase = Flop
        t.dealCommunityCards(3)
    case Flop:
        t.phase = Turn
        t.dealCommunityCards(1)
    case Turn:
        t.phase = River
        t.dealCommunityCards(1)
    case River:
        t.phase = Showdown
        t.resolveShowdown()
    case Showdown:
        t.startNewHand()
    }
}

func (t *GameTable) resolveShowdown() {
    // Evaluate all remaining players' hands
    var activePlayers []string
    for id, player := range t.players {
        if !player.Folded && player.Cards != nil {
            activePlayers = append(activePlayers, id)
        }
    }

    if len(activePlayers) == 1 {
        // Only one player left - they win
        winner := activePlayers[0]
        t.awardPot(winner, t.pot)
        return
    }

    // Multiple players - evaluate hands
    bestRank := uint32(0)
    var winners []string

    for _, playerID := range activePlayers {
        player := t.players[playerID]
        hand := Hand{Cards: player.Cards}
        rank := hand.Evaluate()

        if rank > bestRank {
            bestRank = rank
            winners = []string{playerID}
        } else if rank == bestRank {
            winners = append(winners, playerID)
        }
    }

    // Award pot (split if tie)
    potShare := t.pot / int64(len(winners))
    for _, winnerID := range winners {
        t.awardPot(winnerID, potShare)
    }
}
```

**State Machine Diagram**:

```
┌─────────────┐
│  Hand Start │
└──────┬──────┘
       │
       ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Preflop   │───▶│    Flop     │───▶│    Turn     │
│  Betting    │    │  (3 cards)  │    │  (1 card)   │
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                  │                  │
       │    ┌────────────┴────────────────┴────────────┐
       │    │                                         │
       ▼    ▼                                         ▼
┌─────────────┐                                ┌─────────────┐
│   Showdown  │◀──────────────────────────────────│   River     │
│  Evaluation │                                │  (1 card)   │
└──────┬──────┘                                └─────────────┘
       │
       ▼
┌─────────────┐
│   Pot Award │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Hand Reset │
└─────────────┘
```

**Side Pot Calculation Algorithm**:

```go
func (t *GameTable) calculateSidePots() []SidePot {
    // Identify all-in players and their bet amounts
    var allInPlayers []PlayerBet
    for _, player := range t.players {
        if player.AllInAmount > 0 {
            allInPlayers = append(allInPlayers, PlayerBet{
                PlayerID: player.ID,
                Bet:      player.CurrentRoundBet,
            })
        }
    }

    if len(allInPlayers) == 0 {
        // No side pots - single main pot
        return []SidePot{{Amount: t.pot, EligiblePlayers: t.getEligiblePlayers()}}
    }

    // Sort all-in bets ascending
    sort.Slice(allInPlayers, func(i, j int) bool {
        return allInPlayers[i].Bet < allInPlayers[j].Bet
    })

    var sidePots []SidePot
    currentLevel := int64(0)

    for _, allIn := range allInPlayers {
        // Calculate pot for this level
        betDiff := allIn.Bet - currentLevel
        potAmount := betDiff * int64(len(t.players))

        // Determine eligible players for this pot
        var eligible []string
        for _, player := range t.players {
            if player.CurrentRoundBet >= allIn.Bet {
                eligible = append(eligible, player.ID)
            }
        }

        sidePots = append(sidePots, SidePot{
            Amount:           potAmount,
            EligiblePlayers:  eligible,
            AssociatedPlayer: allIn.PlayerID,
        })

        currentLevel = allIn.Bet
    }

    // Remaining bets go to main pot
    remaining := t.pot - currentLevel * int64(len(t.players))
    if remaining > 0 {
        sidePots = append(sidePots, SidePot{
            Amount:          remaining,
            EligiblePlayers: t.getActivePlayers(),
        })
    }

    return sidePots
}
```

**Performance Targets**:
- Action processing latency: <5ms (P99)
- State broadcast: <10ms (P99)
- Support 5000+ concurrent tables per server
- Memory usage: ~2KB per active table

---

### 2.2.3 Game Rules & Validation (5 weeks, High Complexity)

**Description**: Comprehensive rule enforcement system covering all poker variants, betting limits, and edge case handling.

**Key Features**:
- Texas Hold'em rule set (No Limit, Pot Limit, Fixed Limit)
- Platform fee (formerly rake) calculation based on club configuration
- Timeout enforcement (auto-fold on inactivity)
- Entry/stack management
- Table configuration enforcement (blinds, ante, max players)

**Validation Rules**:

```go
// rule_validator.go
type RuleValidator struct {
    config TableConfig
    hand   *GameTable
}

func (v *RuleValidator) ValidateAction(player *PlayerState, action PlayerAction) bool {
    switch action.ActionType {
    case "check":
        return v.canCheck(player)
    case "call":
        return v.canCall(player, action.Amount)
    case "bet":
        return v.canBet(player, action.Amount)
    case "raise":
        return v.canRaise(player, action.Amount)
    case "fold":
        return true // Always allowed
    default:
        return false
    }
}

func (v *RuleValidator) canCheck(player *PlayerState) bool {
    // Can only check if no bet to call
    return v.hand.currentBet == 0
}

func (v *RuleValidator) canCall(player *PlayerState, amount int64) bool {
    // Amount must match current bet exactly
    if amount != v.hand.currentBet {
        return false
    }

    // Player must have enough chips
    return player.ChipCount >= amount
}

func (v *RuleValidator) canBet(player *PlayerState, amount int64) bool {
    // Must have no current bet (starting the betting)
    if v.hand.currentBet != 0 {
        return false
    }

    // Amount must be at least minimum bet
    if amount < v.hand.minBet {
        return false
    }

    // Cannot bet more than stack size
    return amount <= player.ChipCount
}

func (v *RuleValidator) canRaise(player *PlayerState, amount int64) bool {
    // Must have existing bet to raise
    if v.hand.currentBet == 0 {
        return false
    }

    // Raise must be at least minimum raise size
    minRaise := v.hand.currentBet * 2
    if v.config.LimitType == NoLimit {
        minRaise = v.hand.currentBet + v.hand.minBet
    }

    if amount < minRaise {
        return false
    }

    // Cannot raise more than stack size
    return amount <= player.ChipCount
}
```

**Rake Calculation (Multi-Level Configuration)**:

```go
// rake_calculator.go
func (t *GameTable) calculateRake(pot int64, numPlayers int) int64 {
    // Hierarchy: Table → Club → Agent → System Default
    config := t.getEffectiveRakeConfig()

    switch config.Type {
    case Percentage:
        return t.calculatePercentageRake(pot, config)
    case Fixed:
        return t.calculateFixedRake(config)
    case Hybrid:
        return t.calculateHybridRake(pot, config)
    default:
        return 0
    }
}

func (t *GameTable) calculatePercentageRake(pot int64, config RakeConfig) int64 {
    rake := int64(float64(pot) * config.Percentage)
    rake = min(rake, config.Cap) // Apply cap
    rake = min(rake, pot * config.MaxPotPercentage) // Never exceed max % of pot
    return rake
}

func (t *GameTable) getEffectiveRakeConfig() RakeConfig {
    // 1. Table-specific rule
    if t.tableConfig.Rake != nil {
        return *t.tableConfig.Rake
    }

    // 2. Club rule
    club := t.getClub()
    if club.RakeConfig != nil {
        return *club.RakeConfig
    }

    // 3. Agent rule
    agent := club.GetAgent()
    if agent.DefaultRake != nil {
        return *agent.DefaultRake
    }

    // 4. System default
    return defaultRakeConfig
}
```

---

### 2.2.4 RNG & Certification System (4 weeks, Medium-High Complexity)

**Description**: Cryptographically secure random number generation system designed for third-party RNG certification (eCOGRA, iTech Labs, GLI).

**Key Features**:
- Hardware RNG seed acquisition
- AES-CTR based cryptographic PRNG
- Deterministic shuffle algorithm
- Full audit trail logging
- Certification-ready implementation

**RNG Architecture**:

```go
// rng_system.go
type RNGSystem struct {
    hardwareRNG HardwareRNG
    prng        *ChaCha20PRNG // Or AES-CTR
    auditLog    AuditLogger
}

type HardwareRNG interface {
    GetRandomBytes(count int) ([]byte, error)
}

type ChaCha20PRNG struct {
    cipher cipher.AEAD
    nonce []byte
    counter uint64
}

func (r *RNGSystem) ShuffleDeck(deck []Card) ([]Card, error) {
    // 1. Obtain seed from hardware RNG
    seed, err := r.hardwareRNG.GetRandomBytes(32) // 256-bit seed
    if err != nil {
        return nil, err
    }

    // 2. Initialize PRNG with seed
    r.prng.Initialize(seed)

    // 3. Fisher-Yates shuffle (deterministic with PRNG)
    shuffled := make([]Card, len(deck))
    copy(shuffled, deck)

    for i := len(shuffled) - 1; i > 0; i-- {
        // Generate random index using PRNG
        randIndex := r.prng.RandomIndex(i + 1)

        // Swap
        shuffled[i], shuffled[randIndex] = shuffled[randIndex], shuffled[i]
    }

    // 4. Log shuffle for audit
    r.auditLog.LogShuffle(shuffled, seed)

    return shuffled, nil
}

// Fisher-Yates shuffle implementation
func FisherYatesShuffle(deck []Card, prng *ChaCha20PRNG) []Card {
    shuffled := make([]Card, len(deck))
    copy(shuffled, deck)

    for i := len(shuffled) - 1; i > 0; i-- {
        j := prng.RandomIndex(i + 1)
        shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
    }

    return shuffled
}
```

**Audit Trail Implementation**:

```go
// rng_audit.go
type RNGAuditEntry struct {
    Timestamp   time.Time `json:"timestamp"`
    TableID     string    `json:"tableId"`
    HandID      string    `json:"handId"`
    Seed        []byte    `json:"seed"`        // 256-bit seed
    DeckState   []Card    `json:"deckState"`   // Initial deck order
    ShuffledDeck []Card   `json:"shuffledDeck"` // After shuffle
    Algorithm   string    `json:"algorithm"`   // ChaCha20/AES-CTR
    Checksum    string    `json:"checksum"`    // SHA-256 of entry
}

type AuditLogger struct {
    db *sql.DB
}

func (a *AuditLogger) LogShuffle(deck []Card, seed []byte) error {
    entry := RNGAuditEntry{
        Timestamp:   time.Now().UTC(),
        TableID:     a.tableID,
        HandID:      a.handID,
        Seed:        seed,
        DeckState:   deck,
        Algorithm:   "ChaCha20-256",
    }

    entry.Checksum = a.calculateChecksum(&entry)

    _, err := a.db.Exec(`
        INSERT INTO rng_audit_log
        (timestamp, table_id, hand_id, seed, deck_state, shuffled_deck, algorithm, checksum)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, entry.Timestamp, entry.TableID, entry.HandID,
        entry.Seed, entry.DeckState, entry.ShuffledDeck,
        entry.Algorithm, entry.Checksum)

    return err
}

func (a *AuditLogger) calculateChecksum(entry *RNGAuditEntry) string {
    h := sha256.New()
    json.NewEncoder(h).Encode(entry)
    return hex.EncodeToString(h.Sum(nil))
}
```

**Certification Requirements**:

| Requirement | Standard | Implementation |
|-------------|-----------|----------------|
| **Seed Entropy** | ≥256 bits | Hardware RNG (TRNG) + ChaCha20 |
| **Shuffle Algorithm** | Fisher-Yates | Deterministic with PRNG |
| **Audit Trail** | Immutable logs | PostgreSQL append-only table |
| **Statistical Testing** | NIST SP 800-22 | Pre-certification testing suite |
| **Periodicity Testing** | Quarterly | External auditor access |

**Performance Targets**:
- Shuffle operation: <1ms
- Seed generation: <10ms (hardware RNG)
- Audit log write: <5ms

---

## 2.3 Agent & Club Management Panel

Web-based admin panel built with React/TypeScript, enabling agents to manage clubs, players, tables, and financial operations.

### Module Overview Table

| Module | Effort | Complexity | Key Features | Dependencies |
|--------|--------|------------|--------------|-------------|
| **2.3.1 Dashboard & Analytics** | 5 weeks | Medium-High | Real-time metrics, charts, reports | React Query, Recharts, WebSocket |
| **2.3.2 Player & Club Management** | 6 weeks | Medium-High | CRUD operations, permissions, settings | NestJS, PostgreSQL, TypeORM |
| **2.3.3 Financial Operations** | 5 weeks | High | Deposits/withdrawals, transaction history, balances | Payment gateway integration, audit trails |
| **2.3.4 Table Configuration** | 4 weeks | Medium | Game settings, rake rules, tournament setup | Game Engine API, PostgreSQL |

### 2.3.1 Dashboard & Analytics (5 weeks, Medium-High Complexity)

**Description**: Real-time dashboard displaying key metrics, player activity, revenue analytics, and operational insights.

**Key Features**:
- Real-time active tables and players count
- Revenue metrics (hourly, daily, weekly)
- Player acquisition and retention charts
- Game analytics (hands played, avg pot, rake collected)
- Performance monitoring (latency, error rates)
- Customizable date range filters

**Architecture**:

```typescript
// React Dashboard Component
import { useQuery } from '@tanstack/react-query';
import { LineChart, BarChart, PieChart } from 'recharts';
import { useWebSocket } from '@/hooks/useWebSocket';

function Dashboard() {
    const { data: metrics } = useQuery({
        queryKey: ['dashboard-metrics'],
        queryFn: fetchDashboardMetrics,
        refetchInterval: 30000, // Refresh every 30 seconds
    });

    const { data: realtimeData } = useWebSocket('wss://api.example.com/dashboard');

    return (
        <div className="dashboard">
            <h1>Agent Dashboard</h1>

            {/* Real-time Cards */}
            <div className="metrics-grid">
                <MetricCard
                    title="Active Tables"
                    value={realtimeData?.activeTables || 0}
                    icon="table"
                />
                <MetricCard
                    title="Online Players"
                    value={realtimeData?.onlinePlayers || 0}
                    icon="users"
                />
                <MetricCard
                    title="Today's Revenue"
                    value={`$${metrics?.todayRevenue || 0}`}
                    icon="dollar"
                />
                <MetricCard
                    title="Hands Played"
                    value={metrics?.handsPlayed || 0}
                    icon="cards"
                />
            </div>

            {/* Charts */}
            <div className="charts-grid">
                <div className="chart-card">
                    <h2>Revenue Trend (Last 7 Days)</h2>
                    <LineChart
                        width={600}
                        height={300}
                        data={metrics?.revenueTrend}
                        margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                    >
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" />
                        <YAxis />
                        <Tooltip />
                        <Legend />
                        <Line
                            type="monotone"
                            dataKey="revenue"
                            stroke="#8884d8"
                            name="Revenue ($)"
                        />
                    </LineChart>
                </div>

                <div className="chart-card">
                    <h2>Player Acquisition</h2>
                    <BarChart
                        width={600}
                        height={300}
                        data={metrics?.playerAcquisition}
                    >
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" />
                        <YAxis />
                        <Tooltip />
                        <Legend />
                        <Bar dataKey="newPlayers" fill="#82ca9d" name="New Players" />
                        <Bar dataKey="activePlayers" fill="#8884d8" name="Active Players" />
                    </BarChart>
                </div>
            </div>

            {/* Recent Activity */}
            <div className="activity-card">
                <h2>Recent Transactions</h2>
                <TransactionTable
                    transactions={metrics?.recentTransactions || []}
                />
            </div>
        </div>
    );
}
```

**WebSocket Hook for Real-Time Updates**:

```typescript
// hooks/useWebSocket.ts
import { useEffect, useState } from 'react';

export function useWebSocket(url: string) {
    const [data, setData] = useState<any>(null);
    const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('connecting');

    useEffect(() => {
        const ws = new WebSocket(url);

        ws.onopen = () => {
            setConnectionStatus('connected');
        };

        ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            setData(message);
        };

        ws.onclose = () => {
            setConnectionStatus('disconnected');
            // Attempt reconnect after 5 seconds
            setTimeout(() => {
                setConnectionStatus('connecting');
            }, 5000);
        };

        return () => {
            ws.close();
        };
    }, [url]);

    return { data, connectionStatus };
}
```

---

### 2.3.2 Player & Club Management (6 weeks, Medium-High Complexity)

**Description**: Comprehensive CRUD interface for managing players, clubs, and hierarchical permissions.

**Key Features**:
- Club creation and configuration
- Player registration and profile management
- Role-based access control (Agent, Manager, Moderator)
- Bulk player operations (import, export, suspend)
- Player statistics and game history
- Multi-club support for agents

**NestJS Backend API**:

```typescript
// clubs.controller.ts
@Controller('api/v1/clubs')
@UseGuards(JwtAuthGuard)
export class ClubsController {
    constructor(
        private readonly clubsService: ClubsService,
        private readonly playersService: PlayersService
    ) {}

    @Post()
    async createClub(
        @Body() createClubDto: CreateClubDto,
        @Req() req: Request
    ) {
        const agentId = req.user.agentId; // From JWT claim
        return this.clubsService.create(agentId, createClubDto);
    }

    @Get()
    async getClubs(@Req() req: Request) {
        const agentId = req.user.agentId;
        return this.clubsService.findByAgent(agentId);
    }

    @Get(':id')
    async getClub(@Param('id') clubId: string, @Req() req: Request) {
        const agentId = req.user.agentId;
        // RLS ensures agent can only access their own clubs
        return this.clubsService.findOne(clubId, agentId);
    }

    @Put(':id')
    async updateClub(
        @Param('id') clubId: string,
        @Body() updateClubDto: UpdateClubDto,
        @Req() req: Request
    ) {
        const agentId = req.user.agentId;
        return this.clubsService.update(clubId, agentId, updateClubDto);
    }

    @Delete(':id')
    async deleteClub(@Param('id') clubId: string, @Req() req: Request) {
        const agentId = req.user.agentId;
        return this.clubsService.delete(clubId, agentId);
    }
}

// players.controller.ts
@Controller('api/v1/players')
@UseGuards(JwtAuthGuard)
export class PlayersController {
    constructor(private readonly playersService: PlayersService) {}

    @Post()
    async createPlayer(
        @Body() createPlayerDto: CreatePlayerDto,
        @Req() req: Request
    ) {
        const agentId = req.user.agentId;
        return this.playersService.create(agentId, createPlayerDto);
    }

    @Get('club/:clubId')
    async getClubPlayers(
        @Param('clubId') clubId: string,
        @Query() pagination: PaginationDto,
        @Req() req: Request
    ) {
        const agentId = req.user.agentId;
        return this.playersService.findByClub(clubId, agentId, pagination);
    }

    @Get(':id')
    async getPlayer(@Param('id') playerId: string, @Req() req: Request) {
        const agentId = req.user.agentId;
        return this.playersService.findOne(playerId, agentId);
    }

    @Get(':id/stats')
    async getPlayerStats(@Param('id') playerId: string, @Req() req: Request) {
        const agentId = req.user.agentId;
        return this.playersService.getStats(playerId, agentId);
    }
}
```

**Data Models (TypeORM)**:

```typescript
// club.entity.ts
@Entity('clubs')
export class Club {
    @PrimaryGeneratedColumn('uuid')
    id: string;

    @Column()
    @Index()
    agentId: string; // Foreign key to agents table

    @Column()
    name: string;

    @Column({ type: 'jsonb', nullable: true })
    config: ClubConfig; // Rake rules, table settings, etc.

    @Column({ default: true })
    isActive: boolean;

    @CreateDateColumn()
    createdAt: Date;

    @UpdateDateColumn()
    updatedAt: Date;

    @OneToMany(() => Player, player => player.club)
    players: Player[];
}

// player.entity.ts
@Entity('players')
export class Player {
    @PrimaryGeneratedColumn('uuid')
    id: string;

    @Column()
    @Index()
    agentId: string;

    @Column()
    @Index()
    clubId: string;

    @Column({ unique: true })
    @Index()
    username: string;

    @Column({ select: false }) // Never expose in API responses
    passwordHash: string;

    @Column({ type: 'decimal', precision: 15, scale: 2, default: 0 })
    balance: decimal.DecimalType;

    @Column({ default: true })
    isActive: boolean;

    @Column({ default: false })
    isSuspended: boolean;

    @Column({ type: 'jsonb', nullable: true })
    profile: PlayerProfile;

    @CreateDateColumn()
    createdAt: Date;

    @UpdateDateColumn()
    updatedAt: Date;
}
```

**Row-Level Security (PostgreSQL)**:

```sql
-- Enable RLS on players table
ALTER TABLE players ENABLE ROW LEVEL SECURITY;

-- Policy: Agents can only access their own players
CREATE POLICY agent_isolation ON players
    FOR ALL
    USING (agent_id = current_setting('app.agent_id')::UUID);

-- Enable RLS on clubs table
ALTER TABLE clubs ENABLE ROW LEVEL SECURITY;

-- Policy: Agents can only access their own clubs
CREATE POLICY agent_isolation ON clubs
    FOR ALL
    USING (agent_id = current_setting('app.agent_id')::UUID);
```

---

### 2.3.3 Point Balance Operations (5 weeks, High Complexity)

**Description**: Point balance management system handling point allocations, redemptions, balance adjustments, and transaction auditing.

**Key Features**:
- Point allocation (external agent management)
- Point redemption requests (external agent management)
- Manual balance adjustments (agent-only)
- Transaction history with filters
- Automated platform fee collection
- Multi-point-type support (future)

**Point Balance Service (Phase 1-2)**:

```typescript
// point-balance.service.ts
import { Transaction, TransactionType } from '../entities/transaction.entity';

@Injectable()
export class PointBalanceService {
    async allocatePoints(
        playerId: string,
        points: number,
        agentId: string,
        reason: string
    ): Promise<Transaction> {
        // 1. Credit player balance with points
        await this.creditPlayerBalance(playerId, points);

        // 2. Record transaction
        const transaction = await this.createTransaction({
            playerId,
            agentId,
            type: TransactionType.ALLOCATION,
            amount: points,
            status: 'completed',
            reason
        });

        return transaction;
    }

    async requestPointRedemption(
        playerId: string,
        points: number,
        agentId: string
    ): Promise<Transaction> {
        // 1. Verify player has sufficient balance
        const player = await this.playersService.findOne(playerId, agentId);
        if (player.balance.lt(points)) {
            throw new Error('Insufficient balance');
        }

        // 2. Create pending transaction (external processing by agent)
        const transaction = await this.createTransaction({
            playerId,
            agentId,
            type: TransactionType.REDEMPTION,
            amount: points,
            status: 'pending'
        });

        // 3. Debit player balance (hold amount)
        await this.debitPlayerBalance(playerId, points);

        return transaction;
    }

    async approveRedemption(
        transactionId: string,
        agentId: string
    ): Promise<Transaction> {
        const transaction = await this.transactionRepository.findOne({
            where: { id: transactionId, agentId }
        });

        if (!transaction || transaction.status !== 'pending') {
            throw new Error('Invalid transaction');
        }

        // 1. Agent processes redemption externally
        // Platform only records approval

        // 2. Update transaction status
        transaction.status = 'completed';
        transaction.completedAt = new Date();
        await this.transactionRepository.save(transaction);

        return transaction;
    }

    private async creditPlayerBalance(playerId: string, amount: number) {
        await this.dataSource.transaction(async (manager) => {
            await manager.query(`
                UPDATE players
                SET balance = balance + $1
                WHERE id = $2
            `, [amount, playerId]);

            // Audit log
            await manager.insert(AuditLog, {
                action: 'balance_credit',
                entityType: 'player',
                entityId: playerId,
                details: { amount },
                timestamp: new Date()
            });
        });
    }

    private async debitPlayerBalance(playerId: string, amount: number) {
        await this.dataSource.transaction(async (manager) => {
            await manager.query(`
                UPDATE players
                SET balance = balance - $1
                WHERE id = $2 AND balance >= $1
            `, [amount, playerId]);

            // Audit log
            await manager.insert(AuditLog, {
                action: 'balance_debit',
                entityType: 'player',
                entityId: playerId,
                details: { amount },
                timestamp: new Date()
            });
        });
    }
}
```

**Transaction Entity**:

```typescript
// transaction.entity.ts
@Entity('transactions')
export class Transaction {
    @PrimaryGeneratedColumn('uuid')
    id: string;

    @Column()
    @Index()
    agentId: string;

    @Column()
    @Index()
    playerId: string;

    @Column({ type: 'enum', enum: TransactionType })
    type: TransactionType;

    @Column({ type: 'decimal', precision: 15, scale: 2 })
    amount: decimal.DecimalType;

    @Column({ type: 'decimal', precision: 15, scale: 2 })
    balanceAfter: decimal.DecimalType;

    @Column({ type: 'enum', enum: ['pending', 'completed', 'failed', 'cancelled'] })
    status: string;

    @Column({ nullable: true })
    gateway: string; // agent_external (Phase 3+ real-money: stripe, bank_transfer, paypal)

    @Column({ nullable: true })
    gatewayTransactionId: string; // External reference ID for agent-managed transactions (Phase 3+ only)

    @Column({ type: 'jsonb', nullable: true })
    metadata: Record<string, any>;

    @CreateDateColumn()
    createdAt: Date;

    @Column({ nullable: true })
    completedAt: Date;
}

export enum TransactionType {
    ALLOCATION = 'allocation', // Points allocated by agent
    REDEMPTION = 'redemption', // Points redeemed (external processing by agent)
    ADJUSTMENT = 'adjustment', // Manual balance adjustments
    PLATFORM_FEE = 'platform_fee', // Platform service charge (formerly rake)
    BONUS = 'bonus'
}
```

---

### 2.3.4 Table Configuration (4 weeks, Medium Complexity)

**Description**: Interface for configuring game tables, tournament structures, and game rules.

**Key Features**:
- Table creation wizard
- Blind structure configuration
- Platform fee rules setup (formerly rake)
- Tournament settings (entry fee, prize pool, structure)
- Seat limits and table type (point game, SNG, MTT)

**Table Configuration Form**:

```typescript
// components/TableConfigForm.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';

const tableConfigSchema = z.object({
    name: z.string().min(1).max(50),
    type: z.enum(['point_game', 'sitngo', 'tournament']),
    maxPlayers: z.number().min(2).max(9),
    smallBlind: z.number().positive(),
    bigBlind: z.number().positive(),
    ante: z.number().min(0),
    entryMin: z.number().positive(),
    entryMax: z.number().positive(),
    platformFeeConfig: z.object({
        type: z.enum(['percentage', 'fixed', 'hybrid']),
        percentage: z.number().min(0).max(1),
        cap: z.number().min(0),
        maxPotPercentage: z.number().min(0).max(1)
    })
});

type TableConfigFormData = z.infer<typeof tableConfigSchema>;

function TableConfigForm() {
    const { register, handleSubmit, formState: { errors } } = useForm<TableConfigFormData>({
        resolver: zodResolver(tableConfigSchema)
    });

    const onSubmit = async (data: TableConfigFormData) => {
        await createTable(data);
    };

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            {/* Basic Settings */}
            <div>
                <label>Table Name</label>
                <input {...register('name')} />
                {errors.name && <span>{errors.name.message}</span>}
            </div>

            <div>
                <label>Table Type</label>
                <select {...register('type')}>
                    <option value="point_game">Point Game</option>
                    <option value="sitngo">Sit & Go</option>
                    <option value="tournament">Tournament</option>
                </select>
            </div>

            {/* Player Limits */}
            <div>
                <label>Max Players</label>
                <input type="number" {...register('maxPlayers', { valueAsNumber: true })} />
            </div>

            {/* Blinds */}
            <div>
                <label>Small Blind</label>
                <input type="number" {...register('smallBlind', { valueAsNumber: true })} />
            </div>

            <div>
                <label>Big Blind</label>
                <input type="number" {...register('bigBlind', { valueAsNumber: true })} />
            </div>

            {/* Entry Range */}
            <div>
                <label>Min Entry</label>
                <input type="number" {...register('entryMin', { valueAsNumber: true })} />
            </div>

            <div>
                <label>Max Entry</label>
                <input type="number" {...register('entryMax', { valueAsNumber: true })} />
            </div>

            {/* Platform Fee Configuration */}
            <div>
                <label>Platform Fee Type</label>
                <select {...register('platformFeeConfig.type')}>
                    <option value="percentage">Percentage</option>
                    <option value="fixed">Fixed</option>
                    <option value="hybrid">Hybrid</option>
                </select>
            </div>

            {watch('platformFeeConfig.type') === 'percentage' && (
                <>
                    <div>
                        <label>Platform Fee Percentage</label>
                        <input
                            type="number"
                            step="0.01"
                            {...register('platformFeeConfig.percentage', { valueAsNumber: true })}
                        />
                    </div>
                    <div>
                        <label>Platform Fee Cap</label>
                        <input type="number" {...register('platformFeeConfig.cap', { valueAsNumber: true })} />
                    </div>
                </>
            )}

            <button type="submit">Create Table</button>
        </form>
    );
}
```

---

## 2.4 Super Admin Platform

Centralized admin panel for platform administrators to manage agents, monitor system health, and enforce compliance.

### Module Overview Table

| Module | Effort | Complexity | Key Features | Dependencies |
|--------|--------|------------|--------------|-------------|
| **2.4.1 Agent Management** | 4 weeks | Medium | Onboarding, tiering, configuration audit | Agent Panel API, PostgreSQL |
| **2.4.2 Platform Analytics** | 5 weeks | Medium-High | Aggregate metrics, revenue, growth tracking | PostgreSQL aggregation, Redis cache |
| **2.4.3 Compliance & Auditing** | 6 weeks | High | Regulatory compliance, audit logs, reporting | PostgreSQL, external compliance APIs |
| **2.4.4 System Monitoring** | 4 weeks | Medium-High | Infrastructure health, alerting, scaling | Prometheus, Grafana, Kubernetes API |

### 2.4.1 Agent Management (4 weeks, Medium Complexity)

**Description**: Comprehensive agent lifecycle management including onboarding, tiering, and configuration auditing.

**Key Features**:
- Agent registration and approval workflow
- Tier management (Bronze, Silver, Gold, Platinum)
- Revenue sharing configuration
- Whitelabel branding (logo, colors)
- Performance monitoring per agent
- Suspension and termination workflows

**Agent Entity**:

```typescript
// agent.entity.ts
@Entity('agents')
export class Agent {
    @PrimaryGeneratedColumn('uuid')
    id: string;

    @Column({ unique: true })
    @Index()
    username: string;

    @Column({ select: false })
    passwordHash: string;

    @Column({ type: 'enum', enum: AgentTier })
    tier: AgentTier;

    @Column({ type: 'jsonb' })
    branding: AgentBranding;

    @Column({ type: 'jsonb' })
    revenueShare: RevenueShareConfig;

    @Column({ default: true })
    isActive: boolean;

    @Column({ default: 0 })
    commissionRate: number; // Percentage of revenue

    @CreateDateColumn()
    createdAt: Date;

    @UpdateDateColumn()
    updatedAt: Date;

    @OneToMany(() => Club, club => club.agent)
    clubs: Club[];
}

export enum AgentTier {
    BRONZE = 'bronze',
    SILVER = 'silver',
    GOLD = 'gold',
    PLATINUM = 'platinum'
}

interface AgentBranding {
    logoUrl?: string;
    primaryColor?: string;
    secondaryColor?: string;
    customDomain?: string;
}

interface RevenueShareConfig {
    platformShare: number;  // Percentage for platform
    agentShare: number;     // Percentage for agent
}
```

---

### 2.4.2 Platform Analytics (5 weeks, Medium-High Complexity)

**Description**: Aggregated analytics platform providing insights into overall platform performance, revenue trends, and growth metrics.

**Key Features**:
- Platform-wide revenue dashboard
- Agent performance comparison
- Geographic distribution analysis
- Game type popularity metrics
- Player retention cohorts
- Custom report builder

**Analytics Queries (PostgreSQL)**:

```sql
-- Revenue per agent (last 30 days)
SELECT
    a.id,
    a.username,
    a.tier,
    COUNT(DISTINCT t.id) as total_tables,
    SUM(t.rake_collected) as total_rake,
    AVG(t.pot_size) as avg_pot_size
FROM agents a
JOIN clubs c ON c.agent_id = a.id
JOIN tables t ON t.club_id = c.id
WHERE t.created_at >= NOW() - INTERVAL '30 days'
    AND a.is_active = true
GROUP BY a.id, a.username, a.tier
ORDER BY total_rake DESC;

-- Player retention cohorts (weekly)
WITH player_cohorts AS (
    SELECT
        player_id,
        DATE_TRUNC('week', created_at) as cohort_week,
        MIN(created_at) as first_played
    FROM hands
    GROUP BY player_id, DATE_TRUNC('week', created_at)
),
weekly_retention AS (
    SELECT
        cohort_week,
        EXTRACT(WEEK FROM AGE(first_played, created_at)) as week_number,
        COUNT(DISTINCT player_id) as players
    FROM player_cohorts
    GROUP BY cohort_week, week_number
)
SELECT
    cohort_week,
    week_number,
    players,
    LAG(players, 1) OVER (PARTITION BY cohort_week ORDER BY week_number) as previous_week_players,
    CASE
        WHEN LAG(players, 1) OVER (PARTITION BY cohort_week ORDER BY week_number) > 0
        THEN (players::float / LAG(players, 1) OVER (PARTITION BY cohort_week ORDER BY week_number)) * 100
    END as retention_rate
FROM weekly_retention
WHERE week_number > 0
ORDER BY cohort_week, week_number;

-- Geographic distribution
SELECT
    country,
    COUNT(DISTINCT p.id) as total_players,
    COUNT(DISTINCT t.id) as total_tables,
    SUM(t.rake_collected) as total_rake
FROM players p
JOIN player_locations pl ON pl.player_id = p.id
JOIN clubs c ON c.agent_id = p.agent_id
JOIN tables t ON t.club_id = c.id AND t.created_at >= NOW() - INTERVAL '30 days'
GROUP BY country
ORDER BY total_rake DESC;
```

---

### 2.4.3 Compliance & Auditing (6 weeks, High Complexity)

**Description**: Comprehensive compliance system supporting regulatory requirements, audit logging, and risk reporting.

**Key Features**:
- Immutable audit logs (append-only tables)
- Player KYC verification workflows
- AML (Anti-Money Laundering) monitoring
- Suspicious activity reporting
- Regulatory report generation
- Data export for external audits

**Audit Log Architecture**:

```sql
-- Immutable audit log table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    agent_id UUID NOT NULL,
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB
) PARTITION BY RANGE (timestamp);

-- Create partitions (monthly)
CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Create index for efficient queries
CREATE INDEX idx_audit_logs_agent_id ON audit_logs(agent_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);

-- Trigger to populate audit logs automatically
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit_logs (agent_id, user_id, action, entity_type, entity_id, old_values)
        VALUES (
            NEW.agent_id,
            current_setting('app.user_id')::UUID,
            TG_OP,
            TG_TABLE_NAME,
            NEW.id,
            row_to_json(OLD)
        );
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO audit_logs (agent_id, user_id, action, entity_type, entity_id, old_values, new_values)
        VALUES (
            NEW.agent_id,
            current_setting('app.user_id')::UUID,
            TG_OP,
            TG_TABLE_NAME,
            NEW.id,
            row_to_json(OLD),
            row_to_json(NEW)
        );
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit_logs (agent_id, user_id, action, entity_type, entity_id, new_values)
        VALUES (
            NEW.agent_id,
            current_setting('app.user_id')::UUID,
            TG_OP,
            TG_TABLE_NAME,
            NEW.id,
            row_to_json(NEW)
        );
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to sensitive tables
CREATE TRIGGER audit_players
    AFTER INSERT OR UPDATE OR DELETE ON players
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_transactions
    AFTER INSERT OR UPDATE OR DELETE ON transactions
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();
```

**AML Monitoring Algorithm (Phase 3+ Real-Money Expansion Only)**:

> **Note**: AML (Anti-Money Laundering) monitoring applies to real-money transactions. For the point-based system (Phase 1-2), use the simplified SuspiciousActivityMonitor for unusual point balance patterns. The comprehensive AML system below is reserved for Phase 3+ if real-money deployment is required.

```go
// aml_monitor.go
type AMLMonitor struct {
    db *sql.DB
}

type SuspiciousActivity struct {
    PlayerID    string
    RiskScore   float64
    Reason      string
    Details     map[string]interface{}
    DetectedAt  time.Time
}

func (m *AMLMonitor) AnalyzePlayer(playerID string, timeWindow time.Duration) ([]SuspiciousActivity, error) {
    var activities []SuspiciousActivity

    // 1. Rapid point allocations and redemptions (layering pattern)
    layeringRisk, err := m.detectLayering(playerID, timeWindow)
    if err != nil {
        return nil, err
    }
    if layeringRisk.RiskScore > 0.7 {
        activities = append(activities, layeringRisk)
    }

    // 2. Multiple accounts from same IP/IP range
    multiAccountRisk, err := m.detectMultiAccount(playerID, timeWindow)
    if err != nil {
        return nil, err
    }
    if multiAccountRisk.RiskScore > 0.8 {
        activities = append(activities, multiAccountRisk)
    }

    // 3. Unusual transaction patterns
    patternRisk, err := m.detectUnusualPatterns(playerID, timeWindow)
    if err != nil {
        return nil, err
    }
    if patternRisk.RiskScore > 0.6 {
        activities = append(activities, patternRisk)
    }

    return activities, nil
}

func (m *AMLMonitor) detectLayering(playerID string, timeWindow time.Duration) (SuspiciousActivity, error) {
    query := `
        SELECT
            COUNT(*) as transaction_count,
            SUM(CASE WHEN type = 'allocation' THEN amount ELSE 0 END) as total_allocations,
            SUM(CASE WHEN type = 'redemption' THEN amount ELSE 0 END) as total_redemptions
        FROM transactions
        WHERE player_id = $1
            AND created_at >= NOW() - $2::INTERVAL
            AND status = 'completed'
    `

    var transactionCount int
    var totalAllocations, totalRedemptions float64

    err := m.db.QueryRow(query, playerID, timeWindow).Scan(
        &transactionCount,
        &totalAllocations,
        &totalRedemptions,
    )

    if err != nil {
        return SuspiciousActivity{}, err
    }

    // Risk calculation: high transaction count + high turnover rate
    turnoverRate := totalRedemptions / totalAllocations
    riskScore := float64(transactionCount) * 0.01 + turnoverRate * 0.5

    return SuspiciousActivity{
        PlayerID:   playerID,
        RiskScore:  min(riskScore, 1.0),
        Reason:     "Rapid point allocations and redemptions (layering pattern)",
        Details: map[string]interface{}{
            "transaction_count": transactionCount,
            "total_allocations": totalAllocations,
            "total_redemptions": totalRedemptions,
            "turnover_rate":    turnoverRate,
        },
        DetectedAt: time.Now(),
    }, nil
}
```

---

### 2.4.4 System Monitoring (4 weeks, Medium-High Complexity)

**Description**: Infrastructure monitoring and alerting system ensuring platform reliability and performance.

**Key Features**:
- Real-time service health dashboard
- Performance metrics (CPU, memory, latency)
- Alerting and notification system
- Log aggregation and search
- Capacity planning insights
- Automated scaling triggers

**Prometheus Metrics**:

```go
// metrics.go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Game server metrics
    activeTablesGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "game_server_active_tables",
        Help: "Number of active game tables",
    }, []string{"server_id"})

    activePlayersGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "game_server_active_players",
        Help: "Number of active players",
    }, []string{"server_id"})

    actionDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "game_action_duration_seconds",
        Help:    "Duration of game actions",
        Buckets: prometheus.DefBuckets,
    }, []string{"action_type"})

    websocketConnectionsGauge = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "websocket_active_connections",
        Help: "Number of active WebSocket connections",
    })

    // Database metrics
    dbQueryDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "database_query_duration_seconds",
        Help:    "Duration of database queries",
        Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    }, []string{"query_type", "table"})

    dbConnectionPoolGauge = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "database_connection_pool_size",
        Help: "Current database connection pool size",
    })

    // Cache metrics
    cacheHitRatioGauge = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "cache_hit_ratio",
        Help: "Cache hit ratio (0-1)",
    })
)

func RecordAction(actionType string, duration time.Duration) {
    actionDurationHistogram.WithLabelValues(actionType).Observe(duration.Seconds())
}

func UpdateTableCount(serverID string, count float64) {
    activeTablesGauge.WithLabelValues(serverID).Set(count)
}

func UpdatePlayerCount(serverID string, count float64) {
    activePlayersGauge.WithLabelValues(serverID).Set(count)
}
```

**Grafana Dashboard Queries**:

```promql
# Game Server Performance
# Average action latency by type
rate(game_action_duration_seconds_sum[5m]) / rate(game_action_duration_seconds_count[5m])

# Active tables across all servers
sum(game_server_active_tables)

# WebSocket connection trends
increase(websocket_active_connections[1h])

# Database query performance
histogram_quantile(0.99, rate(database_query_duration_seconds_bucket[5m]))

# Cache hit ratio
cache_hit_ratio

# CPU usage across game servers
avg by (instance) (rate(process_cpu_seconds_total[5m])) * 100

# Memory usage
avg by (instance) (process_resident_memory_bytes / 1024 / 1024)
```

---

## 2.5 Security & Anti-Cheat System

Multi-layered security system utilizing machine learning, behavioral analysis, and real-time monitoring to detect fraud, bots, and collusion.

### Module Overview Table

| Module | Effort | Complexity | Key Features | Dependencies |
|--------|--------|------------|--------------|-------------|
| **2.5.1 Bot Detection Engine** | 8 weeks | Very High | Behavioral patterns, ML classification, timing analysis | Go, ML models (Python), Kafka |
| **2.5.2 Collusion Detection** | 7 weeks | Very High | Hand correlation, network analysis, statistical anomalies | Go, Graph algorithms, PostgreSQL |
| **2.5.3 Device Fingerprinting** | 5 weeks | High | Multi-account prevention, proxy detection, device tracking | DeviceAtlas/FingerprintJS, Redis |
| **2.5.4 Real-Time Monitoring** | 4 weeks | Medium-High | Event streaming, risk scoring, automated flags | Kafka, Go consumers, Alerting |
| **2.5.5 Investigation Tools** | 5 weeks | Medium | Case management, evidence collection, reporting | PostgreSQL, Web UI (React) |

### 2.5.1 Bot Detection Engine (8 weeks, Very High Complexity)

**Description**: ML-powered bot detection system analyzing player behavior patterns, decision timing, and statistical anomalies.

**Key Features**:
- Behavioral pattern analysis (betting patterns, decision timing)
- Timing anomaly detection (reaction times variance)
- Statistical fingerprinting (win rate, VPIP, PFR metrics)
- ML model ensemble (Isolation Forest, Autoencoder, Neural Network)
- Real-time risk scoring
- Adaptive thresholds based on player count

**Research-Based Implementation Strategy**:

Based on research into poker bot detection, the following algorithms have proven effective:

| Detection Method | Algorithm | Accuracy | False Positive Rate | Complexity |
|------------------|------------|-----------|-------------------|-------------|
| **Behavioral Analysis** | Random Forest | 92-95% | 3-5% | Medium |
| **Timing Anomalies** | Isolation Forest | 88-92% | 5-8% | Low-Medium |
| **Pattern Recognition** | LSTM Neural Network | 94-97% | 2-4% | High |
| **Statistical Outliers** | Autoencoder | 90-93% | 4-6% | Medium-High |

**Recommendation**: Ensemble approach combining Isolation Forest (for outliers), LSTM (for patterns), and behavioral rules (for known bot signatures).

**Implementation Architecture**:

```go
// bot_detection.go
type BotDetectionEngine struct {
    timingAnalyzer    *TimingAnalyzer
    patternAnalyzer  *PatternAnalyzer
    statisticalAnalyzer *StatisticalAnalyzer
    mlModel          *MLModelEnsemble
    riskThreshold    float64
}

type PlayerBehavior struct {
    PlayerID         string
    ActionHistory    []PlayerAction
    TimingData       []TimingMetric
    Statistics       PlayerStatistics
    SessionHistory   []SessionData
}

type TimingMetric struct {
    ActionID      string
    ActionTime    time.Duration
    Timestamp     time.Time
}

type PlayerStatistics struct {
    TotalHands         int
    HandsWon          int
    WinRate           float64
    VPIP              float64 // Voluntarily Put $ In Pot
    PFR               float64 // Pre-Flop Raise
    AggressionFactor   float64
    ShowdownRate      float64
    AverageBetSize    float64
}

func (e *BotDetectionEngine) AnalyzePlayer(playerID string) (float64, []string, error) {
    // 1. Gather player behavior data
    behavior, err := e.gatherPlayerBehavior(playerID)
    if err != nil {
        return 0, nil, err
    }

    // 2. Run multiple detection algorithms in parallel
    var wg sync.WaitGroup
    var riskScores []float64
    var reasons []string
    var mu sync.Mutex

    algorithms := []struct {
        name string
        fn   func(*PlayerBehavior) (float64, string)
    }{
        {"Timing", e.timingAnalyzer.Analyze},
        {"Pattern", e.patternAnalyzer.Analyze},
        {"Statistical", e.statisticalAnalyzer.Analyze},
    }

    for _, algo := range algorithms {
        wg.Add(1)
        go func(name string, fn func(*PlayerBehavior) (float64, string)) {
            defer wg.Done()
            score, reason := fn(behavior)

            mu.Lock()
            riskScores = append(riskScores, score)
            if score > 0.5 {
                reasons = append(reasons, fmt.Sprintf("[%s] %s", name, reason))
            }
            mu.Unlock()
        }(algo.name, algo.fn)
    }

    wg.Wait()

    // 3. Combine scores using weighted ensemble
    combinedRisk := e.combineRiskScores(riskScores)

    // 4. If risk exceeds threshold, flag player
    if combinedRisk > e.riskThreshold {
        e.flagPlayer(playerID, combinedRisk, reasons)
    }

    return combinedRisk, reasons, nil
}

func (e *BotDetectionEngine) combineRiskScores(scores []float64) float64 {
    // Weighted ensemble: Timing (30%), Pattern (40%), Statistical (30%)
    weights := []float64{0.3, 0.4, 0.3}

    if len(scores) != len(weights) {
        return 0
    }

    total := 0.0
    for i, score := range scores {
        total += score * weights[i]
    }

    return total
}
```

**Timing Anomaly Detection (Isolation Forest)**:

```python
# timing_analyzer.py (Python ML model)
import numpy as np
from sklearn.ensemble import IsolationForest
from scipy import stats

class TimingAnalyzer:
    def __init__(self):
        self.model = IsolationForest(
            contamination=0.05,  # Expect 5% anomalies
            n_estimators=100,
            max_samples='auto',
            random_state=42
        )
        self.is_trained = False

    def train(self, data):
        """
        Train model on historical human player timing data.
        data: array of timing metrics (milliseconds)
        """
        # Features: mean, std, min, max, kurtosis, skewness
        features = self.extract_features(data)
        self.model.fit(features)
        self.is_trained = True

    def extract_features(self, timing_data):
        """
        Extract statistical features from timing sequences.
        """
        features = []
        for timings in timing_data:
            if len(timings) < 10:  # Need minimum samples
                continue

            feature_vector = [
                np.mean(timings),           # Mean reaction time
                np.std(timings),            # Standard deviation
                np.min(timings),             # Fastest reaction
                np.max(timings),             # Slowest reaction
                stats.kurtosis(timings),     # Kurtosis (peakedness)
                stats.skew(timings),         # Skewness (asymmetry)
                np.percentile(timings, 50),  # Median
                np.percentile(timings, 95),  # 95th percentile
            ]
            features.append(feature_vector)

        return np.array(features)

    def analyze(self, player_timings):
        """
        Analyze player timing for bot-like patterns.
        Returns risk score (0-1) and explanation.
        """
        if not self.is_trained:
            return 0.5, "Model not trained"

        features = self.extract_features([player_timings])
        anomaly_score = self.model.decision_function(features)[0]

        # Convert to 0-1 range (higher = more suspicious)
        risk_score = (1 - anomaly_score) / 2
        risk_score = max(0, min(1, risk_score))

        # Generate explanation
        explanation = self.generate_explanation(player_timings, risk_score)

        return risk_score, explanation

    def generate_explanation(self, timings, risk_score):
        mean_time = np.mean(timings)
        std_time = np.std(timings)

        if risk_score > 0.8:
            if std_time < 50:  # Very consistent timing
                return "Extremely consistent reaction times (<50ms variance)"
            elif mean_time < 500:  # Very fast reactions
                return "Unusually fast reaction times (avg <500ms)"
            else:
                return "Statistically unlikely timing pattern"

        elif risk_score > 0.5:
            return "Suspicious timing variability"

        return "Normal human-like timing patterns"
```

**Pattern Recognition (LSTM Neural Network)**:

```python
# pattern_analyzer.py
import numpy as np
import torch
import torch.nn as nn

class BotPatternLSTM(nn.Module):
    def __init__(self, input_size=10, hidden_size=64, num_layers=2, output_size=1):
        super(BotPatternLSTM, self).__init__()
        self.lstm = nn.LSTM(input_size, hidden_size, num_layers, batch_first=True)
        self.fc = nn.Linear(hidden_size, output_size)
        self.sigmoid = nn.Sigmoid()

    def forward(self, x):
        # LSTM layer
        out, _ = self.lstm(x)

        # Take the last time step's output
        out = out[:, -1, :]

        # Fully connected layer
        out = self.fc(out)
        out = self.sigmoid(out)

        return out

class PatternAnalyzer:
    def __init__(self):
        self.model = BotPatternLSTM()
        self.model.eval()
        self.sequence_length = 50  # Analyze last 50 actions

    def encode_actions(self, actions):
        """
        Encode player actions into feature vectors.
        Returns: shape (batch_size, sequence_length, feature_dim)
        """
        features = []
        for action in actions:
            # Feature vector: [action_type, position, pot_size, bet_amount, stack_size, phase]
            vector = [
                self.encode_action_type(action['type']),
                action['position'],
                action['pot_size'] / 1000,  # Normalize
                action['bet_amount'] / 1000,
                action['stack_size'] / 1000,
                self.encode_phase(action['phase']),
                action['is_all_in'],
                action['is_check'],
                action['is_call'],
                action['is_fold']
            ]
            features.append(vector)

        # Pad or truncate to sequence_length
        if len(features) < self.sequence_length:
            features.extend([[0] * 10] * (self.sequence_length - len(features)))
        else:
            features = features[:self.sequence_length]

        return np.array(features)[np.newaxis, :, :]  # Add batch dimension

    def encode_action_type(self, action_type):
        encoding = {'fold': 0, 'check': 1, 'call': 2, 'bet': 3, 'raise': 4}
        return encoding.get(action_type, 0)

    def encode_phase(self, phase):
        encoding = {'preflop': 0, 'flop': 1, 'turn': 2, 'river': 3}
        return encoding.get(phase, 0)

    def analyze(self, action_history):
        """
        Analyze action sequence for bot-like patterns.
        """
        if len(action_history) < 10:
            return 0.0, "Insufficient data"

        # Encode actions
        features = self.encode_actions(action_history)
        features_tensor = torch.FloatTensor(features)

        # Predict
        with torch.no_grad():
            risk_score = self.model(features_tensor).item()

        explanation = self.generate_explanation(action_history, risk_score)

        return risk_score, explanation

    def generate_explanation(self, actions, risk_score):
        # Analyze patterns
        fold_rate = sum(1 for a in actions if a['type'] == 'fold') / len(actions)
        raise_rate = sum(1 for a in actions if a['type'] == 'raise') / len(actions)

        if risk_score > 0.8:
            if fold_rate > 0.7:
                return "Excessive folding rate (>70%)"
            elif raise_rate > 0.6:
                return "Aggressive raising pattern (>60%)"
            else:
                return "Bot-like action sequence detected"

        elif risk_score > 0.5:
            return "Suspicious action pattern"

        return "Normal human-like action patterns"
```

**Performance Targets**:
- Analysis time per player: <500ms (ML inference)
- Real-time processing: Support 1000+ concurrent analyses
- False positive rate: <5%
- True positive rate: >90%

---

### 2.5.2 Collusion Detection (7 weeks, Very High Complexity)

**Description**: Advanced collusion detection analyzing hand histories, player networks, and statistical correlations between players.

**Key Features**:
- Hand history correlation analysis
- Player network graph construction
- Statistical collusion metrics
- Tournament collusion detection
- Chip dumping detection
- Soft-play identification

**Research-Based Implementation Strategy**:

Based on research into poker collusion detection, the following approaches have shown effectiveness:

| Method | Accuracy | Complexity | Detection Capability |
|---------|-----------|-------------|---------------------|
| **Hand Correlation** | 85-90% | Medium | Chip dumping, soft-play |
| **Network Analysis** | 80-88% | High | Organized rings, multi-accounting |
| **Statistical Outliers** | 75-85% | Low-Medium | Unusual win rates together |
| **Graph Clustering** | 88-93% | Very High | Large-scale collusion rings |

**Recommendation**: Graph-based approach combining hand correlation analysis with network clustering (Louvain algorithm) for identifying collusion rings.

**Implementation Architecture**:

```go
// collusion_detection.go
type CollusionDetector struct {
    handAnalyzer     *HandCorrelationAnalyzer
    networkAnalyzer  *PlayerNetworkAnalyzer
    statAnalyzer     *StatisticalAnalyzer
    graphBuilder     *PlayerGraphBuilder
    riskThreshold    float64
}

type PlayerPair struct {
    Player1   string
    Player2   string
    TogetherHands int
    TotalHands1  int
    TotalHands2  int
    Correlation  float64
    ChiSquared   float64
}

type CollusionRisk struct {
    Players    []string
    RiskScore  float64
    RiskLevel  string  // low, medium, high, critical
    Reasons    []string
    Evidence   CollusionEvidence
    DetectedAt time.Time
}

type CollusionEvidence struct {
    HandCorrelation     []HandPair
    NetworkMetrics      NetworkMetrics
    StatisticalAnomalies []StatisticalAnomaly
}

func (d *CollusionDetector) AnalyzePlayerPairs(playerIDs []string) ([]CollusionRisk, error) {
    var risks []CollusionRisk

    // 1. Analyze all player pairs
    for i := 0; i < len(playerIDs); i++ {
        for j := i + 1; j < len(playerIDs); j++ {
            pairRisk, err := d.analyzePair(playerIDs[i], playerIDs[j])
            if err != nil {
                continue
            }

            if pairRisk.RiskScore > d.riskThreshold {
                risks = append(risks, pairRisk)
            }
        }
    }

    // 2. Build player network graph
    graph, err := d.graphBuilder.BuildNetwork(playerIDs)
    if err != nil {
        return nil, err
    }

    // 3. Detect collusion rings using graph clustering
    rings := d.detectCollusionRings(graph)

    // 4. Analyze each ring
    for _, ring := range rings {
        ringRisk, err := d.analyzeRing(ring)
        if err != nil {
            continue
        }

        if ringRisk.RiskScore > d.riskThreshold {
            risks = append(risks, ringRisk)
        }
    }

    return risks, nil
}

func (d *CollusionDetector) analyzePair(player1, player2 string) (CollusionRisk, error) {
    // 1. Hand correlation analysis
    correlation, err := d.handAnalyzer.AnalyzeCorrelation(player1, player2)
    if err != nil {
        return CollusionRisk{}, err
    }

    // 2. Statistical analysis
    statRisk := d.statAnalyzer.AnalyzePair(player1, player2)

    // 3. Calculate combined risk
    combinedRisk := correlation.CorrelationScore * 0.6 + statRisk * 0.4

    // 4. Generate explanation
    var reasons []string
    if correlation.FoldTogetherRate > 0.7 {
        reasons = append(reasons, fmt.Sprintf("High fold-together rate (%.1f%%)", correlation.FoldTogetherRate*100))
    }
    if correlation.RarelyFoldToEachOther < 0.1 {
        reasons = append(reasons, "Rarely fold to each other (soft-play indicator)")
    }
    if correlation.WinTogetherRate > 0.6 {
        reasons = append(reasons, fmt.Sprintf("High win-together rate (%.1f%%)", correlation.WinTogetherRate*100))
    }

    riskLevel := d.calculateRiskLevel(combinedRisk)

    return CollusionRisk{
        Players:   []string{player1, player2},
        RiskScore: combinedRisk,
        RiskLevel: riskLevel,
        Reasons:   reasons,
        Evidence: CollusionEvidence{
            HandCorrelation: []HandPair{correlation},
        },
        DetectedAt: time.Now(),
    }, nil
}
```

**Hand Correlation Analysis**:

```go
// hand_correlation_analyzer.go
type HandCorrelationAnalyzer struct {
    db *sql.DB
}

type HandCorrelation struct {
    Player1          string
    Player2          string
    TogetherHands    int
    TotalHands       int
    FoldTogetherRate float64
    NeverFoldToRate float64
    WinTogetherRate  float64
    ChiSquared      float64
    CorrelationScore float64
}

func (a *HandCorrelationAnalyzer) AnalyzeCorrelation(player1, player2 string) (HandCorrelation, error) {
    // 1. Get hands where both players participated
    togetherHandsQuery := `
        SELECT
            COUNT(*) as together_hands,
            SUM(CASE WHEN h1.folded = true AND h2.folded = true THEN 1 ELSE 0 END) as fold_together,
            SUM(CASE WHEN h1.won = true AND h2.won = true THEN 1 ELSE 0 END) as win_together,
            SUM(CASE WHEN h1.action = 'fold' AND h2.action != 'fold' THEN 1 ELSE 0 END) as p1_fold_p2_not,
            SUM(CASE WHEN h2.action = 'fold' AND h1.action != 'fold' THEN 1 ELSE 0 END) as p2_fold_p1_not
        FROM (
            SELECT
                h.id,
                MAX(CASE WHEN ha.player_id = $1 THEN ha.won ELSE NULL END) as won,
                MAX(CASE WHEN ha.player_id = $1 THEN ha.folded ELSE NULL END) as folded,
                MAX(CASE WHEN ha.player_id = $1 THEN ha.last_action ELSE NULL END) as action
            FROM hands h
            JOIN hand_actions ha ON ha.hand_id = h.id
            WHERE ha.player_id IN ($1, $2)
            GROUP BY h.id
            HAVING COUNT(DISTINCT ha.player_id) = 2
        ) h1, (
            SELECT
                ha.player_id,
                ha.last_action,
                ha.won,
                ha.folded
            FROM hand_actions ha
            WHERE ha.hand_id IN (
                SELECT h.id
                FROM hands h
                JOIN hand_actions ha ON ha.hand_id = h.id
                WHERE ha.player_id IN ($1, $2)
                GROUP BY h.id
                HAVING COUNT(DISTINCT ha.player_id) = 2
            )
        ) h2
        WHERE h1.id = h2.hand_id
    `

    var togetherHands, foldTogether, winTogether, p1FoldP2Not, p2FoldP1Not int
    err := a.db.QueryRow(
        togetherHandsQuery,
        player1, player2,
    ).Scan(&togetherHands, &foldTogether, &winTogether, &p1FoldP2Not, &p2FoldP1Not)

    if err != nil {
        return HandCorrelation{}, err
    }

    if togetherHands < 10 {  // Need minimum samples
        return HandCorrelation{}, fmt.Errorf("insufficient data")
    }

    // 2. Calculate correlation metrics
    foldTogetherRate := float64(foldTogether) / float64(togetherHands)
    winTogetherRate := float64(winTogether) / float64(togetherHands)

    // Soft-play metric: rarely fold to each other
    p1NeverFoldToP2 := 1.0 - (float64(p1FoldP2Not) / float64(togetherHands))
    p2NeverFoldToP1 := 1.0 - (float64(p2FoldP1Not) / float64(togetherHands))
    neverFoldToRate := (p1NeverFoldToP2 + p2NeverFoldToP1) / 2.0

    // 3. Chi-squared test for independence
    chiSquared := a.calculateChiSquared(foldTogether, winTogether, togetherHands)

    // 4. Calculate correlation score
    correlationScore := a.calculateCorrelationScore(
        foldTogetherRate,
        neverFoldToRate,
        winTogetherRate,
        chiSquared,
        togetherHands,
    )

    return HandCorrelation{
        Player1:          player1,
        Player2:          player2,
        TogetherHands:     togetherHands,
        TotalHands:       togetherHands,
        FoldTogetherRate:  foldTogetherRate,
        NeverFoldToRate:  neverFoldToRate,
        WinTogetherRate:   winTogetherRate,
        ChiSquared:       chiSquared,
        CorrelationScore: correlationScore,
    }, nil
}

func (a *HandCorrelationAnalyzer) calculateCorrelationScore(
    foldTogetherRate,
    neverFoldToRate,
    winTogetherRate,
    chiSquared float64,
    sampleSize int,
) float64 {
    var score float64

    // High fold-together rate (chip dumping indicator)
    if foldTogetherRate > 0.6 {
        score += 0.3 * (foldTogetherRate - 0.6) * 2.5
    }

    // Soft-play indicator (rarely fold to each other)
    if neverFoldToRate > 0.8 {
        score += 0.25 * (neverFoldToRate - 0.8) * 5
    }

    // High win-together rate
    if winTogetherRate > 0.5 {
        score += 0.2 * (winTogetherRate - 0.5) * 2
    }

    // Chi-squared significance
    chiSignificance := 1 - chiSquaredToPValue(chiSquared, 1)  // 1 degree of freedom
    if chiSignificance > 0.95 {
        score += 0.25
    }

    // Apply confidence based on sample size
    confidence := min(1.0, float64(sampleSize)/1000)  // Max confidence at 1000 hands

    return min(1.0, score * confidence)
}
```

**Player Network Graph Construction**:

```go
// player_network.go
type PlayerGraph struct {
    nodes map[string]*PlayerNode
    edges map[string]map[string]*Edge
}

type PlayerNode struct {
    ID        string
    Degree    int
    Weight    float64  // Risk weight
    ClusterID int
}

type Edge struct {
    Weight     float64  // Correlation score
    HandCount  int
    Timestamps []time.Time
}

type PlayerGraphBuilder struct {
    correlationAnalyzer *HandCorrelationAnalyzer
    db *sql.DB
}

func (b *PlayerGraphBuilder) BuildNetwork(playerIDs []string) (*PlayerGraph, error) {
    graph := &PlayerGraph{
        nodes: make(map[string]*PlayerNode),
        edges: make(map[string]map[string]*Edge),
    }

    // 1. Add all nodes
    for _, playerID := range playerIDs {
        graph.nodes[playerID] = &PlayerNode{
            ID:     playerID,
            Degree: 0,
            Weight: 0,
        }
    }

    // 2. Analyze all pairs and add edges
    for i := 0; i < len(playerIDs); i++ {
        for j := i + 1; j < len(playerIDs); j++ {
            correlation, err := b.correlationAnalyzer.AnalyzeCorrelation(
                playerIDs[i],
                playerIDs[j],
            )
            if err != nil {
                continue
            }

            // Only add edge if correlation score exceeds threshold
            if correlation.CorrelationScore > 0.3 {
                b.addEdge(graph, playerIDs[i], playerIDs[j], correlation)
            }
        }
    }

    return graph, nil
}

func (b *PlayerGraphBuilder) addEdge(
    graph *PlayerGraph,
    player1, player2 string,
    correlation HandCorrelation,
) {
    if graph.edges[player1] == nil {
        graph.edges[player1] = make(map[string]*Edge)
    }
    if graph.edges[player2] == nil {
        graph.edges[player2] = make(map[string]*Edge)
    }

    edge := &Edge{
        Weight:     correlation.CorrelationScore,
        HandCount:  correlation.TogetherHands,
    }

    graph.edges[player1][player2] = edge
    graph.edges[player2][player1] = edge

    // Update node degrees
    graph.nodes[player1].Degree++
    graph.nodes[player2].Degree++
}
```

**Graph Clustering for Collusion Rings**:

```go
// clustering.go
func detectCollusionRings(graph *PlayerGraph) [][]string {
    // Use Louvain algorithm for community detection
    clusters := louvainClustering(graph)

    var rings [][]string

    // Convert to player lists
    for clusterID, nodes := range clusters {
        if len(nodes) < 2 {
            continue  // Need at least 2 players for collusion
        }

        var players []string
        for playerID := range nodes {
            players = append(players, playerID)
        }

        // Calculate cluster risk
        clusterRisk := calculateClusterRisk(graph, players, clusterID)

        // Update node weights
        for _, playerID := range players {
            graph.nodes[playerID].Weight = clusterRisk
            graph.nodes[playerID].ClusterID = clusterID
        }

        rings = append(rings, players)
    }

    // Sort by cluster size (largest first)
    sort.Slice(rings, func(i, j int) bool {
        return len(rings[i]) > len(rings[j])
    })

    return rings
}

func louvainClustering(graph *PlayerGraph) map[int]map[string]bool {
    // Initialize: each player in own cluster
    clusters := make(map[int]map[string]bool)
    clusterID := 0
    for playerID := range graph.nodes {
        clusters[clusterID] = map[string]bool{playerID: true}
        graph.nodes[playerID].ClusterID = clusterID
        clusterID++
    }

    // Iterative clustering
    changed := true
    iterations := 0
    maxIterations := 100

    for changed && iterations < maxIterations {
        changed = false
        iterations++

        for playerID, node := range graph.nodes {
            // Find best cluster to move to
            bestCluster := node.ClusterID
            bestModularity := calculateModularity(graph, clusters, node.ClusterID, playerID)

            for neighbor := range graph.edges[playerID] {
                neighborCluster := graph.nodes[neighbor].ClusterID
                modularity := calculateModularity(graph, clusters, neighborCluster, playerID)

                if modularity > bestModularity {
                    bestCluster = neighborCluster
                    bestModularity = modularity
                }
            }

            // Move to best cluster
            if bestCluster != node.ClusterID {
                // Remove from old cluster
                clusters[node.ClusterID][playerID] = false
                if len(clusters[node.ClusterID]) == 0 {
                    delete(clusters, node.ClusterID)
                }

                // Add to new cluster
                if clusters[bestCluster] == nil {
                    clusters[bestCluster] = make(map[string]bool)
                }
                clusters[bestCluster][playerID] = true

                node.ClusterID = bestCluster
                changed = true
            }
        }
    }

    return clusters
}

func calculateModularity(
    graph *PlayerGraph,
    clusters map[int]map[string]bool,
    clusterID int,
    playerID string,
) float64 {
    // Simplified modularity calculation
    // In production, use full modularity formula

    var internalWeight float64
    var totalWeight float64

    for neighbor, edge := range graph.edges[playerID] {
        totalWeight += edge.Weight

        if graph.nodes[neighbor].ClusterID == clusterID {
            internalWeight += edge.Weight
        }
    }

    // Modularity = (internal_weight / total_weight) - (degree / (2 * m))^2
    degree := float64(graph.nodes[playerID].Degree)
    m := float64(len(graph.edges)) / 2  // Total edge weight

    modularity := (internalWeight / totalWeight) - math.Pow(degree/(2*m), 2)

    return modularity
}
```

---

### 2.5.3 Device Fingerprinting (5 weeks, High Complexity)

**Description**: Multi-account prevention system tracking device fingerprints, IP addresses, and network characteristics.

**Key Features**:
- Browser/device fingerprint collection
- IP address and subnet tracking
- Device-IP association analysis
- Proxy and VPN detection
- Fingerprint hashing (privacy-preserving)
- Multi-account flagging

**Implementation Architecture**:

```go
// device_fingerprint.go
type DeviceFingerprint struct {
    ID              string
    DeviceID         string  // Hashed device fingerprint
    IPAddress       string
    UserAgent       string
    ScreenResolution string
    TimeZone        string
    Language        string
    CanvasFingerprint string
    WebGLFingerprint  string
    AudioFingerprint string
    Fonts           string
    Plugins         string
}

type FingerprintService struct {
    db      *sql.DB
    redis   *redis.Client
}

func (s *FingerprintService) RecordFingerprint(
    playerID string,
    fingerprint DeviceFingerprint,
) error {
    // 1. Hash device fingerprint (privacy-preserving)
    deviceHash := s.hashFingerprint(fingerprint)
    fingerprint.DeviceID = deviceHash

    // 2. Check for existing devices with same fingerprint
    existingPlayers, err := s.findPlayersByDevice(deviceHash)
    if err != nil {
        return err
    }

    // 3. Check for IP matches
    ipMatches, err := s.findPlayersByIP(fingerprint.IPAddress)
    if err != nil {
        return err
    }

    // 4. Store fingerprint
    _, err = s.db.Exec(`
        INSERT INTO device_fingerprints
        (player_id, device_id, ip_address, user_agent, created_at)
        VALUES ($1, $2, $3, $4, NOW())
    `, playerID, deviceHash, fingerprint.IPAddress, fingerprint.UserAgent)

    if err != nil {
        return err
    }

    // 5. Flag potential multi-accounting
    if len(existingPlayers) > 0 || len(ipMatches) > 0 {
        s.flagMultiAccount(playerID, existingPlayers, ipMatches)
    }

    return nil
}

func (s *FingerprintService) hashFingerprint(fp DeviceFingerprint) string {
    // Create hash from fingerprint components
    data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
        fp.ScreenResolution,
        fp.TimeZone,
        fp.Language,
        fp.CanvasFingerprint,
        fp.WebGLFingerprint,
        fp.AudioFingerprint,
        fp.Fonts,
    )

    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])[:16]  // First 16 chars
}

func (s *FingerprintService) findPlayersByDevice(deviceID string) ([]string, error) {
    query := `
        SELECT DISTINCT player_id
        FROM device_fingerprints
        WHERE device_id = $1
            AND created_at >= NOW() - INTERVAL '30 days'
    `

    rows, err := s.db.Query(query, deviceID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var players []string
    for rows.Next() {
        var playerID string
        if err := rows.Scan(&playerID); err != nil {
            continue
        }
        players = append(players, playerID)
    }

    return players, nil
}

func (s *FingerprintService) flagMultiAccount(
    playerID string,
    deviceMatches []string,
    ipMatches []string,
) {
    allMatches := append(deviceMatches, ipMatches...)
    uniqueMatches := unique(allMatches)

    if len(uniqueMatches) == 0 {
        return
    }

    // Create security alert
    alert := SecurityAlert{
        Type:          "multi_account",
        Severity:       "high",
        PlayerID:       playerID,
        RelatedPlayers: uniqueMatches,
        Details: map[string]interface{}{
            "device_matches": len(deviceMatches),
            "ip_matches":    len(ipMatches),
            "total_matches":  len(uniqueMatches),
        },
        DetectedAt: time.Now(),
    }

    s.createAlert(alert)
}
```

**Client-Side Fingerprint Collection**:

```typescript
// utils/fingerprint.ts
export async function collectFingerprint(): Promise<DeviceFingerprint> {
  const fingerprint: DeviceFingerprint = {
    screenResolution: `${screen.width}x${screen.height}`,
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    language: navigator.language,
    canvasFingerprint: await getCanvasFingerprint(),
    webglFingerprint: getWebGLFingerprint(),
    audioFingerprint: await getAudioFingerprint(),
    fonts: await detectFonts(),
    plugins: navigator.plugins?.length || 0,
  };

  return fingerprint;
}

async function getCanvasFingerprint(): Promise<string> {
  const canvas = document.createElement('canvas');
  const ctx = canvas.getContext('2d');
  if (!ctx) return '';

  ctx.textBaseline = 'top';
  ctx.font = '14px Arial';
  ctx.fillStyle = '#f60';
  ctx.fillRect(125, 1, 62, 20);
  ctx.fillStyle = '#069';
  ctx.fillText('Hello, world! 👋', 2, 15);
  ctx.fillStyle = 'rgba(102, 204, 0, 0.7)';
  ctx.fillText('Hello, world! 👋', 4, 17);

  return canvas.toDataURL().substring(0, 100);  // First 100 chars
}

function getWebGLFingerprint(): string {
  const canvas = document.createElement('canvas');
  const gl = canvas.getContext('webgl');
  if (!gl) return '';

  const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
  const vendor = gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL);
  const renderer = gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL);

  return `${vendor}|${renderer}`;
}

async function getAudioFingerprint(): Promise<string> {
  try {
    const audioContext = new AudioContext();
    const oscillator = audioContext.createOscillator();
    const analyser = audioContext.createAnalyser();
    const gain = audioContext.createGain();
    const scriptProcessor = audioContext.createScriptProcessor(4096, 1, 1);

    oscillator.connect(analyser);
    analyser.connect(scriptProcessor);
    scriptProcessor.connect(gain);
    gain.connect(audioContext.destination);

    oscillator.start(0);

    const buffer = new Float32Array(4096);
    scriptProcessor.onaudioprocess = (e) => {
      e.inputBuffer.getChannelData(0).copyToChannel(buffer, 0);
    };

    oscillator.stop(0);
    audioContext.close();

    return Array.from(buffer).slice(0, 10).join(',');
  } catch {
    return '';
  }
}

async function detectFonts(): Promise<string> {
  const baseFonts = ['monospace', 'sans-serif', 'serif'];
  const testFonts = [
    'Arial', 'Courier New', 'Georgia', 'Times New Roman',
    'Verdana', 'Helvetica', 'Impact', 'Comic Sans MS'
  ];

  const canvas = document.createElement('canvas');
  const ctx = canvas.getContext('2d');
  if (!ctx) return '';

  const detectedFonts: string[] = [];

  testFonts.forEach(font => {
    ctx.font = `72px ${font}`;
    const testText = 'mmmmmmmmmmlli';
    ctx.fillText(testText, 0, 50);

    baseFonts.forEach(baseFont => {
      ctx.font = `72px ${baseFont}`;
      const baseline = ctx.measureText(testText).width;

      ctx.font = `72px ${font}, ${baseFont}`;
      const testWidth = ctx.measureText(testText).width;

      if (testWidth !== baseline) {
        detectedFonts.push(font);
      }
    });
  });

  return detectedFonts.join(',');
}
```

---

### 2.5.4 Real-Time Monitoring (4 weeks, Medium-High Complexity)

**Description**: Real-time event processing system analyzing game actions, player behavior, and security events as they occur.

**Key Features**:
- Kafka-based event streaming
- Real-time risk scoring
- Automated flagging and alerts
- Dashboard integration
- Historical trend analysis

**Kafka Consumer Architecture**:

```go
// real_time_monitor.go
type RealTimeMonitor struct {
    kafkaConsumer sarama.ConsumerGroupHandler
    botDetector    *BotDetectionEngine
    collusionDetector *CollusionDetector
    riskStore      *RiskScoreStore
    alertService   *AlertService
}

func (m *RealTimeMonitor) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    for message := range claim.Messages() {
        // Parse event
        event, err := m.parseEvent(message.Value)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            continue
        }

        // Route to appropriate detector
        switch event.Type {
        case "player_action":
            m.handlePlayerAction(session, message, event)
        case "hand_complete":
            m.handleHandComplete(session, message, event)
        case "table_create":
            m.handleTableCreate(session, message, event)
        }
    }

    return nil
}

func (m *RealTimeMonitor) handlePlayerAction(
    session sarama.ConsumerGroupSession,
    message *sarama.ConsumerMessage,
    event GameEvent,
) {
    playerID := event.PlayerID

    // 1. Analyze action for bot patterns
    riskScore, reasons, err := m.botDetector.AnalyzePlayer(playerID)
    if err != nil {
        log.Printf("Bot detection error for %s: %v", playerID, err)
        return
    }

    // 2. Update risk score in store
    m.riskStore.UpdateRiskScore(playerID, riskScore, reasons)

    // 3. Check threshold
    if riskScore > 0.8 {
        alert := SecurityAlert{
            Type:      "bot_detected",
            Severity:   "critical",
            PlayerID:   playerID,
            RiskScore:  riskScore,
            Reasons:    reasons,
            DetectedAt: time.Now(),
            EventID:    message.Key,
        }

        m.alertService.CreateAlert(alert)
    }

    // Mark message as processed
    session.MarkMessage(message, "")
}

func (m *RealTimeMonitor) handleHandComplete(
    session sarama.ConsumerGroupSession,
    message *sarama.ConsumerMessage,
    event GameEvent,
) {
    playerIDs := event.PlayerIDs

    // 1. Check for collusion among participating players
    risks, err := m.collusionDetector.AnalyzePlayerPairs(playerIDs)
    if err != nil {
        log.Printf("Collusion detection error: %v", err)
        return
    }

    // 2. Process collusion risks
    for _, risk := range risks {
        if risk.RiskScore > 0.7 {
            alert := SecurityAlert{
                Type:          "collusion_detected",
                Severity:       "high",
                Players:        risk.Players,
                RiskScore:      risk.RiskScore,
                Reasons:        risk.Reasons,
                Evidence:       risk.Evidence,
                DetectedAt:     risk.DetectedAt,
                EventID:        message.Key,
            }

            m.alertService.CreateAlert(alert)
        }
    }

    session.MarkMessage(message, "")
}
```

---

### 2.5.5 Investigation Tools (5 weeks, Medium Complexity)

**Description**: Web-based tools for security analysts to investigate flagged players, review evidence, and manage security cases.

**Key Features**:
- Player investigation dashboard
- Timeline visualization of events
- Hand history replay
- Evidence collection tools
- Case management system
- Report generation

**React Investigation Dashboard**:

```typescript
// pages/Investigation.tsx
import { useQuery } from '@tanstack/react-query';
import { Timeline, TimelineItem } from 'react-event-timeline';
import { PlayerTimeline } from '@/components/PlayerTimeline';
import { HandReplayer } from '@/components/HandReplayer';

function Investigation({ playerID }: { playerID: string }) {
    const { data: player } = useQuery({
        queryKey: ['player', playerID],
        queryFn: () => fetchPlayer(playerID),
    });

    const { data: alerts } = useQuery({
        queryKey: ['alerts', playerID],
        queryFn: () => fetchAlerts(playerID),
    });

    const { data: timeline } = useQuery({
        queryKey: ['timeline', playerID],
        queryFn: () => fetchPlayerTimeline(playerID),
    });

    const { data: statistics } = useQuery({
        queryKey: ['stats', playerID],
        queryFn: () => fetchPlayerStatistics(playerID),
    });

    return (
        <div className="investigation">
            <h1>Player Investigation: {player?.username}</h1>

            {/* Player Stats */}
            <div className="stats-grid">
                <StatCard title="Total Hands" value={statistics?.totalHands || 0} />
                <StatCard title="Win Rate" value={`${(statistics?.winRate || 0).toFixed(2)}%`} />
                <StatCard title="VPIP" value={`${(statistics?.vpip || 0).toFixed(2)}%`} />
                <StatCard title="PFR" value={`${(statistics?.pfr || 0).toFixed(2)}%`} />
                <StatCard title="Aggression Factor" value={(statistics?.aggression || 0).toFixed(2)} />
                <StatCard title="Risk Score" value={(player?.riskScore || 0).toFixed(2)} />
            </div>

            {/* Security Alerts */}
            <div className="alerts-section">
                <h2>Security Alerts</h2>
                {alerts?.map(alert => (
                    <AlertCard key={alert.id} alert={alert} />
                ))}
            </div>

            {/* Timeline */}
            <div className="timeline-section">
                <h2>Activity Timeline</h2>
                <PlayerTimeline events={timeline || []} />
            </div>

            {/* Hand History */}
            <div className="hand-history-section">
                <h2>Recent Hands</h2>
                <HandHistoryTable playerID={playerID} />
            </div>

            {/* Actions */}
            <div className="actions-section">
                <h2>Investigation Actions</h2>
                <button onClick={() => suspendPlayer(playerID)}>Suspend Player</button>
                <button onClick={() => requestReview(playerID)}>Request Manual Review</button>
                <button onClick={() => dismissAlerts(playerID)}>Dismiss Alerts</button>
            </div>
        </div>
    );
}

function AlertCard({ alert }: { alert: SecurityAlert }) {
    const severityColors = {
        low: 'green',
        medium: 'yellow',
        high: 'orange',
        critical: 'red',
    };

    return (
        <div className={`alert-card ${severityColors[alert.severity]}`}>
            <h3>{alert.type}</h3>
            <p>Risk Score: {alert.riskScore.toFixed(2)}</p>
            <p>Detected: {formatDate(alert.detectedAt)}</p>
            <ul>
                {alert.reasons?.map((reason, i) => (
                    <li key={i}>{reason}</li>
                ))}
            </ul>
        </div>
    );
}
```

---

## Summary

Section 2 provides a comprehensive breakdown of 22 core modules across 5 major components:

| Component | Modules | Total Effort | Avg Complexity |
|-----------|----------|---------------|----------------|
| **Player Mobile App** | 5 | 26 weeks | Medium-High |
| **Game Engine (Server)** | 4 | 22 weeks | Very High |
| **Agent & Club Panel** | 4 | 20 weeks | Medium-High |
| **Super Admin Platform** | 4 | 19 weeks | Medium-High |
| **Security & Anti-Cheat** | 5 | 29 weeks | Very High |
| **Total** | **22** | **116 weeks** | **High** |

### Key Technical Highlights

**Performance Benchmarks**:
- Hand evaluation: 1.2 Billion evaluations/sec (Rust-based)
- Game action latency: <100ms (P99)
- WebSocket connections: 15K+ per server
- Bot detection analysis: <500ms per player

**ML/AI Complexity**:
- Bot detection: Ensemble of Isolation Forest + LSTM + Behavioral rules
- Collusion detection: Graph-based clustering (Louvain algorithm)
- False positive rate target: <5%
- True positive rate target: >90%

**Security Architecture**:
- Multi-layered defense (client → API → auth → business logic → DB)
- Real-time Kafka event streaming for anti-cheat
- Immutable audit logs (append-only PostgreSQL partitions)
- Device fingerprinting for multi-account prevention

---

*Next Section: Section 3 - Milestone-Wise Delivery Plan*
