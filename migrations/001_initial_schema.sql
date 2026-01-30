-- B2B Poker Platform Database Schema
-- PostgreSQL 15+ with partitioning for multi-tenancy

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================
-- AGENTS (B2B customers - club owners/administrators)
-- =====================================================
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    contact_name VARCHAR(255),
    phone VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'inactive')),
    settings JSONB DEFAULT '{}',
    branding JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_agents_email ON agents(email);
CREATE INDEX idx_agents_status ON agents(status);

-- =====================================================
-- CLUBS (Managed by agents)
-- =====================================================
CREATE TABLE clubs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'inactive')),
    config JSONB DEFAULT '{}',
    rake_config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_clubs_agent_id ON clubs(agent_id);
CREATE INDEX idx_clubs_status ON clubs(status);
CREATE INDEX idx_clubs_name ON clubs(name);

-- =====================================================
-- PLAYERS (Partitioned by agent_id for multi-tenant isolation)
-- =====================================================
CREATE TABLE players (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    username VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    password_hash VARCHAR(255),
    balance DECIMAL(15,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'banned', 'inactive')),
    profile JSONB DEFAULT '{}',
    statistics JSONB DEFAULT '{}',
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(agent_id, username)
) PARTITION BY HASH (agent_id);

-- Create 16 partitions for players (for even distribution)
CREATE TABLE players_partition_0 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 0);
CREATE TABLE players_partition_1 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 1);
CREATE TABLE players_partition_2 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 2);
CREATE TABLE players_partition_3 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 3);
CREATE TABLE players_partition_4 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 4);
CREATE TABLE players_partition_5 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 5);
CREATE TABLE players_partition_6 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 6);
CREATE TABLE players_partition_7 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 7);
CREATE TABLE players_partition_8 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 8);
CREATE TABLE players_partition_9 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 9);
CREATE TABLE players_partition_10 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 10);
CREATE TABLE players_partition_11 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 11);
CREATE TABLE players_partition_12 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 12);
CREATE TABLE players_partition_13 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 13);
CREATE TABLE players_partition_14 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 14);
CREATE TABLE players_partition_15 PARTITION OF players FOR VALUES WITH (MODULUS 16, REMAINDER 15);

CREATE INDEX idx_players_agent_id ON players(agent_id);
CREATE INDEX idx_players_club_id ON players(club_id);
CREATE INDEX idx_players_username ON players(agent_id, username);
CREATE INDEX idx_players_status ON players(status);

-- =====================================================
-- TABLES (Poker tables within clubs)
-- =====================================================
CREATE TABLE tables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    table_type VARCHAR(20) DEFAULT 'cash' CHECK (table_type IN ('cash', 'tournament', 'sit_go')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'maintenance')),
    max_players INT DEFAULT 9 CHECK (max_players BETWEEN 2 AND 9),
    min_buyin DECIMAL(15,2),
    max_buyin DECIMAL(15,2),
    small_blind DECIMAL(15,2) NOT NULL,
    big_blind DECIMAL(15,2) NOT NULL,
    ante DECIMAL(15,2) DEFAULT 0,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tables_club_id ON tables(club_id);
CREATE INDEX idx_tables_status ON tables(status);
CREATE INDEX idx_tables_type ON tables(table_type);

-- =====================================================
-- HANDS (Hand history - partitioned by month)
-- =====================================================
CREATE TABLE hands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_id UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    club_id UUID NOT NULL,
    hand_number BIGINT NOT NULL,
    game_type VARCHAR(20) DEFAULT 'texas_hold'em' CHECK (game_type IN ('texas_hold'em', 'omaha', 'omaha_hi')),
    hole_cards JSONB NOT NULL,
    community_cards JSONB NOT NULL,
    pot_amount DECIMAL(15,2) NOT NULL,
    rake_amount DECIMAL(15,2) DEFAULT 0,
    winners JSONB NOT NULL,
    actions JSONB NOT NULL,
    duration_ms INT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
) PARTITION BY RANGE (started_at);

-- Create monthly partitions for hands (example for 2026)
CREATE TABLE hands_2026_01 PARTITION OF hands FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE hands_2026_02 PARTITION OF hands FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE hands_2026_03 PARTITION OF hands FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE INDEX idx_hands_table_id ON hands(table_id);
CREATE INDEX idx_hands_club_id ON hands(club_id);
CREATE INDEX idx_hands_started_at ON hands(started_at DESC);
CREATE INDEX idx_hands_hand_number ON hands(table_id, hand_number);

-- =====================================================
-- TRANSACTIONS (Balance operations)
-- =====================================================
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL,
    agent_id UUID NOT NULL,
    club_id UUID NOT NULL,
    type VARCHAR(30) NOT NULL CHECK (type IN ('deposit', 'withdrawal', 'transfer', 'rake', 'win', 'bonus', 'adjustment')),
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    reference_type VARCHAR(50),
    reference_id UUID,
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Create monthly partitions for transactions
CREATE TABLE transactions_2026_01 PARTITION OF transactions FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE transactions_2026_02 PARTITION OF transactions FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE transactions_2026_03 PARTITION OF transactions FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE INDEX idx_transactions_player_id ON transactions(player_id);
CREATE INDEX idx_transactions_agent_id ON transactions(agent_id);
CREATE INDEX idx_transactions_club_id ON transactions(club_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);

-- =====================================================
-- AUDIT_LOGS (Append-only for compliance)
-- =====================================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID,
    club_id UUID,
    player_id UUID,
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Create monthly partitions for audit logs
CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE audit_logs_2026_02 PARTITION OF audit_logs FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE audit_logs_2026_03 PARTITION OF audit_logs FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE INDEX idx_audit_logs_agent_id ON audit_logs(agent_id);
CREATE INDEX idx_audit_logs_club_id ON audit_logs(club_id);
CREATE INDEX idx_audit_logs_player_id ON audit_logs(player_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- =====================================================
-- SECURITY_EVENTS (Anti-cheat and fraud detection)
-- =====================================================
CREATE TABLE security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL,
    player_id UUID,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('bot_detected', 'collusion', 'fraud', 'suspicious_login', 'multi_account', 'cheat_detected', 'alert')),
    severity VARCHAR(20) DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    description TEXT NOT NULL,
    evidence JSONB DEFAULT '{}',
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_security_events_agent_id ON security_events(agent_id);
CREATE INDEX idx_security_events_player_id ON security_events(player_id);
CREATE INDEX idx_security_events_type ON security_events(event_type);
CREATE INDEX idx_security_events_severity ON security_events(severity);
CREATE INDEX idx_security_events_created_at ON security_events(created_at DESC);

-- =====================================================
-- RNG_AUDIT_LOG (Shuffle audit trail for certification)
-- =====================================================
CREATE TABLE rng_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_id UUID NOT NULL,
    hand_id UUID NOT NULL,
    server_id VARCHAR(100) NOT NULL,
    seed BYTEA NOT NULL,
    seed_hash BYTEA NOT NULL,
    deck_before BYTEA NOT NULL,
    deck_after BYTEA NOT NULL,
    algorithm VARCHAR(50) NOT NULL,
    checksum BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_rng_audit_table_id ON rng_audit_log(table_id);
CREATE INDEX idx_rng_audit_hand_id ON rng_audit_log(hand_id);
CREATE INDEX idx_rng_audit_created_at ON rng_audit_log(created_at DESC);

-- =====================================================
-- ROW-LEVEL SECURITY POLICIES
-- =====================================================

-- Enable RLS on tenant-scoped tables
ALTER TABLE agents ENABLE ROW LEVEL SECURITY;
ALTER TABLE clubs ENABLE ROW LEVEL SECURITY;
ALTER TABLE players ENABLE ROW LEVEL SECURITY;
ALTER TABLE hands ENABLE ROW LEVEL SECURITY;
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_events ENABLE ROW LEVEL SECURITY;

-- Agents can only see themselves
CREATE POLICY agents_select ON agents
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY agents_update ON agents
    FOR UPDATE USING (auth.uid() = id);

-- Clubs are visible only to their agent
CREATE POLICY clubs_select ON clubs
    FOR SELECT USING (agent_id IN (
        SELECT id FROM agents WHERE auth.uid() = id
    ));

CREATE POLICY clubs_insert ON clubs
    FOR INSERT WITH CHECK (agent_id IN (
        SELECT id FROM agents WHERE auth.uid() = id
    ));

CREATE POLICY clubs_update ON clubs
    FOR UPDATE USING (agent_id IN (
        SELECT id FROM agents WHERE auth.uid() = id
    ));

-- Players are visible only within their agent's scope
CREATE POLICY players_select ON players
    FOR SELECT USING (agent_id IN (
        SELECT id FROM agents WHERE auth.uid() = agent_id
    ));

CREATE POLICY players_insert ON players
    FOR INSERT WITH CHECK (agent_id IN (
        SELECT id FROM agents WHERE auth.uid() = agent_id
    ));

CREATE POLICY players_update ON players
    FOR UPDATE USING (agent_id IN (
        SELECT id FROM agents WHERE auth.uid() = agent_id
    ));

-- Hands are visible only within agent scope
CREATE POLICY hands_select ON hands
    FOR SELECT USING (club_id IN (
        SELECT id FROM clubs WHERE agent_id IN (
            SELECT id FROM agents WHERE auth.uid() = agent_id
        )
    ));

-- Transactions are visible only within agent scope
CREATE POLICY transactions_select ON transactions
    FOR SELECT USING (agent_id IN (
        SELECT id FROM agents WHERE auth.uid() = agent_id
    ));

-- Audit logs are visible to super admins and the agent they belong to
CREATE POLICY audit_logs_select ON audit_logs
    FOR SELECT USING (
        agent_id IN (SELECT id FROM agents WHERE auth.uid() = id)
        OR auth.role() = 'super_admin'
    );

-- Security events are visible to the agent and super admins
CREATE POLICY security_events_select ON security_events
    FOR SELECT USING (
        agent_id IN (SELECT id FROM agents WHERE auth.uid() = agent_id)
        OR auth.role() = 'super_admin'
    );

-- =====================================================
-- FUNCTIONS
-- =====================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_clubs_updated_at BEFORE UPDATE ON clubs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_players_updated_at BEFORE UPDATE ON players
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tables_updated_at BEFORE UPDATE ON tables
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to create monthly partitions automatically
CREATE OR REPLACE FUNCTION create_monthly_partition()
RETURNS TRIGGER AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    start_date := DATE_TRUNC('month', NEW.created_at)::DATE;
    end_date := (start_date + INTERVAL '1 month')::DATE;
    partition_name := TG_TABLE_NAME || '_' || TO_CHAR(start_date, 'YYYY_MM');

    -- Check if partition exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_tables
        WHERE tablename = partition_name
        AND schemaname = 'public'
    ) THEN
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
            partition_name, TG_TABLE_NAME, start_date, end_date
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- VIEWS
-- =====================================================

-- Agent summary view
CREATE VIEW agent_summary AS
SELECT
    a.id AS agent_id,
    a.company_name,
    a.status,
    COUNT(DISTINCT c.id) AS club_count,
    COUNT(DISTINCT p.id) AS player_count,
    COALESCE(SUM(t.amount), 0) AS total_volume,
    a.created_at
FROM agents a
LEFT JOIN clubs c ON c.agent_id = a.id
LEFT JOIN players p ON p.agent_id = a.id
LEFT JOIN transactions t ON t.agent_id = a.id
GROUP BY a.id, a.company_name, a.status, a.created_at;

-- Club summary view
CREATE VIEW club_summary AS
SELECT
    c.id AS club_id,
    c.agent_id,
    c.name,
    c.status,
    COUNT(DISTINCT t.id) AS table_count,
    COUNT(DISTINCT p.id) AS player_count,
    COALESCE(SUM(h.pot_amount), 0) AS total_raked,
    c.created_at
FROM clubs c
LEFT JOIN tables tbl ON tbl.club_id = c.id
LEFT JOIN players p ON p.club_id = c.id
LEFT JOIN hands h ON h.club_id = c.id
GROUP BY c.id, c.agent_id, c.name, c.status, c.created_at;

-- Player statistics view
CREATE VIEW player_statistics AS
SELECT
    p.id AS player_id,
    p.agent_id,
    p.club_id,
    p.username,
    p.balance,
    p.status,
    COUNT(h.id) AS hands_played,
    COALESCE(SUM(h.pot_amount), 0) AS total_wagered,
    p.created_at,
    p.last_login_at
FROM players p
LEFT JOIN hands h ON h.id IS NOT NULL
GROUP BY p.id, p.agent_id, p.club_id, p.username, p.balance, p.status, p.created_at, p.last_login_at;
