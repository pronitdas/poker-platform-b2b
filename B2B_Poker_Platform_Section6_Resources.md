# Section 6: Resource Plan (Roles & Effort)

## 6.1 Core Team Structure

### Team Composition Overview

The B2B poker platform requires a **cross-functional team** with specialized expertise in real-time gaming, distributed systems, and mobile development. The team is structured to handle the full development lifecycle from architecture through production deployment.

### Core Team Roles and Responsibilities

| Role | Quantity | Engagement | Phase | Key Responsibilities |
|------|----------|------------|-------|---------------------|
| **Technical Architect** | 1 | Full-time | P1, P2 | System design, technology stack selection, code reviews, technical leadership |
| **Backend Engineer (Go)** | 2-3 | Full-time | P1, P2, P3 | Game engine development, real-time logic, performance optimization |
| **Backend Engineer (Node.js)** | 1-2 | Full-time | P1, P2, P3 | API development, user management, admin services |
| **Game Developer (Cocos)** | 2-3 | Full-time | P1, P2, P3 | Mobile client development, game UI, client-server synchronization |
| **Frontend Developer (Web)** | 2 | Full-time | P1, P2, P3 | Web admin panel, dashboard development, reporting UI |
| **ML Engineer** | 1 | Full-time | P2 | Anti-cheat algorithms, player behavior analysis, fraud detection |
| **DevOps Engineer** | 1-2 | Full-time | P1, P2, P3 | CI/CD pipelines, infrastructure provisioning, monitoring setup |
| **QA Engineer** | 2-3 | Full-time | P1, P2, P3 | Test automation, performance testing, security testing |
| **UI/UX Designer** | 1 | Full-time | P1, P2 | Design systems, user flows, visual assets, accessibility |
| **Project Manager** | 1 | Full-time | P1, P2, P3 | Sprint planning, risk management, stakeholder communication |
| **Product Owner** | 1 | Part-time | P1, P2, P3 | Requirements gathering, backlog management, user acceptance |

### Role-Specific Expertise Requirements

| Role | Required Skills | Experience Level | Critical Success Factor |
|------|------------------|------------------|------------------------|
| **Technical Architect** | DDD, microservices, Go, Node.js, PostgreSQL, Kafka | 10+ years | Ability to design for 100K+ concurrent players |
| **Backend (Go)** | Goroutines, channels, Redis, WebSocket, concurrency patterns | 5-7 years | Real-time latency optimization (<100ms) |
| **Backend (Node.js)** | NestJS, TypeScript, REST APIs, authentication | 5-7 years | Clean API design and security |
| **Game (Cocos)** | Cocos Creator 3.8+, TypeScript, mobile game optimization | 4-6 years | Smooth 60fps animations, 2s app load time |
| **Frontend (Web)** | React, TypeScript, responsive design, data visualization | 4-6 years | Admin panel usability at scale |
| **ML Engineer** | Python, scikit-learn, anomaly detection, real-time ML | 5-7 years | 95%+ fraud detection accuracy |
| **DevOps** | Kubernetes, Docker, AWS/GCP, Terraform, Prometheus | 5-7 years | Zero-downtime deployments, auto-scaling |
| **QA** | Automated testing, load testing (JMeter/Locust), security | 4-6 years | 99.9% test coverage, performance validation |
| **UI/UX** | Figma, mobile-first design, accessibility (WCAG 2.1) | 5-7 years | Intuitive B2B admin interface |
| **Project Manager** | Agile/Scrum, risk management, cross-team coordination | 8+ years | On-time delivery, scope management |
| **Product Owner** | B2B SaaS, poker domain knowledge, user research | 5-7 years | Clear requirements, stakeholder alignment |

---

## 6.2 Effort Distribution by Phase

### Phase-Based Resource Allocation

The project is executed in **three phases** with varying team sizes based on development priorities:

| Phase | Team Size | Duration | Person-Months | Primary Focus |
|-------|-----------|----------|---------------|---------------|
| **Phase 1: Foundation** | 15-22 members | 8 months | 120-176 | Core architecture, MVP game engine, basic APIs |
| **Phase 2: Enhancement** | 10-14 members | 5 months | 50-70 | Advanced features, anti-cheat, analytics |
| **Phase 3: Productionization** | 5-7 members | 2 months | 10-14 | Hardening, performance tuning, documentation |
| **Total** | - | **15 months** | **180-260** | - |

### Phase 1: Foundation (Months 1-8)

**Team Composition:** 15-22 full-time equivalents

| Role | Quantity | Allocation | Key Deliverables |
|------|----------|------------|------------------|
| Technical Architect | 1 | 100% | Architecture documents, code standards, technical decisions |
| Backend (Go) | 2-3 | 100% | Game engine core, table logic, state management |
| Backend (Node.js) | 1-2 | 100% | Authentication, user APIs, admin services foundation |
| Game (Cocos) | 2-3 | 100% | Mobile client MVP, card rendering, basic UI |
| Frontend (Web) | 2 | 100% | Admin panel MVP, user management UI |
| DevOps | 1-2 | 100% | CI/CD setup, Kubernetes infrastructure, monitoring |
| QA | 2-3 | 100% | Test framework, functional testing, load testing |
| UI/UX | 1 | 100% | Design system, user flows, visual assets |
| Project Manager | 1 | 100% | Sprint planning, team coordination, progress tracking |
| Product Owner | 1 | 50% | Backlog grooming, requirements clarification |

**Phase 1 Person-Months Calculation:**

- Full-time roles (9 roles × 1 person × 8 months) = 72 person-months
- Variable roles (Backend Go +2, Backend Node +1, Game Dev +1, DevOps +1, QA +1) = 6 additional person-months
- Part-time Product Owner (1 × 0.5 × 8) = 4 person-months
- **Total: 120-176 person-months** (depending on variable role allocations)

### Phase 2: Enhancement (Months 9-13)

**Team Composition:** 10-14 full-time equivalents

| Role | Quantity | Allocation | Key Deliverables |
|------|----------|------------|------------------|
| Technical Architect | 1 | 50% | Architecture reviews, optimization guidance |
| Backend (Go) | 2 | 100% | Performance tuning, advanced game features |
| Backend (Node.js) | 1 | 100% | Advanced APIs, reporting services |
| Game (Cocos) | 2 | 100% | Enhanced mobile features, animations |
| Frontend (Web) | 1-2 | 100% | Advanced admin features, analytics dashboard |
| ML Engineer | 1 | 100% | Anti-cheat system, player behavior analysis |
| DevOps | 1 | 100% | Scaling optimization, monitoring enhancements |
| QA | 2 | 100% | Regression testing, security testing |
| Project Manager | 1 | 100% | Release planning, bug prioritization |
| Product Owner | 1 | 50% | Feature validation, stakeholder demos |

**Phase 2 Person-Months Calculation:**

- Full-time roles (7 roles × 1 person × 5 months) = 35 person-months
- Part-time roles (Architect + Product Owner: 2 × 0.5 × 5) = 5 person-months
- Variable roles (Frontend +1) = 5 additional person-months
- **Total: 50-70 person-months**

### Phase 3: Productionization (Months 14-15)

**Team Composition:** 5-7 full-time equivalents

| Role | Quantity | Allocation | Key Deliverables |
|------|----------|------------|------------------|
| Technical Architect | 1 | 25% | Final review, documentation sign-off |
| Backend (Go) | 1 | 100% | Performance hardening, bug fixes |
| Backend (Node.js) | 1 | 100% | API stability, load testing fixes |
| Game (Cocos) | 1 | 100% | Final polish, app store submission |
| DevOps | 1 | 100% | Production deployment, monitoring finalization |
| QA | 1-2 | 100% | Final acceptance testing, security audit |
| Project Manager | 1 | 100% | Release coordination, handover |

**Phase 3 Person-Months Calculation:**

- Full-time roles (5 roles × 1 person × 2 months) = 10 person-months
- Part-time Architect (1 × 0.25 × 2) = 0.5 person-months
- Variable QA (+1) = 2 additional person-months
- **Total: 10-14 person-months**

### Total Effort Summary

| Phase | Team Size (Min-Max) | Duration | Person-Months (Min-Max) | Percentage of Total Effort |
|-------|---------------------|----------|--------------------------|---------------------------|
| Phase 1 | 15-22 | 8 months | 120-176 | 67% |
| Phase 2 | 10-14 | 5 months | 50-70 | 27% |
| Phase 3 | 5-7 | 2 months | 10-14 | 6% |
| **Total** | **15-22** | **15 months** | **180-260** | **100%** |

---

## 6.3 Skill Matrix and Competency Mapping

### Technology-Specific Expertise Distribution

| Technology Area | Required Competency Level | Resources Allocated | Primary Phases |
|-----------------|---------------------------|---------------------|----------------|
| **Go (Golang)** | Expert | Technical Architect, 2-3 Backend Engineers | P1, P2, P3 |
| **Node.js/TypeScript** | Advanced | 1-2 Backend Engineers, 1-2 Frontend Developers | P1, P2, P3 |
| **Cocos Creator 3.8+** | Advanced | 2-3 Game Developers | P1, P2, P3 |
| **React/TypeScript** | Intermediate-Advanced | 2 Frontend Developers | P1, P2, P3 |
| **PostgreSQL** | Advanced | Technical Architect, Backend Engineers | P1, P2, P3 |
| **Redis** | Intermediate-Advanced | Backend Engineers, DevOps | P1, P2, P3 |
| **Kafka** | Intermediate | Backend Engineers, ML Engineer | P1, P2 |
| **Kubernetes/Docker** | Advanced | 1-2 DevOps Engineers | P1, P2, P3 |
| **Machine Learning** | Intermediate-Advanced | 1 ML Engineer | P2 |
| **Mobile Optimization** | Intermediate-Advanced | Game Developers | P1, P2, P3 |

### Competency Level Definitions

| Level | Definition | Example Indicators |
|-------|------------|--------------------|
| **Expert** | Deep domain knowledge, architecture-level decisions, published industry work | 10+ years, open source contributions, conference speaking |
| **Advanced** | Production-level implementation, complex problem-solving, mentorship | 5-7 years, multiple successful projects, team lead experience |
| **Intermediate** | Competent implementation, standard patterns, independent delivery | 3-5 years, feature ownership, standard workflows |
| **Beginner** | Learning the stack, requires guidance, simpler tasks | 0-3 years, supervised development, training programs |

### Cross-Training Plan

| Team Member | Primary Skill | Secondary Skills to Develop | Training Period |
|-------------|---------------|------------------------------|-----------------|
| Backend (Go) | Go, real-time systems | Node.js, Kafka | 2 months (Phase 1) |
| Backend (Node.js) | Node.js, REST APIs | Go, PostgreSQL optimization | 2 months (Phase 1) |
| Game (Cocos) | Cocos Creator, mobile dev | Go game protocol, TypeScript | 1 month (Phase 1) |
| Frontend (Web) | React, TypeScript | API design, data visualization | 1 month (Phase 1) |
| DevOps | Kubernetes, CI/CD | Application monitoring, performance tuning | Ongoing |
| QA | Test automation | Security testing, load testing | 1 month (Phase 1) |

---

## 6.4 Hiring and Contracting Strategy

### Hiring Approach: Hybrid Model

The team combines **core full-time employees** with **specialized contractors** to optimize cost and expertise access.

#### Internal Hires (Full-Time Employees)

| Role | Hiring Priority | Time to Fill | Onboarding Period |
|------|----------------|--------------|-------------------|
| Technical Architect | Critical (Month 0) | 6-8 weeks | 2 weeks |
| Backend Engineer (Go) Lead | Critical (Month 0) | 4-6 weeks | 2 weeks |
| Project Manager | Critical (Month 0) | 4-6 weeks | 1 week |
| Backend Engineer (Go) | High (Month 1-2) | 4-6 weeks | 2 weeks |
| Game Developer (Cocos) Lead | High (Month 1) | 6-8 weeks | 3 weeks |
| DevOps Engineer | High (Month 1) | 4-6 weeks | 2 weeks |
| Backend Engineer (Node.js) | Medium (Month 2) | 4-6 weeks | 2 weeks |
| Frontend Developer | Medium (Month 2) | 4-6 weeks | 2 weeks |
| QA Engineer | Medium (Month 2) | 3-4 weeks | 1 week |

#### Contractors (Specialized Needs)

| Role | Contract Duration | Engagement Model | Justification |
|------|------------------|------------------|---------------|
| ML Engineer | 5 months (Phase 2 only) | Full-time contractor | Specialized expertise, project-specific |
| UI/UX Designer | 8 months (Phase 1-2) | Full-time contractor | Flexible design capacity, phase-dependent |
| Security Consultant | 2 weeks (Phase 3) | Fixed-price project | Penetration testing, audit certification |

### Hiring Channels

| Channel | Primary Use | Target Roles | Success Rate |
|---------|-------------|--------------|--------------|
| **LinkedIn Recruiter** | Senior roles (Architect, Leads) | Architect, Senior Engineers | 60% |
| **Toptal/Upwork** | Specialized contractors | ML Engineer, UI/UX Designer | 40% |
| **GitHub Jobs/Hired** | Backend, Game Developers | Backend (Go/Node.js), Game Devs | 50% |
| **Referrals** | All roles | All roles (priority pipeline) | 75% |
| **Local Tech Meetups** | Junior/Mid-level | QA, Frontend Developers | 35% |
| **Recruitment Agencies** | Urgent fills | Critical path roles (Architect) | 55% |

### Onboarding Checklist

| Phase | Activity | Owner | Timeline |
|-------|----------|-------|----------|
| **Pre-Start** | Equipment provisioning, access setup | HR + DevOps | Week -1 |
| **Week 1** | Technical orientation, architecture deep-dive, coding challenge | Tech Architect | Day 1-5 |
| **Week 2** | Team pairing, small feature implementation, code review practice | Tech Lead | Day 6-10 |
| **Week 3-4** | Independent feature delivery, cross-team collaboration | Project Manager | Day 11-20 |
| **Month 2** | Full production responsibilities, architectural input (senior roles) | Tech Architect | Day 21-40 |

### Risk Mitigation: Staffing Contingencies

| Risk | Impact | Mitigation Strategy | Backup Plan |
|------|--------|---------------------|-------------|
| **Key role unfilled** | Project delay | Parallel interviewing, agency backup | Contractor gap fill |
| **Team member turnover** | Knowledge loss | Documentation, code reviews, pair programming | Knowledge transfer sessions |
| **Skill gap** | Quality issues | Cross-training, external training budget | Consultant guidance |
| **Contractor unavailability** | Phase delay | Multiple contractor options, buffer time | Internal resource reallocation |

---

## 6.5 External Dependencies

### Consultants and Subject Matter Experts

| Consultant Type | Engagement Duration | Primary Deliverables | Cost Impact |
|-----------------|---------------------|----------------------|-------------|
| **Poker Domain Expert** | 40 hours (spread across P1-P2) | Game rules validation, edge case identification | Included in R&D budget |
| **Security Auditor** | 2 weeks (Phase 3) | Penetration testing report, security certification recommendations | Fixed-price project |
| **Legal Compliance** | Ad-hoc (as needed) | GDPR/CCPA compliance review, terms of service | Hourly consulting |
| **Mobile App Store Specialist** | 10 hours (Phase 3) | App submission guidance, approval support | Hourly consulting |

### Third-Party Vendors

| Vendor | Service | Engagement Type | Critical Path Dependency |
|--------|---------|----------------|-------------------------|
| **AWS/GCP** | Cloud infrastructure | Pay-as-you-go | Critical (Phase 1+) |
| **Datadog/New Relic** | APM and monitoring | Subscription | High (Phase 1+) |
| **GitHub/GitLab** | Code repository, CI/CD | Subscription | Critical (Phase 1+) |
| **Figma** | Design collaboration | Free/Team tier | Medium (Phase 1-2) |
| **Jira/Linear** | Project management | Subscription | High (Phase 1+) |
| **Postman** | API testing | Free/Team tier | Medium (Phase 1+) |
| **Load Testing Service (k6/Locust)** | Performance testing | Open source + cloud credits | High (Phase 2-3) |

### Integration Points

| External System | Integration Complexity | Required Effort | Phasing |
|------------------|---------------------|-----------------|---------|
| **Payment Gateway** | Medium (PCI compliance, fraud detection) | 1-2 person-months | Phase 2 |
| **Email/SMS Provider** | Low (standard APIs) | 0.5 person-months | Phase 1 |
| **CDN (CloudFront/Cloudflare)** | Low (static assets) | 0.25 person-months | Phase 1 |
| **Customer Support Ticketing** | Medium (custom integrations) | 1 person-month | Phase 2 |
| **Analytics Platform (Google Analytics/Mixpanel)** | Low-Medium (event tracking) | 1 person-month | Phase 2 |

### Vendor SLA Requirements

| Service | Critical SLA Metric | Required Level | Fallback Strategy |
|---------|--------------------|----------------|-------------------|
| **Cloud Provider** | Uptime | 99.99% | Multi-region redundancy |
| **APM/Monitoring** | Data retention | 90 days | Local logging backup |
| **CI/CD Pipeline** | Build time | <10 minutes | Self-hosted runners |
| **Database (Managed)** | Latency (P99) | <50ms | Connection pooling optimization |
| **Payment Gateway** | Processing time | <5 seconds | Multiple gateway providers |

---

## Summary

This resource plan provides a **structured approach** to building the B2B poker platform with:

✅ **Clear team structure**: 11 core roles with defined responsibilities and expertise levels
✅ **Phase-based allocation**: 180-260 person-months across 15 months, optimized for each phase's objectives
✅ **Skill mapping**: Technology-specific expertise distribution ensuring all critical areas covered
✅ **Hybrid hiring model**: Balance of full-time employees and specialized contractors for cost optimization
✅ **External dependencies**: Defined consultants and third-party vendors with clear engagement terms

The team composition ensures **high-quality delivery** with expertise in real-time gaming, distributed systems, mobile development, and B2B SaaS. The phased approach allows for **efficient resource utilization** while maintaining the flexibility to adapt to changing requirements and technical challenges.

---

*Next Section: Section 7 - Implementation Timeline*
