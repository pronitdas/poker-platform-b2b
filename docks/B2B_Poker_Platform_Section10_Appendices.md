# Section 10: Appendices

## A. Glossary of Terms

### Poker Terminology

| Term | Definition | Context |
|------|------------|---------|
| **VPIP** | Voluntarily Put Money In Pot - Percentage of hands where a player voluntarily puts money into the pot before the flop | Player Statistics |
| **PFR** | Pre-Flop Raise - Percentage of hands where a player raises before the flop | Player Statistics |
| **AF** | Aggression Factor - Ratio of aggressive actions (bet + raise) to passive actions (call) | Player Statistics |
| **Blinds** | Forced bets that start the pot (Small Blind and Big Blind) | Game Rules |
| **Pot** | Total amount of chips/money currently in play for the hand | Game State |
| **Bet** | Placing chips into the pot as an action | Player Action |
| **Raise** | Increasing the previous bet amount | Player Action |
| **Fold** | Discarding hand and forfeiting the pot | Player Action |
| **Check** | Passing action without betting (when no bet has been made) | Player Action |
| **Call** | Matching the current bet amount | Player Action |
| **All-in** | Betting all remaining chips | Player Action |
| **Rake** | Commission taken by the house from each pot | Financial |
| **Ante** | Mandatory bet required from all players before cards are dealt | Game Rules |
| **Showdown** | Final phase where remaining players reveal their cards | Game Phase |
| **Flop** | First three community cards dealt | Game Phase |
| **Turn** | Fourth community card dealt | Game Phase |
| **River** | Fifth and final community card dealt | Game Phase |
| **Kicker** | Card used to break ties when players have same hand rank | Hand Evaluation |
| **Suited** | Cards of the same suit | Card Properties |
| **Off-suit** | Cards of different suits | Card Properties |
| **Pocket Pair** | Two cards of the same rank dealt to a player | Hand Type |

### Technical Terminology

| Term | Definition | Context |
|------|------------|---------|
| **WebSocket** | Full-duplex communication protocol over TCP for real-time bidirectional communication | Network Protocol |
| **Goroutine** | Lightweight thread in Go with 2KB memory footprint | Go Concurrency |
| **RLS (Row-Level Security)** | PostgreSQL feature restricting data access based on user roles | Database Security |
| **FFI (Foreign Function Interface)** | Mechanism for one programming language to call code from another language | Interoperability |
| **ML (Machine Learning)** | Algorithms and statistical models enabling systems to improve through experience | Anti-Cheat |
| **LSTM (Long Short-Term Memory)** | Type of recurrent neural network for sequence prediction tasks | Anti-Cheat/ML |
| **Microservices** | Architectural style structuring application as collection of loosely coupled services | Architecture |
| **Domain-Driven Design (DDD)** | Software development approach focusing on domain logic and modeling | Architecture |
| **Event Sourcing** | Pattern where state changes are stored as sequence of events | Data Architecture |
| **CQRS (Command Query Responsibility Segregation)** | Pattern separating read and write operations | Data Architecture |
| **Circuit Breaker** | Design pattern preventing cascading failures by detecting and handling failures | Resilience |
| **Rate Limiting** | Technique controlling rate of incoming traffic to protect services | Security/Performance |
| **Horizontal Scaling** | Adding more machines/nodes to handle increased load | Scalability |
| **Vertical Scaling** | Adding resources (CPU, RAM) to existing machine | Scalability |
| **Load Balancer** | Distributes incoming network traffic across multiple servers | Infrastructure |
| **Replication** | Copying data from one database to another for redundancy | Database |
| **Sharding** | Partitioning data across multiple database instances | Database |
| **Partitioning** | Dividing table data into manageable chunks within same database | Database |
| **B-tree Index** | Balanced tree data structure for efficient data retrieval | Database |
| **BRIN Index** | Block Range Index for time-series or ordered data | Database |
| **GIN Index** | Generalized Inverted Index for array/JSONB data | Database |
| **JWT (JSON Web Token)** | Compact, URL-safe token format for authentication | Security |
| **OAuth 2.0** | Authorization framework granting third-party applications limited access | Security |
| **RBAC (Role-Based Access Control)** | Security model restricting system access based on user roles | Security |
| **TLS 1.3** | Latest version of Transport Layer Security protocol for encrypted communication | Security |
| **AES-256** | Advanced Encryption Standard with 256-bit key length | Encryption |
| **KPI (Key Performance Indicator)** | Quantifiable metrics for evaluating success | Business |
| **SLA (Service Level Agreement)** | Contractual commitment between service provider and customer | Business |
| **SLI (Service Level Indicator)** | Measured aspect of service quality | Observability |
| **SLO (Service Level Objective)** | Target value for a service level indicator | Observability |
| **P99 Latency** | 99th percentile of latency measurements | Performance |
| **QPS (Queries Per Second)** | Throughput metric measuring queries processed | Performance |
| **TPS (Transactions Per Second)** | Throughput metric measuring transactions processed | Performance |

### Business Terminology

| Term | Definition | Context |
|------|------------|---------|
| **Agent** | B2B customer who operates their own branded poker platform using our infrastructure | Business Model |
| **Club** | Agent-created poker room with customized settings, branding, and rules | Multi-Tenancy |
| **White-label** | Product/service rebranded by another company as their own | Business Model |
| **Multi-tenant** | Architecture where single instance serves multiple customers with isolated data | Architecture |
| **Rakeback** | Percentage of rake returned to players as loyalty incentive | Marketing |
| **Player** | End-user who plays poker games on the platform | End User |
| **Operator** | Company managing the poker platform infrastructure | Service Provider |
| **Licensee** | Legal entity holding gaming license (usually the Agent) | Regulatory |
| **VIP Program** | Loyalty program for high-value players | Marketing |
| **Bankroll** | Total points available to a player for playing | Point Balance |
| **Entry** | Points required to join a point game table | Game Entry (Phase 3+ real-money: "Buy-in") |
| **Entry Fee** | Fixed fee to enter a tournament | Tournament |
| **Prize Pool** | Total money available to be won in a tournament | Tournament |
| **Guarantee** | Minimum prize pool guaranteed by the operator | Tournament |
| **Overlay** | Amount operator adds when prize pool exceeds entries collected | Tournament |
| **Freeroll** | Tournament with no entry fee | Marketing/Tournament |
| **Satellite** | Tournament where winners qualify for larger tournament | Tournament |
| **Point Game** | Non-tournament poker with flexible point entries | Game Type (Phase 3+ real-money: "Cash Game") |
| **Sit & Go** | Single-table tournament starting when seats filled | Game Type |
| **MTT (Multi-Table Tournament)** | Tournament spanning multiple tables with scheduled start | Game Type |

---

## B. Technology References & Links

### Cocos Creator Documentation

| Resource | URL | Purpose |
|----------|-----|---------|
| **Official Documentation** | https://docs.cocos.com/creator/3.8/ | Core reference |
| **TypeScript API** | https://docs.cocos.com/creator/3.8/api/ | API documentation |
| **Component Lifecycle** | https://docs.cocos.com/creator/3.8/manual/en/scripting/lifecycle-hooks.html | Game state management |
| **Networking** | https://docs.cocos.com/creator/3.8/manual/en/network/ | Socket.IO integration |
| **Prefab System** | https://docs.cocos.com/creator/3.8/manual/en/asset/prefab.html | Reusable game assets |
| **Spine Animation** | https://docs.cocos.com/creator/3.8/manual/en/animation/spine/ | Card/dealer animations |
| **Build & Publish** | https://docs.cocos.com/creator/3.8/manual/en/publish/ | iOS/Android deployment |

### Socket.IO Documentation

| Resource | URL | Purpose |
|----------|-----|---------|
| **Official Documentation** | https://socket.io/docs/v4/ | Core reference |
| **Server API** | https://socket.io/docs/v4/server-api/ | Room management |
| **Client API** | https://socket.io/docs/v4/client-api/ | Mobile client integration |
| **Rooms & Namespaces** | https://socket.io/docs/v4/rooms/ | Table isolation |
| **Emitting Events** | https://socket.io/docs/v4/emitting-events/ | Game state broadcasting |
| **Middleware** | https://socket.io/docs/v4/middlewares/ | Authentication |
| **Error Handling** | https://socket.io/docs/v4/server-instance/#error-handling | Resilience |

### Go (Golang) Resources

| Resource | URL | Purpose |
|----------|-----|---------|
| **Official Documentation** | https://go.dev/doc/ | Core reference |
| **Effective Go** | https://go.dev/doc/effective_go | Best practices |
| **Concurrency Patterns** | https://go.dev/doc/effective_go#concurrency | Goroutine patterns |
| **The Go Blog** | https://go.dev/blog/ | Deep-dive articles |
| **Package Documentation** | https://pkg.go.dev/std | Standard library |
| **Go Modules** | https://go.dev/doc/modules-get-started | Dependency management |
| **Testing** | https://go.dev/doc/tutorial/add-a-test | Unit testing |
| **Profiling** | https://go.dev/doc/diagnostics | Performance optimization |

### Go Gaming Libraries

| Library | URL | Purpose |
|---------|-----|---------|
| **gorilla/websocket** | https://github.com/gorilla/websocket | WebSocket server |
| **redis/go-redis** | https://github.com/redis/go-redis | Redis client |
| **go-pg** | https://github.com/go-pg/pg | PostgreSQL ORM |
| **Shopify/sarama** | https://github.com/Shopify/sarama | Kafka client |
| **golang-migrate** | https://github.com/golang-migrate/migrate | Database migrations |
| **uber-go/zap** | https://github.com/uber-go/zap | Structured logging |
| **stretchr/testify** | https://github.com/stretchr/testify | Testing utilities |

### Node.js & TypeScript Resources

| Resource | URL | Purpose |
|----------|-----|---------|
| **Node.js Documentation** | https://nodejs.org/docs/ | Core reference |
| **TypeScript Handbook** | https://www.typescriptlang.org/docs/ | Language guide |
| **NestJS Documentation** | https://docs.nestjs.com/ | Framework reference |
| **Express.js** | https://expressjs.com/ | HTTP server |
| **Socket.IO Server** | https://socket.io/docs/v4/server/ | Real-time API |
| **JWT Documentation** | https://jwt.io/ | Authentication |
| **class-validator** | https://github.com/typestack/class-validator | Request validation |
| **TypeORM** | https://typeorm.io/ | PostgreSQL ORM |

### PostgreSQL Resources

| Resource | URL | Purpose |
|----------|-----|---------|
| **Official Documentation** | https://www.postgresql.org/docs/15/ | Core reference |
| **Performance Tuning** | https://wiki.postgresql.org/wiki/Performance_Optimization | Optimization |
| **Partitioning** | https://www.postgresql.org/docs/15/ddl-partitioning.html | Table partitioning |
| **Row-Level Security** | https://www.postgresql.org/docs/15/ddl-rowsecurity.html | Data isolation |
| **JSONB Functions** | https://www.postgresql.org/docs/15/functions-json.html | JSON data |
| **Index Types** | https://www.postgresql.org/docs/15/indexes-types.html | Indexing strategies |
| **EXPLAIN ANALYZE** | https://www.postgresql.org/docs/15/sql-explain.html | Query analysis |

### Redis Resources

| Resource | URL | Purpose |
|----------|-----|---------|
| **Official Documentation** | https://redis.io/docs/ | Core reference |
| **Commands Reference** | https://redis.io/commands/ | Command list |
| **Data Types** | https://redis.io/docs/manual/data-types/ | Data structures |
| **Pub/Sub** | https://redis.io/docs/manual/pubsub/ | Event notifications |
| **Clustering** | https://redis.io/docs/manual/scaling/ | High availability |
| **Persistence** | https://redis.io/docs/manual/persistence/ | Data durability |

### Kafka Resources

| Resource | URL | Purpose |
|----------|-----|---------|
| **Official Documentation** | https://kafka.apache.org/documentation/ | Core reference |
| **Producer API** | https://kafka.apache.org/documentation/#producerapi | Event publishing |
| **Consumer API** | https://kafka.apache.org/documentation/#consumerapi | Event processing |
| **Streams API** | https://kafka.apache.org/documentation/streams/ | Stream processing |
| **Topic Management** | https://kafka.apache.org/documentation/#topicconfigs | Topic configuration |
| **Replication** | https://kafka.apache.org/documentation/#replication | Data redundancy |

### Security Standards

| Standard | URL | Purpose |
|---------|-----|---------|
| **TLS 1.3 RFC** | https://tools.ietf.org/html/rfc8446 | Encryption protocol |
| **JWT RFC** | https://tools.ietf.org/html/rfc7519 | Token format |
| **OAuth 2.0 RFC** | https://tools.ietf.org/html/rfc6749 | Authorization framework |
| **OWASP Top 10** | https://owasp.org/www-project-top-ten/ | Security risks |
| **AES Specification** | https://csrc.nist.gov/publications/detail/fips/197/final | Encryption algorithm |
| **PCI DSS** | https://www.pcisecuritystandards.org/ | Payment security (Phase 3+ real-money only) |

---

## C. Regulatory Compliance Notes

> **Note**: All regulatory compliance sections below apply only if real-money deployment is required in target markets. For point-based system (Phase 1-2), RNG certification remains relevant for game fairness, but PCI DSS, gaming license, KYC/AML requirements are reserved for Phase 3+ real-money expansion.



### RNG Certification Requirements

| Certification Body | Standards | Validity | Testing Frequency | Cost Estimate |
|--------------------|-----------|----------|--------------------|---------------|
| **eCOGRA** | iTech Labs, GLI | 2 years | Annual re-certification | $15,000 - $25,000 |
| **iTech Labs** | GLI-19, BMM Testlabs | 2 years | Annual review | $12,000 - $20,000 |
| **GLI (Gaming Laboratories International)** | GLI-19 | 2 years | Quarterly audits | $20,000 - $35,000 |
| **BMM Testlabs** | Various jurisdictions | 2 years | Annual verification | $18,000 - $30,000 |

### eCOGRA RNG Certification Process

**Phase 1: Source Code Review**
- Review random number generation algorithm
- Analyze entropy sources
- Verify seed generation
- Validate statistical properties

**Phase 2: Statistical Testing**
- Chi-square test
- Kolmogorov-Smirnov test
- Runs test
- Serial correlation test
- Poker test (specific to card games)

**Phase 3: Functional Testing**
- Verify card distribution uniformity
- Test all game scenarios
- Validate betting logic
- Confirm payout accuracy

**Phase 4: Continuous Monitoring**
- Periodic sampling of live games
- Statistical anomaly detection
- Ongoing compliance verification

### iTech Labs Standards

| Test Suite | Description | Pass Criteria |
|------------|-------------|---------------|
| **Diehard Tests** | Battery of statistical randomness tests | All tests pass |
| **FIPS 140-2** | Cryptographic module validation | Level 1 minimum (applies only to cryptographic modules used for real-money transactions; point-based operation uses standard TLS) |
| **AIS 31** | German gambling standards | Full compliance |
| **GLI-19** | Global gaming standard | Complete certification |

### Gaming License Requirements (Client Responsibility)

**Jurisdiction-Specific Requirements**

| Jurisdiction | License Type | Minimum Capital | Annual Fee | Compliance Duration |
|--------------|--------------|-----------------|------------|---------------------|
| **Malta (MGA)** | B2B/B2C | €100,000 | €25,000 | 5 years |
| **Isle of Man** | OGR | €850,000 | £35,000 | 5 years |
| **Gibraltar** | Remote Gambling | €100,000 | €30,000 | 5 years |
| **Curacao** | Master License | €40,000 | $0 | Unlimited |
| **New Jersey (DGE)** | Vendor Registration | $100,000 | $2,500 | Annual |
| **UK (UKGC)** | Remote Gambling | N/A | £17,500 | 5 years |

**Compliance Documentation Requirements**

1. **Game Rules Documentation**
   - Detailed game logic
   - Payout tables
   - House edge calculations
   - Game variation specifications

2. **Player Protection**
   - Self-exclusion mechanisms
   - Deposit limits
   - Time-out features
   - Responsible gambling resources

3. **Financial Controls**
    - AML (Anti-Money Laundering) procedures (Phase 3+ real-money only)
    - KYC (Know Your Customer) verification (Phase 3+ real-money only)
    - Payment processing records (Phase 3+ real-money only)
    - Audit trails

4. **Technical Documentation**
   - System architecture
   - Security measures
   - Disaster recovery plans
   - Business continuity procedures

### Data Protection Regulations

**GDPR (General Data Protection Regulation)**

| Requirement | Implementation | Compliance Status |
|--------------|----------------|-------------------|
| **Data Portability** | Export API for player data | ✅ Required |
| **Right to Erasure** | Account deletion workflow | ✅ Required |
| **Data Breach Notification** | Automated alert system | ✅ Required |
| **Data Processing Agreement** | Legal contracts | ✅ Required |
| **Data Protection Officer (DPO)** | Designated personnel | ⚠️ Client responsibility |
| **Consent Management** | Cookie/tracking consent | ✅ Required |

**CCPA (California Consumer Privacy Act)**

| Requirement | Implementation | Compliance Status |
|--------------|----------------|-------------------|
| **Do Not Sell My Info** | Opt-out mechanism | ✅ Required |
| **Right to Delete** | Account deletion workflow | ✅ Required |
| **Data Disclosure** | Transparency report | ✅ Required |
| **Opt-out Link** | Footer in UI | ✅ Required |

**Other Regional Requirements**

| Region | Regulation | Key Requirements |
|--------|------------|------------------|
| **Canada (PIPEDA)** | PIPEDA | Consent for data collection, reasonable security |
| **Brazil (LGPD)** | LGPD | Data minimization, data portability |
| **India (DPDP Act)** | DPDP Act 2023 | Data principal rights, consent management |
| **Australia (Privacy Act)** | APPs | Australian Privacy Principles compliance |

### Anti-Money Laundering (AML) Compliance (Phase 3+ Real-Money Only)

| Requirement | Implementation | Threshold |
|--------------|----------------|-----------|
| **Customer Due Diligence (CDD)** | Identity verification | All players |
| **Enhanced Due Diligence (EDD)** | Additional verification | VIP players >$10K deposits |
| **Transaction Monitoring** | Automated suspicious activity detection | All transactions |
| **Suspicious Activity Reports (SAR)** | Manual review and reporting | Threshold-based |
| **Geolocation Restrictions** | IP/prohibited country blocking | Real-time |

**AML Red Flags**
- Rapid deposits and withdrawals
- Multiple accounts from same IP
- Structured transactions below reporting thresholds
- Unusual betting patterns
- High-volume transactions from new accounts

> **Note**: For point-based system (Phase 1-2), implement simplified SuspiciousActivityMonitor for unusual point balance patterns without full AML compliance.

### Fairness & Transparency

**House Edge Disclosure Requirements**

| Game Type | Required Disclosure | Display Location |
|-----------|---------------------|------------------|
| **Point Games** | Platform fee percentage | Table rules, lobby (Phase 3+ real-money: "Cash Games") |
| **Tournaments** | Entry fee structure | Tournament info |
| **Side Games** | RTP (Return to Player) | Game rules |
| **VIP Program** | Rakeback percentages | Club dashboard |

**Audit Trail Requirements**

| Data Point | Retention Period | Purpose |
|------------|------------------|---------|
| **Game Logs** | 5 years minimum | Dispute resolution |
| **Financial Transactions** | 7 years minimum | Compliance audit |
| **Player Actions** | 1 year minimum | Anti-cheat analysis |
| **System Events** | 1 year minimum | Operational audit |

---

## D. Security Standards Reference

### Transport Layer Security (TLS) Configuration

**TLS 1.3 Requirements**

| Setting | Value | Rationale |
|---------|-------|-----------|
| **Protocol Version** | TLS 1.3 | Latest security, improved performance |
| **Cipher Suites** | TLS_AES_256_GCM_SHA384 | Strong encryption |
| **Certificate** | ECDSA (P-256) or RSA (2048+) | Strong key exchange |
| **OCSP Stapling** | Enabled | Certificate revocation checking |
| **HSTS** | Enabled | Enforce HTTPS |

**Certificate Management**
- Automatic certificate renewal via Let's Encrypt
- Certificate pinning for mobile apps
- Multi-domain SAN certificates for white-label
- Key rotation every 90 days

### Encryption Standards

**AES-256-GCM Encryption**

| Use Case | Implementation | Key Management |
|----------|----------------|----------------|
| **Data at Rest (DB)** | pgcrypto extension | Managed by cloud provider (AWS KMS, GCP KMS) |
| **Sensitive Fields** | Column-level encryption | Application-managed keys |
| **File Storage (S3)** | Server-side encryption | Cloud provider managed keys |
| **Backup Archives** | GPG encryption | Offline key storage |
| **API Payloads** | Application-level encryption | JWT-based key derivation |

**Key Hierarchy**

```
Master Key (AWS KMS/GCP KMS)
├─ Database Encryption Key (DEK)
│  └─ Table-specific keys
│     └─ Column-specific keys (PII data)
├─ Application Secrets Key
│  ├─ JWT signing keys
│  ├─ API keys
│  └─ OAuth secrets
└─ Backup Encryption Key
   └─ Archive keys
```

### Authentication & Authorization

**JWT Token Standards**

| Claim | Purpose | Format |
|-------|---------|--------|
| **iss** | Issuer (agent ID) | UUID |
| **sub** | Subject (player ID) | UUID |
| **aud** | Audience (service) | String |
| **exp** | Expiration | Numeric timestamp |
| **iat** | Issued at | Numeric timestamp |
| **jti** | Token ID (revocation) | UUID |
| **roles** | User permissions | String array |
| **agent_id** | Multi-tenant isolation | UUID |
| **club_id** | Club access | UUID (optional) |

**Token Lifetimes**

| Token Type | Lifetime | Refresh | Storage |
|------------|----------|---------|---------|
| **Access Token** | 1 hour | Yes | Memory (React Native) |
| **Refresh Token** | 30 days | No | Secure storage (Keychain) |
| **ID Token** | 1 hour | Yes | Memory |
| **API Key** | 1 year | Manual rotation | Server-side |

**RBAC (Role-Based Access Control)**

| Role | Permissions | Scope |
|------|-------------|-------|
| **Super Admin** | All system access | Platform-wide |
| **Agent Admin** | Agent management, clubs, players | Agent-level |
| **Club Manager** | Club settings, table management | Club-level |
| **Player** | Play games, view history | Personal data only |
| **Support Agent** | Read-only access for disputes | Ticket-assigned only |
| **Auditor** | Read-only logs, transactions | Platform-wide |

### OWASP Security Guidelines

**Top 10 Mitigations**

| OWASP Risk | Mitigation | Implementation |
|------------|-------------|----------------|
| **A01: Broken Access Control** | JWT validation, RLS policies | Middleware, DB triggers |
| **A02: Cryptographic Failures** | TLS 1.3, AES-256 | Infrastructure, code |
| **A03: Injection** | Parameterized queries, input validation | ORM, class-validator |
| **A04: Insecure Design** | Threat modeling, security reviews | Architecture phase |
| **A05: Security Misconfiguration** | Automated config scanning, secrets management | CI/CD, vault |
| **A06: Vulnerable Components** | Dependency scanning, SBOM | GitHub Dependabot |
| **A07: Identification & Failures** | Rate limiting, account lockout | API gateway, auth service |
| **A08: Data Integrity Failures** | Digital signatures, checksums | File uploads, API responses |
| **A09: Security Logging** | Immutable audit logs | Kafka, PostgreSQL |
| **A10: Server-Side Request Forgery** | Input validation, allowlists | API layer |

**Input Validation**

| Input Type | Validation Method | Sanitization |
|------------|-------------------|---------------|
| **Username** | Regex (alphanumeric, 3-20 chars) | HTML entity encode |
| **Email** | RFC 5322 validation | Lowercase |
| **Bet Amount** | Numeric, range check | Round to 2 decimals |
| **Chat Message** | Length limit (500 chars), profanity filter | HTML strip, XSS prevent |
| **Table Name** | Regex, reserved word block | HTML entity encode |

**Output Encoding**

| Context | Encoding Method |
|---------|-----------------|
| **HTML** | HTML entity encoding |
| **JavaScript** | JSON serialization |
| **URL** | URL encoding |
| **CSS** | CSS escaping |

### Network Security

**Rate Limiting Configuration**

| Endpoint | Limit | Window | Algorithm |
|----------|-------|--------|-----------|
| **WebSocket Connect** | 10/minute | 1 minute | Fixed window |
| **REST API** | 1000/minute | 1 minute | Sliding window |
| **Login** | 5/minute | 5 minutes | Fixed window |
| **Password Reset** | 3/hour | 1 hour | Fixed window |
| **Game Action** | 10/second | 1 second | Token bucket |

**IP Whitelisting (Optional)**

| Service | Default Access | Whitelist Required |
|----------|----------------|--------------------|
| **Admin Panel** | 0.0.0.0/0 | Optional (recommended) |
| **Game Server** | 0.0.0.0/0 | No |
| **API Internal** | 0.0.0.0/0 | Recommended (VPC internal) |
| **Database** | VPC internal only | Required |

**DDoS Protection**

| Layer | Protection Mechanism | Provider |
|-------|---------------------|----------|
| **Layer 3/4** | Cloudflare Enterprise, AWS Shield | Cloudflare, AWS |
| **Layer 7** | Rate limiting, WAF, bot detection | Cloudflare, Akamai |
| **Application** | Circuit breakers, auto-scaling | Kubernetes |

### Application Security

**Secrets Management**

| Secret | Storage | Rotation Policy |
|--------|---------|------------------|
| **Database Passwords** | AWS Secrets Manager / HashiCorp Vault | 90 days |
| **API Keys** | Environment variables, encrypted at rest | 180 days |
| **JWT Signing Keys** | Vault-managed HSM | 180 days |
| **OAuth Client Secrets** | AWS Secrets Manager | 180 days |
| **Third-party API Keys** | Vault, scoped keys | Per provider policy |

**Secure Coding Practices**

| Practice | Implementation |
|----------|----------------|
| **Least Privilege** | Minimal IAM roles, database grants |
| **Defense in Depth** | Multiple security layers |
| **Fail Secure** | Deny by default, explicit allow |
| **Secure Defaults** | No admin passwords pre-set |
| **Audit Logging** | All auth attempts, admin actions |
| **Code Review** | Mandatory peer review |
| **Static Analysis** | SAST tools in CI/CD |
| **Dependency Scanning** | Automated vulnerability checks |

**Third-Party Dependencies**

| Category | Tool | Frequency |
|----------|------|-----------|
| **Vulnerability Scanning** | npm audit, go mod audit | Every build |
| **License Compliance** | FOSSA, Snyk | Daily |
| **SBOM Generation** | Syft, Trivy | Every release |
| **Supply Chain Verification** | Sigstore, Cosign | Every deployment |

---

## E. Data Flow Diagrams

### Player Registration Flow

```
┌──────────┐
│  Client  │
│ (Mobile) │
└────┬─────┘
     │ POST /api/v1/auth/register
     │ {username, email, password, agent_code}
     ▼
┌──────────────┐
│   Nginx LB   │
│  (SSL Term)  │
└──────┬───────┘
       │
       ▼
┌────────────────────┐
│   API Gateway      │
│  (Node.js/NestJS)  │
│  - Rate Limit      │
│  - Input Validation│
└──────┬─────────────┘
       │
       ├─────────────────┐
       │                 │
       ▼                 ▼
┌──────────────┐  ┌──────────────┐
│ Auth Service │  │ Player DB    │
│  (JWT)       │  │  (PostgreSQL)│
│  - Hash PW   │  │  - Insert    │
│  - Create JWT│  │  - RLS Check │
└──────┬───────┘  └──────────────┘
       │
       │ JWT + Player Data
       ▼
┌──────────┐
│  Client  │
└──────────┘
```

### Game Table Join Flow

```
┌──────────┐
│  Client  │
└────┬─────┘
     │ Socket.IO connect
     │ {token}
     ▼
┌────────────────────┐
│   Socket.IO Server │
│     (Node.js)      │
│  - Validate JWT    │
│  - Extract playerId│
└──────┬─────────────┘
       │
       ▼
┌──────────────┐
│  Game Lobby  │
│ (Go Service) │
│  - List Tables│
└──────┬───────┘
       │
       │ Request: JOIN_TABLE
       │ {tableId, seat}
       ▼
┌─────────────────────┐
│   Table Manager     │
│    (Go Service)     │
│  - Validate seat    │
│  - Check balance    │
│  - Join table room  │
└──────┬──────────────┘
       │
       ├──────────────────┐
       │                  │
       ▼                  ▼
┌──────────────┐   ┌──────────────┐
│ Redis State  │   │ Kafka Event  │
│  - Table:... │   │ player_joined│
└──────────────┘   └──────────────┘
       │
       │ Broadcast: PLAYER_JOINED
       ▼
┌──────────┐
│  Client  │
└──────────┘
```

### Real-Time Game Action Flow

```
┌──────────┐
│  Client  │
└────┬─────┘
     │ Action: BET {amount: 100}
     ▼
┌────────────────────┐
│   Socket.IO Server │
│     (Node.js)      │
│  - Validate JWT    │
│  - Extract playerId│
└──────┬─────────────┘
       │
       │ Forward to Game Engine
       ▼
┌─────────────────────┐
│   Game Engine       │
│    (Go Service)     │
│  - Validate action  │
│  - Update game state│
│  - Calculate pot    │
└──────┬──────────────┘
       │
       ├──────────────────┬──────────────────┐
       │                  │                  │
       ▼                  ▼                  ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│ Redis State  │   │ Kafka Event  │   │ Anti-Cheat   │
│  - Update    │   │ game_action  │   │  (Async)     │
└──────────────┘   └──────────────┘   └──────────────┘
       │
       │ Broadcast: STATE_UPDATE
       ▼
┌──────────┐
│  Client  │
└──────────┘
```

### Financial Transaction Flow

```
┌──────────┐
│  Client  │
└────┬─────┘
     │ Action: DEPOSIT {amount, method}
     ▼
┌────────────────────┐
│   API Gateway      │
│  (Node.js/NestJS)  │
└──────┬─────────────┘
       │
       ▼
┌─────────────────────┐
│  Payment Service    │
│    (Node.js)        │
│  - Validate amount  │
│  - Check limits     │
└──────┬──────────────┘
       │
       ├──────────────────┐
       │                  │
       ▼                  ▼
┌──────────────┐   ┌──────────────┐
│ Payment GW   │   │ Player DB    │
│ (Stripe/...) │   │  - Balance   │
└──────┬───────┘   └──────────────┘
       │
       │ Payment Success
       ▼
┌─────────────────────┐
│  Transaction Log    │
│  (PostgreSQL)       │
│  - Insert transaction│
│  - Update balance   │
└──────┬──────────────┘
       │
       ├──────────────────┐
       │                  │
       ▼                  ▼
┌──────────────┐   ┌──────────────┐
│ Kafka Event  │   │ Redis Cache  │
│  - tx_event  │   │  - Balance   │
└──────────────┘   └──────────────┘
       │
       │ Response: SUCCESS
       ▼
┌──────────┐
│  Client  │
└──────────┘
```

### Anti-Cheat Analysis Flow

```
┌─────────────┐
│   Kafka     │
│ game-actions│
└──────┬──────┘
       │ Stream
       ▼
┌─────────────────────┐
│   Anti-Cheat        │
│   Consumer (Go)     │
│  - Process messages │
└──────┬──────────────┘
       │
       ├─────────────────────────────────┐
       │                                 │
       ▼                                 ▼
┌─────────────────────┐        ┌─────────────────────┐
│  ML Model           │        │  Rule Engine        │
│  (LSTM)             │        │  - Collusion        │
│  - Bot detection    │        │  - Chip dumping     │
│  - Pattern analysis  │        │  - Unusual behavior  │
└──────┬──────────────┘        └──────┬──────────────┘
       │                               │
       └───────────────┬───────────────┘
                       │
                       ▼
              ┌────────────────┐
              │ Risk Score     │
              │ Calculation    │
              └────┬───────────┘
                   │
          ┌────────┴────────┐
          │                 │
          ▼                 ▼
   ┌──────────────┐  ┌──────────────┐
   │ Score < 0.8  │  │ Score >= 0.8 │
   │ - Log info   │  │ - Flag player│
   │ - No action  │  │ - Alert admin│
   └──────────────┘  └──────┬───────┘
                            │
                            ▼
                     ┌──────────────┐
                     │ Kafka Event  │
                     │ security-alert│
                     └──────────────┘
```

### Data Persistence Flow

```
┌─────────────┐
│   Kafka     │
│  Events     │
└──────┬──────┘
       │
       ├─────────────────────┬─────────────────────┐
       │                     │                     │
       ▼                     ▼                     ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Raw Events  │    │  Aggregated  │    │  Security    │
│  Consumer    │    │  Analytics   │    │  Consumer    │
└──────┬───────┘    └──────┬───────┘    └──────┬───────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  PostgreSQL  │    │  PostgreSQL  │    │  PostgreSQL  │
│  raw_events  │    │  analytics   │    │  security    │
│  (7-day ret) │    │  (90-day ret)│    │  (1-year ret)│
└──────────────┘    └──────────────┘    └──────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  S3 Archive  │    │  S3 Archive  │    │  S3 Archive  │
│  (Parquet)   │    │  (Parquet)   │    │  (CSV)       │
└──────────────┘    └──────────────┘    └──────────────┘
```

---

## F. API Specification Overview

### REST API Endpoints

#### Authentication API

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/auth/register` | Register new player | No |
| POST | `/api/v1/auth/login` | Player login | No |
| POST | `/api/v1/auth/logout` | Player logout | Yes |
| POST | `/api/v1/auth/refresh` | Refresh access token | Refresh token |
| POST | `/api/v1/auth/forgot-password` | Initiate password reset | No |
| POST | `/api/v1/auth/reset-password` | Complete password reset | No |

#### Player API

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/players/me` | Get current player profile | Yes |
| PUT | `/api/v1/players/me` | Update player profile | Yes |
| GET | `/api/v1/players/me/balance` | Get player balance | Yes |
| GET | `/api/v1/players/me/history` | Get hand history | Yes |
| GET | `/api/v1/players/me/transactions` | Get transaction history | Yes |

#### Club API (Agent/Club Manager)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/clubs` | List clubs (filtered by agent) | Agent |
| POST | `/api/v1/clubs` | Create new club | Agent |
| GET | `/api/v1/clubs/:id` | Get club details | Agent/Club Manager |
| PUT | `/api/v1/clubs/:id` | Update club settings | Agent/Club Manager |
| DELETE | `/api/v1/clubs/:id` | Delete club | Agent |
| GET | `/api/v1/clubs/:id/players` | List club players | Agent/Club Manager |
| GET | `/api/v1/clubs/:id/tables` | List club tables | Agent/Club Manager |

#### Table API

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/tables` | List available tables | Yes |
| GET | `/api/v1/tables/:id` | Get table details | Yes |
| POST | `/api/v1/tables` | Create table (club manager) | Club Manager |
| PUT | `/api/v1/tables/:id` | Update table settings | Club Manager |
| DELETE | `/api/v1/tables/:id` | Close table | Club Manager |

#### Tournament API

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/tournaments` | List tournaments | Yes |
| GET | `/api/v1/tournaments/:id` | Get tournament details | Yes |
| POST | `/api/v1/tournaments/:id/register` | Register for tournament | Yes |
| POST | `/api/v1/tournaments` | Create tournament | Club Manager |
| PUT | `/api/v1/tournaments/:id` | Update tournament | Club Manager |

#### Financial API

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/transactions` | List transactions (filtered) | Agent |
| POST | `/api/v1/transactions/deposit` | Deposit funds | Player |
| POST | `/api/v1/transactions/withdraw` | Withdraw funds | Player |
| GET | `/api/v1/transactions/:id` | Get transaction details | Owner/Agent |
| GET | `/api/v1/reports/financial` | Financial report | Agent |
| GET | `/api/v1/reports/rake` | Rake report | Agent |

#### Admin API (Agent Level)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/admin/agents/me` | Get agent profile | Agent |
| PUT | `/api/v1/admin/agents/me` | Update agent profile | Agent |
| GET | `/api/v1/admin/players` | List all players (filtered) | Agent |
| GET | `/api/v1/admin/players/:id` | Get player details | Agent |
| PUT | `/api/v1/admin/players/:id/balance` | Adjust player balance | Agent |
| POST | `/api/v1/admin/players/:id/ban` | Ban player | Agent |
| GET | `/api/v1/admin/reports/daily` | Daily report | Agent |
| GET | `/api/v1/admin/reports/monthly` | Monthly report | Agent |

### WebSocket Events

#### Client → Server Events

| Event | Payload | Description |
|-------|---------|-------------|
| `connect` | `{token}` | Authenticate connection |
| `join_table` | `{tableId, seat}` | Join poker table |
| `leave_table` | `{}` | Leave current table |
| `player_action` | `{action, data}` | Player bet/raise/fold/check |
| `chat_message` | `{message}` | Send chat message |
| `ping` | `{}` | Keep-alive ping |

#### Server → Client Events

| Event | Payload | Description |
|-------|---------|-------------|
| `connection_success` | `{playerId, username}` | Connection authenticated |
| `error` | `{code, message}` | Error notification |
| `table_state` | `{tableId, gameState}` | Complete table state |
| `player_joined` | `{playerId, username, seat}` | New player joined |
| `player_left` | `{playerId, seat}` | Player left table |
| `game_action` | `{playerId, action, data}` | Player action broadcast |
| `hand_complete` | `{handId, winner, pot}` | Hand result |
| `chat_message` | `{playerId, username, message}` | Chat message broadcast |

### API Response Format

**Success Response**

```json
{
  "success": true,
  "data": {
    // Response data
  },
  "meta": {
    "timestamp": "2026-01-28T10:30:00Z",
    "requestId": "req_abc123"
  }
}
```

**Error Response**

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  },
  "meta": {
    "timestamp": "2026-01-28T10:30:00Z",
    "requestId": "req_abc123"
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict (duplicate) |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |
| `INSUFFICIENT_BALANCE` | 400 | Not enough funds for action |
| `TABLE_FULL` | 400 | Table has no available seats |
| `INVALID_ACTION` | 400 | Action not valid for current state |

### Pagination

Standard pagination for list endpoints:

```
GET /api/v1/tables?page=1&limit=20&sort=created_at&order=desc
```

**Response**

```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "totalPages": 8
    }
  }
}
```

### Rate Limits

| Endpoint | Rate Limit | Burst |
|----------|------------|-------|
| Authentication endpoints | 10 req/min | 20 |
| Player profile endpoints | 100 req/min | 200 |
| Table listing | 60 req/min | 100 |
| Financial operations | 20 req/min | 30 |
| Admin endpoints | 200 req/min | 300 |

**Rate Limit Headers**

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1643324400
```

---

## G. Database Schema Overview

### User Domain Tables

#### `agents`
```sql
CREATE TABLE agents (
    agent_id UUID PRIMARY KEY,
    business_name VARCHAR(100) NOT NULL,
    contact_email VARCHAR(255) NOT NULL UNIQUE,
    contact_phone VARCHAR(50),
    address JSONB,
    tax_id VARCHAR(100),
    settings JSONB,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
) PARTITION BY HASH (agent_id);

-- Indexes
CREATE INDEX idx_agents_email ON agents(contact_email);
CREATE INDEX idx_agents_status ON agents(status);
```

#### `clubs`
```sql
CREATE TABLE clubs (
    club_id UUID PRIMARY KEY,
    agent_id UUID NOT NULL REFERENCES agents(agent_id),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    logo_url VARCHAR(500),
    settings JSONB,
    rake_config JSONB,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(agent_id, slug)
);

-- Indexes
CREATE INDEX idx_clubs_agent ON clubs(agent_id);
CREATE INDEX idx_clubs_status ON clubs(status);
```

#### `players`
```sql
CREATE TABLE players (
    player_id UUID PRIMARY KEY,
    agent_id UUID NOT NULL REFERENCES agents(agent_id),
    club_id UUID REFERENCES clubs(club_id),
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    avatar_url VARCHAR(500),
    balance DECIMAL(15,2) DEFAULT 0.00,
    vip_level INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(agent_id, username)
) PARTITION BY HASH (agent_id);

-- Indexes
CREATE INDEX idx_players_email ON players(email);
CREATE INDEX idx_players_agent_username ON players(agent_id, username);
CREATE INDEX idx_players_status ON players(status);
```

### Game Domain Tables

#### `tables`
```sql
CREATE TABLE tables (
    table_id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(club_id),
    name VARCHAR(100) NOT NULL,
    game_type VARCHAR(50) NOT NULL,
    variant VARCHAR(50),
    max_players INTEGER NOT NULL DEFAULT 9,
    small_blind DECIMAL(10,2),
    big_blind DECIMAL(10,2),
    min_buyin DECIMAL(10,2),
    max_buyin DECIMAL(10,2),
    rake_percentage DECIMAL(5,2),
    rake_cap DECIMAL(10,2),
    settings JSONB,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_tables_club ON tables(club_id);
CREATE INDEX idx_tables_status ON tables(status);
```

#### `hands`
```sql
CREATE TABLE hands (
    hand_id UUID PRIMARY KEY,
    table_id UUID NOT NULL REFERENCES tables(table_id),
    hand_number BIGSERIAL,
    started_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP,
    pot DECIMAL(15,2),
    rake DECIMAL(10,2),
    community_cards JSONB,
    action_history JSONB,
    winners JSONB,
    status VARCHAR(20) DEFAULT 'active'
) PARTITION BY RANGE (started_at);

-- Monthly partitions
CREATE TABLE hands_2026_01 PARTITION OF hands
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Indexes
CREATE INDEX idx_hands_table ON hands(table_id);
CREATE INDEX idx_hands_started_at ON hands(started_at);
CREATE INDEX idx_hands_status ON hands(status);
```

#### `hand_players`
```sql
CREATE TABLE hand_players (
    hand_player_id UUID PRIMARY KEY,
    hand_id UUID NOT NULL REFERENCES hands(hand_id),
    player_id UUID NOT NULL REFERENCES players(player_id),
    seat INTEGER NOT NULL,
    starting_stack DECIMAL(10,2),
    ending_stack DECIMAL(10,2),
    profit_loss DECIMAL(10,2),
    hole_cards JSONB,
    action_summary JSONB,
    won BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_hand_players_hand ON hand_players(hand_id);
CREATE INDEX idx_hand_players_player ON hand_players(player_id);
```

#### `tournaments`
```sql
CREATE TABLE tournaments (
    tournament_id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(club_id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    game_type VARCHAR(50) NOT NULL,
    variant VARCHAR(50),
    max_players INTEGER,
    buyin DECIMAL(10,2),
    entry_fee DECIMAL(10,2),
    prize_pool DECIMAL(15,2),
    guarantee DECIMAL(15,2),
    blind_structure JSONB,
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'scheduled',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_tournaments_club ON tournaments(club_id);
CREATE INDEX idx_tournaments_starts_at ON tournaments(starts_at);
CREATE INDEX idx_tournaments_status ON tournaments(status);
```

### Financial Domain Tables

#### `transactions`
```sql
CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY,
    player_id UUID NOT NULL REFERENCES players(player_id),
    type VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2),
    balance_after DECIMAL(15,2),
    reference_id VARCHAR(100),
    description TEXT,
    status VARCHAR(20) DEFAULT 'completed',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE transactions_2026_01 PARTITION OF transactions
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Indexes
CREATE INDEX idx_transactions_player ON transactions(player_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_status ON transactions(status);
```

#### `rake_records`
```sql
CREATE TABLE rake_records (
    rake_id UUID PRIMARY KEY,
    hand_id UUID REFERENCES hands(hand_id),
    table_id UUID NOT NULL REFERENCES tables(table_id),
    agent_id UUID NOT NULL REFERENCES agents(agent_id),
    club_id UUID NOT NULL REFERENCES clubs(club_id),
    amount DECIMAL(10,2) NOT NULL,
    recorded_at TIMESTAMP DEFAULT NOW()
) PARTITION BY RANGE (recorded_at);

-- Monthly partitions
CREATE TABLE rake_records_2026_01 PARTITION OF rake_records
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Indexes
CREATE INDEX idx_rake_records_agent ON rake_records(agent_id);
CREATE INDEX idx_rake_records_club ON rake_records(club_id);
CREATE INDEX idx_rake_records_recorded_at ON rake_records(recorded_at);
```

### Security Domain Tables

#### `audit_logs`
```sql
CREATE TABLE audit_logs (
    log_id UUID PRIMARY KEY,
    actor_id UUID NOT NULL,
    actor_type VARCHAR(20) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

-- Monthly partitions
CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Indexes
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
```

#### `security_events`
```sql
CREATE TABLE security_events (
    event_id UUID PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) DEFAULT 'low',
    player_id UUID REFERENCES players(player_id),
    table_id UUID REFERENCES tables(table_id),
    details JSONB,
    risk_score DECIMAL(5,2),
    resolved BOOLEAN DEFAULT FALSE,
    resolved_by UUID,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_security_events_player ON security_events(player_id);
CREATE INDEX idx_security_events_type ON security_events(event_type);
CREATE INDEX idx_security_events_severity ON security_events(severity);
CREATE INDEX idx_security_events_created_at ON security_events(created_at);
```

### Row-Level Security (RLS) Policies

```sql
-- Enable RLS on players
ALTER TABLE players ENABLE ROW LEVEL SECURITY;

-- Agent can only access their own players
CREATE POLICY agent_player_access ON players
    FOR ALL
    USING (agent_id = current_setting('app.agent_id')::UUID);

-- Enable RLS on hands
ALTER TABLE hands ENABLE ROW LEVEL SECURITY;

-- Players can only see hands from their agent's tables
CREATE POLICY player_hand_access ON hands
    FOR SELECT
    USING (
        table_id IN (
            SELECT table_id FROM tables
            WHERE club_id IN (
                SELECT club_id FROM clubs
                WHERE agent_id = current_setting('app.agent_id')::UUID
            )
        )
    );
```

---

## H. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| **1.0** | 2026-01-28 | Technical Team | Initial creation of Section 10: Appendices |
| | | | Added Glossary of Terms (A) |
| | | | Added Technology References & Links (B) |
| | | | Added Regulatory Compliance Notes (C) |
| | | | Added Security Standards Reference (D) |
| | | | Added Data Flow Diagrams (E) |
| | | | Added API Specification Overview (F) |
| | | | Added Database Schema Overview (G) |
| | | | Added Revision History (H) |

---

*End of Section 10*
