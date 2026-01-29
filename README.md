# B2B Poker Platform

A scalable, enterprise-grade poker platform designed for B2B deployment with agents and club management.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Layer                            │
│  ┌──────────────┐          ┌──────────────┐                 │
│  │ Mobile App   │          │ Web Admin    │                 │
│  │ (Cocos)      │          │ (React)      │                 │
│  └──────────────┘          └──────────────┘                 │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  API Gateway / Load Balancer                │
│              (Nginx + Rate Limiting + SSL)                 │
└─────────────────────────────────────────────────────────────┘
                            │
             ┌──────────────┼──────────────┐
             ▼              ▼              ▼
┌──────────────────┐ ┌──────────────┐ ┌──────────────────┐
│ Game Engine     │ │ Real-Time    │ │ User Auth       │
│ (Go)            │ │ Socket.IO    │ │ (Node.js)        │
│ - Table Logic   │ │ - Rooms      │ │ - JWT/OAuth      │
│ - Game State    │ │ - Events     │ │ - Sessions       │
└──────────────────┘ └──────────────┘ └──────────────────┘
```

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Game Engine** | Go 1.21+ | Real-time poker logic, 10K+ concurrent tables |
| **API Services** | Node.js + NestJS | REST APIs, authentication, admin operations |
| **Database** | PostgreSQL 15+ | ACID compliance, partitioning, RLS |
| **Cache** | Redis 7+ | Sessions, game state, pub/sub |
| **Event Streaming** | Apache Kafka | Anti-cheat, analytics, audit logs |
| **Mobile Client** | Cocos Creator 3.8+ | Cross-platform iOS/Android |
| **Admin Panel** | React + TypeScript | Agent/Super-admin dashboards |

## Project Structure

```
poker-platform/
├── cmd/                    # Entry points
│   └── game-server/        # Go game server
├── internal/               # Internal Go packages
│   └── game/               # Poker game engine
├── pkg/                    # Shared packages
│   ├── poker/              # Hand evaluation
│   └── rng/                # Cryptographic RNG
├── api/                    # Node.js API service
│   ├── src/
│   │   ├── modules/        # NestJS modules
│   │   │   ├── auth/       # Authentication
│   │   │   ├── clubs/      # Club management
│   │   │   ├── players/    # Player management
│   │   │   └── tables/     # Table management
│   │   └── main.ts
│   └── package.json
├── admin/                  # React admin panel
├── migrations/             # Database migrations
├── tests/                  # Shared tests
├── docker-compose.yml      # Local development
└── README.md
```

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.21+
- Node.js 20+
- PostgreSQL 15+ (if not using Docker)

### Quick Start (Development)

```bash
# Start all services
docker-compose up -d

# API server will be available at http://localhost:3001
# Game WebSocket at ws://localhost:3002/ws/{tableId}
# Admin panel at http://localhost:3000
```

### Manual Development

```bash
# Start PostgreSQL and Redis
docker-compose up postgres redis -d

# Set up API
cd api
npm install
npm run start:dev

# In another terminal, start game server
cd cmd/game-server
go run main.go
```

## Key Features

### Game Engine
- Texas Hold'em with No Limit, Pot Limit, Fixed Limit
- Tournament support (SNG, MTT)
- Side pot calculation
- Dealer button rotation
- One goroutine per table pattern

### Security
- Cryptographic RNG (AES-CTR-256)
- Hardware entropy seeding
- Full audit trail for certification (eCOGRA/iTech Labs)
- Row-level security in PostgreSQL
- JWT authentication

### Multi-Tenancy
- Agent-level data isolation
- Club management per agent
- White-label branding support
- Configurable rake per club

## API Documentation

### Authentication
```bash
POST /api/v1/auth/register
POST /api/v1/auth/login
GET /api/v1/auth/me
```

### Clubs
```bash
GET /api/v1/clubs
POST /api/v1/clubs
GET /api/v1/clubs/:id
PUT /api/v1/clubs/:id
DELETE /api/v1/clubs/:id
```

### Players
```bash
GET /api/v1/players
POST /api/v1/players
GET /api/v1/players/:id
PUT /api/v1/players/:id
```

### Tables
```bash
GET /api/v1/tables
POST /api/v1/tables
GET /api/v1/tables/:id
PUT /api/v1/tables/:id
```

## WebSocket Protocol

### Connect
```javascript
const ws = new WebSocket('ws://localhost:3002/ws/{tableId}');
ws.on('open', () => {
  ws.send(JSON.stringify({
    type: 'join',
    player_id: 'player-123',
    player_name: 'John',
    chips: 10000
  }));
});
```

### Actions
```javascript
ws.send(JSON.stringify({
  type: 'action',
  player_id: 'player-123',
  action: 'bet',
  amount: 100
}));
```

## Performance Targets

| Metric | Target |
|--------|--------|
| Game action latency (P99) | <100ms |
| Concurrent players per server | 10,000+ |
| Tables active per server | 5,000+ |
| Hand evaluations/second | 200M+ |

## License

Proprietary - All rights reserved
