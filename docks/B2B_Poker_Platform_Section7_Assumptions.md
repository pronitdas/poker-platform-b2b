# Section 7: Assumptions

## 7.1 Technical Assumptions

### Technology and Platform Choices

| Category | Assumption | Rationale | Impact if Invalid |
|----------|------------|-----------|-------------------|
| **Primary Market** | Southeast Asian market priority | Supports smaller app size preference (Cocos advantage) | Larger app size may affect download rates in bandwidth-constrained regions |
| **Mobile Engine** | Cocos Creator 3.8+ for mobile client | 15-25 MB footprint vs 80-150 MB for Unity/Unreal | Higher user acquisition cost due to download friction |
| **Game Type (MVP)** | Texas Hold'em only | Reduces complexity, focuses resources on core gameplay | Delayed market entry if additional game types required in MVP |
| **Game Mode** | Point games only | Tournaments require additional logic (schedules, prize pools) | Extended development timeline if tournaments needed in MVP |
| **Economy Model** | Point-based system (no real-money transactions in app) | Simplifies compliance, reduces regulatory burden | Increased legal/regulatory complexity if real-money required |
| **Real-Money Path** | Real-money features require Phase 3+ conditional expansion | All infrastructure below is for point-based operation; payment gateway integration, PCI DSS, and KYC/AML apply only if real-money deployment is required in target markets | Platform architecture supports future real-money expansion with additional compliance modules |
| **Cloud Provider** | AWS, GCP, or Azure (client choice) | All three provide equivalent managed services | Potential re-architecture if specific provider features are mandatory |
| **Communication Protocol** | WebSocket-first architecture | Essential for sub-100ms game latency | Inacceptable game experience without WebSocket support |
| **Validation Strategy** | Server-side validation for all game logic | Prevents client-side cheating exploits | Security vulnerabilities if client can manipulate game state |

### Infrastructure Capabilities

| Resource | Assumption | Justification | Mitigation if Assumption Fails |
|----------|------------|---------------|-------------------------------|
| **Network Quality** | Average latency <50ms in target markets | Based on Southeast Asian network metrics | Implement adaptive UI, optimistic updates for higher latency |
| **Device Performance** | Target devices: 2GB+ RAM, quad-core CPU | Covers 80% of active smartphones in target markets | Progressive enhancement for lower-end devices |
| **Concurrent Users** | Peak load: 10,000 concurrent players per server cluster | Based on Go goroutine benchmarks | Horizontal scaling capabilities for unexpected growth |
| **Database Throughput** | PostgreSQL: 5,000 writes/sec, 20,000 reads/sec | Using connection pooling and read replicas | Read replicas for scale, caching layer for hot data |
| **Redis Capacity** | 100,000 ops/sec with <5ms P99 latency | Redis Cluster with 6 nodes | Automatic sharding and failover built-in |

### Technology Constraints

| Constraint | Details | Workaround Available |
|------------|---------|---------------------|
| **Initial Platforms** | iOS and Android only (no desktop/web in MVP) | Cross-platform Cocos Creator supports web in Phase 2 |
| **Languages** | English primary (multi-language support in Phase 2) | Internationalization architecture ready for future languages |
| **Authentication** | JWT-based auth only (OAuth integration in Phase 2) | Modular auth layer allows provider addition |
| **Payment Integration** | External only (Phase 3+ if real-money required) | API hooks for webhook notifications from external systems |

---

## 7.2 Business Assumptions

### Revenue Model and Pricing

| Assumption | Details | Business Rationale |
|------------|---------|-------------------|
| **B2B Model** | Agents pay subscription + transaction fees | Recurring revenue, predictable cash flow |
| **White-Label Capability** | Each agent can customize branding (logo, colors, UI text) | Increases marketability to diverse operators |
| **Multi-Tenant Support** | Architecture supports 100+ agents at scale | Economies of scale across customer base |
| **Pricing Flexibility** | Agent can set own rake structures and point values | Competitive advantage for agents in local markets |
| **No Revenue Share** | Platform fee only, no take of agent's game revenue | Simpler accounting, client retains full profit |

### Operational Model

| Aspect | Assumption | Implications |
|--------|------------|--------------|
| **Support Scope** | 3 months free post-delivery support | Bug fixes, minor adjustments, monitoring setup |
| **Maintenance Window** | Scheduled downtime <4 hours/month | Database backups, system upgrades |
| **SLA Targets** | 99.5% uptime for production environments | ~3.65 hours downtime allowed per month |
| **Data Ownership** | Client owns all data (agents, players, transactions) | Platform provides data export tools on request |
| **Compliance Responsibility** | Client handles gaming licenses and regulatory approvals | Platform provides audit logs and reporting tools only |

### Market Positioning

| Category | Assumption | Justification |
|----------|------------|---------------|
| **Competition** | Existing competitors with monolithic architectures | Microservices architecture provides agility and scalability advantage |
| **Customer Segment** | Small-to-medium poker clubs (1-50 tables) | MVP optimized for this segment, enterprise features in Phase 2 |
| **Geographic Focus** | Southeast Asia initial expansion | Growing online poker market, favorable regulatory environment |
| **Technology Maturity** | Target agents have technical teams or can hire | Required for white-label customization and API integrations |

---

## 7.3 Timeline Assumptions

### Development Approach

| Assumption | Detail | Impact if Violated |
|------------|--------|-------------------|
| **Parallel Development** | Frontend (Cocos) and backend (Go/Node.js) teams work simultaneously | Integrated testing critical at all milestones |
| **Requirements Freeze** | Product requirements finalized before Phase 1 development | Scope changes after Phase 1 cause significant delays |
| **Dedicated Resources** | Team works exclusively on this project (minimum 80% allocation) | Resource contention extends timeline 20-30% |
| **Clear Milestones** | 4-week sprint cycles with weekly demos | Missed demos indicate schedule risk requiring intervention |
| **Incremental Delivery** | Working software delivered every 4 weeks | Allows early feedback, reduces rework risk |

### Client Responsibilities

| Responsibility | Expected SLA | Consequence of Delay |
|----------------|--------------|----------------------|
| **Feedback on Demos** | Within 48 hours of delivery | Development pauses awaiting approval |
| **Requirement Clarifications** | Within 24 hours of request | Incorrect implementation possible without clarification |
| **Test Environment** | Provide staging environment before Phase 1 | Integration testing delayed until environment available |
| **Branding Assets** | Logos, color schemes, fonts by week 2 | Development continues with placeholders |
| **Access Credentials** | Cloud provider access by week 1 | Deployment planning delayed |

### Risk Buffer

| Risk Category | Buffer Included | Trigger for Buffer Usage |
|---------------|-----------------|-------------------------|
| **Technical Risks** | 20% additional time per phase | Performance issues, integration complexities |
| **Scope Changes** | Not included in MVP timeline | Requires timeline renegotiation |
| **External Dependencies** | 2-week buffer for app store approvals | Rejection or extended review processes |
| **Team Availability** | 10% buffer for personnel issues | Illness, turnover, competing priorities |

---

## 7.4 External Dependencies

### Third-Party Services

| Dependency | Criticality | Fallback Strategy |
|------------|-------------|-------------------|
| **Cloud Provider** | High (AWS/GCP/Azure) | Architecture designed for portability between providers |
| **CDN Services** | Medium (for static assets, branding) | Fallback to origin server with caching |
| **Monitoring Services** | Medium (Datadog/New Relic) | Open-source alternatives available (Prometheus/Grafana) |
| **Email/SMS Services** | Low (for notifications) | Alternative providers available via API |
| **Analytics Platform** | Low (Google Analytics/Mixpanel) | Self-hosted analytics if external not preferred |

### Regulatory and Legal

| Dependency | Client Responsibility | Platform Support |
|------------|---------------------|------------------|
| **Gaming Licenses** | Obtain and maintain for target jurisdictions | Audit logs, reporting tools, compliance dashboards |
| **Data Privacy Laws** | GDPR, PDPA compliance for player data | Data export, anonymization, retention policies |
| **Anti-Money Laundering** | KYC/AML processes for agents (Phase 3+ if real-money required) | Transaction history, suspicious activity reporting |
| **Payment Processing** | Integration with preferred payment gateways (Phase 3+ if real-money required) | Webhook handlers, balance synchronization APIs |
| **App Store Policies** | Compliance with Apple/Google review guidelines | Technical support for resubmission if rejected |

### Infrastructure Dependencies

| Dependency | Assumption | Impact if Unavailable |
|------------|------------|----------------------|
| **Internet Connectivity** | Stable broadband for server infrastructure | Requires multi-region deployment for redundancy |
| **DNS Services** | Reliable DNS provider (Route 53, Cloudflare) | Implement DNS failover across providers |
| **SSL Certificates** | Automated certificate management (Let's Encrypt) | Manual certificate rotation increases operational overhead |
| **Time Synchronization** | NTP servers for distributed systems | Clock drift causes data consistency issues |

---

## 7.5 Constraints and Limitations

### Technical Constraints

| Constraint | Limitation | Mitigation |
|------------|------------|------------|
| **Mobile Platforms** | iOS and Android only in MVP | Web support in Phase 2 using same Cocos codebase |
| **Game Variants** | Texas Hold'em only (Omaha in Phase 2) | Architecture designed for game type extensibility |
| **Tournament Support** | Point games only (tournaments in Phase 2) | Separation of concerns allows independent tournament module |
| **Real-Time Players per Table** | Maximum 9 players per table | Standard poker table size, future scaling to tournament tables |
| **WebSocket Connections** | 10,000 concurrent connections per server | Horizontal scaling via load balancer |

### Performance Limitations

| Metric | Limitation | Scaling Strategy |
|--------|------------|-------------------|
| **Game Action Latency** | P99 <100ms requires <50ms network latency | Deploy in regions close to players |
| **Database Write Throughput** | 5,000 writes/sec per PostgreSQL instance | Write sharding, read replicas for analytics |
| **Cache Memory** | 1 TB RAM max per Redis cluster | Redis Cluster with automatic sharding |
| **Concurrent Players** | 10,000 per game server cluster | Add server instances linearly |
| **API Request Rate** | 1,000 req/sec per API gateway instance | Auto-scale based on CPU/latency metrics |

### Operational Limitations

| Limitation | Details | Planned Resolution |
|------------|---------|-------------------|
| **Support Hours** | 24/7 monitoring, business hours for non-critical issues | On-call rotation for critical incidents |
| **Data Retention** | Game logs retained 90 days, audit logs 1 year | Configurable retention via policy |
| **Feature Requests** | Custom features require timeline negotiation | Roadmap prioritization process |
| **Backward Compatibility** | API versioning required for breaking changes | Semantic versioning, deprecation notices |
| **Multi-Language** | English only in MVP | i18n architecture ready for Phase 2 |

### Compliance Limitations

| Limitation | Platform Scope | Client Responsibility |
|------------|----------------|----------------------|
| **Jurisdiction Restrictions** | Platform provides blocking by IP/country | Client maintains blocklist for restricted regions |
| **Age Verification** | Platform provides age field and validation API | Client implements third-party age verification |
| **Player Funds Protection** | Point system only (no real-money handling) | Client manages any cash-out mechanisms externally |
| **Fair Play Certification** | Platform provides audit trails | Client submits to certification bodies |
| **Tax Reporting** | Platform provides transaction export | Client handles tax compliance locally |

---

## 7.6 Assumption Validation Strategy

### Ongoing Validation Process

| Assumption Category | Validation Method | Frequency |
|--------------------|-------------------|-----------|
| **Technical** | Performance testing, load testing, user testing | Every sprint |
| **Business** | Customer feedback, market research, competitive analysis | Monthly |
| **Timeline** | Sprint reviews, burndown tracking, milestone assessments | Weekly |
| **External Dependencies** | Service level monitoring, contract reviews | Continuous |

### Risk Triggers

| Trigger | Action Required |
|---------|----------------|
| **Game latency >100ms P99** | Investigate network, optimize code, consider regional deployment |
| **Client feedback delayed >5 days** | Pause development, escalate to project management |
| **Scope change request >10% of remaining work** | Assess impact, negotiate timeline, update plan |
| **External service SLA breach** | Activate fallback strategy, review SLA terms |
| **App store rejection** | Technical support for resubmission, address review feedback |

### Communication Protocol

| Situation | Communication Method | Response Time |
|-----------|---------------------|---------------|
| **Critical Issue** | Phone/video call | Within 2 hours |
| **Major Blocker** | Email + Slack | Within 8 hours |
| **Standard Question** | Email | Within 24 hours |
| **Project Update** | Weekly report | Every Friday |
| **Demo/Review** | Video conference | Every 4 weeks |

---

## Summary

This section documents the foundational assumptions underpinning the B2B poker platform proposal:

✅ **Technical Assumptions**: Southeast Asian market focus, Cocos Creator for mobile, Texas Hold'em MVP, WebSocket-first architecture, server-side validation

✅ **Business Assumptions**: B2B subscription model, white-label customization, multi-tenant architecture, client-managed compliance

✅ **Timeline Assumptions**: Parallel development, requirements freeze, dedicated resources, 48-hour feedback SLA

✅ **External Dependencies**: Cloud infrastructure, third-party monitoring, gaming licenses, payment processing

✅ **Constraints and Limitations**: Platform-specific limits (9 players/table, Texas Hold'em only), operational boundaries, compliance scope

These assumptions serve as the basis for accurate cost estimation, timeline planning, and risk management. Regular validation and open communication ensure the project stays aligned with reality as development progresses.

---

*Next Section: Section 8 - Risk Assessment and Mitigation*
