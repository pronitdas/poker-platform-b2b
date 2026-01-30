package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"poker-platform/internal/fraud"
)

// SessionPostgresStorage implements SessionStore for PostgreSQL
type SessionPostgresStorage struct {
	db *sql.DB
}

// NewSessionPostgresStorage creates a new PostgreSQL session storage
func NewSessionPostgresStorage(db *sql.DB) *SessionPostgresStorage {
	return &SessionPostgresStorage{db: db}
}

// CreateSession creates a new player session
func (s *SessionPostgresStorage) CreateSession(ctx context.Context, session *fraud.PlayerSession) error {
	query := `
		INSERT INTO player_sessions (
			session_id, player_id, table_id, agent_id, club_id,
			connected_at, disconnected_at, duration,
			total_hands, total_wins, total_losses,
			total_chips, starting_chips, ending_chips
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := s.db.ExecContext(ctx, query,
		session.SessionID,
		session.PlayerID,
		session.TableID,
		session.AgentID,
		session.ClubID,
		session.ConnectedAt,
		session.DisconnectedAt,
		session.Duration,
		session.TotalHands,
		session.TotalWins,
		session.TotalLosses,
		session.TotalChips,
		session.StartingChips,
		session.EndingChips,
	)

	return err
}

// GetSession retrieves a session by ID
func (s *SessionPostgresStorage) GetSession(ctx context.Context, sessionID string) (*fraud.PlayerSession, error) {
	query := `
		SELECT session_id, player_id, table_id, agent_id, club_id,
			   connected_at, disconnected_at, duration,
			   total_hands, total_wins, total_losses,
			   total_chips, starting_chips, ending_chips
		FROM player_sessions
		WHERE session_id = $1
	`

	session := &fraud.PlayerSession{}
	var disconnectedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.SessionID,
		&session.PlayerID,
		&session.TableID,
		&session.AgentID,
		&session.ClubID,
		&session.ConnectedAt,
		&disconnectedAt,
		&session.Duration,
		&session.TotalHands,
		&session.TotalWins,
		&session.TotalLosses,
		&session.TotalChips,
		&session.StartingChips,
		&session.EndingChips,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if disconnectedAt.Valid {
		session.DisconnectedAt = &disconnectedAt.Time
	}

	return session, nil
}

// GetPlayerSessions retrieves all sessions for a player
func (s *SessionPostgresStorage) GetPlayerSessions(ctx context.Context, playerID string, startTime, endTime time.Time) ([]fraud.PlayerSession, error) {
	query := `
		SELECT session_id, player_id, table_id, agent_id, club_id,
			   connected_at, disconnected_at, duration,
			   total_hands, total_wins, total_losses,
			   total_chips, starting_chips, ending_chips
		FROM player_sessions
		WHERE player_id = $1 AND connected_at BETWEEN $2 AND $3
		ORDER BY connected_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, playerID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSessions(rows)
}

// GetActiveSessions retrieves active sessions for a player
func (s *SessionPostgresStorage) GetActiveSessions(ctx context.Context, playerID string) ([]fraud.PlayerSession, error) {
	query := `
		SELECT session_id, player_id, table_id, agent_id, club_id,
			   connected_at, disconnected_at, duration,
			   total_hands, total_wins, total_losses,
			   total_chips, starting_chips, ending_chips
		FROM player_sessions
		WHERE player_id = $1 AND disconnected_at IS NULL
		ORDER BY connected_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSessions(rows)
}

// EndSession marks a session as ended
func (s *SessionPostgresStorage) EndSession(ctx context.Context, sessionID string, endingChips int64) error {
	query := `
		UPDATE player_sessions
		SET disconnected_at = $1, ending_chips = $2, duration = $1 - connected_at
		WHERE session_id = $3 AND disconnected_at IS NULL
	`

	now := time.Now()
	result, err := s.db.ExecContext(ctx, query, now, endingChips, sessionID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("session not found or already ended")
	}
	return nil
}

// UpdateSessionStats updates session statistics
func (s *SessionPostgresStorage) UpdateSessionStats(ctx context.Context, sessionID string, handsPlayed, totalWins, totalLosses int) error {
	query := `
		UPDATE player_sessions
		SET total_hands = $1, total_wins = $2, total_losses = $3
		WHERE session_id = $4
	`

	_, err := s.db.ExecContext(ctx, query, handsPlayed, totalWins, totalLosses, sessionID)
	return err
}

// GetAllPlayerIDs retrieves all unique player IDs
func (s *SessionPostgresStorage) GetAllPlayerIDs(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT player_id FROM player_sessions
		ORDER BY player_id
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []string
	for rows.Next() {
		var playerID string
		if err := rows.Scan(&playerID); err != nil {
			return nil, err
		}
		players = append(players, playerID)
	}

	return players, rows.Err()
}

// DeleteOldSessions removes sessions older than the specified time
func (s *SessionPostgresStorage) DeleteOldSessions(ctx context.Context, before time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM player_sessions WHERE connected_at < $1
	`, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CreateSessionTable creates the sessions table if it doesn't exist
func (s *SessionPostgresStorage) CreateSessionTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS player_sessions (
			session_id VARCHAR(64) PRIMARY KEY,
			player_id VARCHAR(64) NOT NULL,
			table_id VARCHAR(64),
			agent_id VARCHAR(64),
			club_id VARCHAR(64),
			connected_at TIMESTAMP NOT NULL,
			disconnected_at TIMESTAMP,
			duration INTERVAL,
			total_hands INTEGER DEFAULT 0,
			total_wins INTEGER DEFAULT 0,
			total_losses INTEGER DEFAULT 0,
			total_chips BIGINT DEFAULT 0,
			starting_chips BIGINT DEFAULT 0,
			ending_chips BIGINT DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_sessions_player_id ON player_sessions(player_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_connected_at ON player_sessions(connected_at);
	`

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// scanSessions is a helper function to scan sessions from rows
func scanSessions(rows *sql.Rows) ([]fraud.PlayerSession, error) {
	var sessions []fraud.PlayerSession
	for rows.Next() {
		session := fraud.PlayerSession{}
		var disconnectedAt sql.NullTime

		err := rows.Scan(
			&session.SessionID,
			&session.PlayerID,
			&session.TableID,
			&session.AgentID,
			&session.ClubID,
			&session.ConnectedAt,
			&disconnectedAt,
			&session.Duration,
			&session.TotalHands,
			&session.TotalWins,
			&session.TotalLosses,
			&session.TotalChips,
			&session.StartingChips,
			&session.EndingChips,
		)
		if err != nil {
			return nil, err
		}

		if disconnectedAt.Valid {
			session.DisconnectedAt = &disconnectedAt.Time
		}

		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

// FingerprintPostgresStorage implements FingerprintDatabase for PostgreSQL
type FingerprintPostgresStorage struct {
	db *sql.DB
}

// NewFingerprintPostgresStorage creates a new PostgreSQL fingerprint storage
func NewFingerprintPostgresStorage(db *sql.DB) *FingerprintPostgresStorage {
	return &FingerprintPostgresStorage{db: db}
}

// StoreFingerprint stores a device fingerprint
func (s *FingerprintPostgresStorage) StoreFingerprint(ctx context.Context, fp fraud.DeviceFingerprint) error {
	query := `
		INSERT INTO device_fingerprints (
			player_id, fingerprint, user_agent, screen_resolution,
			color_depth, timezone, language, platform,
			hardware_concurrency, device_memory, touch_support,
			webgl_renderer, ip_address, first_seen, last_seen
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (fingerprint) DO UPDATE SET last_seen = EXCLUDED.last_seen
	`

	_, err := s.db.ExecContext(ctx, query,
		fp.PlayerID,
		fp.Fingerprint,
		fp.UserAgent,
		fp.ScreenResolution,
		fp.ColorDepth,
		fp.Timezone,
		fp.Language,
		fp.Platform,
		fp.HardwareConcurrency,
		fp.DeviceMemory,
		fp.TouchSupport,
		fp.WebGLRenderer,
		fp.IPAddress,
		fp.FirstSeen,
		fp.LastSeen,
	)

	return err
}

// GetFingerprintHistory retrieves fingerprint history for a player
func (s *FingerprintPostgresStorage) GetFingerprintHistory(ctx context.Context, playerID string) ([]fraud.DeviceFingerprint, error) {
	query := `
		SELECT player_id, fingerprint, user_agent, screen_resolution,
			   color_depth, timezone, language, platform,
			   hardware_concurrency, device_memory, touch_support,
			   webgl_renderer, ip_address, first_seen, last_seen
		FROM device_fingerprints
		WHERE player_id = $1
		ORDER BY last_seen DESC
	`

	rows, err := s.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFingerprints(rows)
}

// FindAccountsByFingerprint finds accounts sharing a device fingerprint
func (s *FingerprintPostgresStorage) FindAccountsByFingerprint(ctx context.Context, fingerprint string) ([]string, error) {
	query := `
		SELECT DISTINCT player_id FROM device_fingerprints WHERE fingerprint = $1
		ORDER BY player_id
	`

	rows, err := s.db.QueryContext(ctx, query, fingerprint)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPlayerIDs(rows)
}

// FindAccountsByIP finds accounts from the same IP
func (s *FingerprintPostgresStorage) FindAccountsByIP(ctx context.Context, ip string) ([]string, error) {
	query := `
		SELECT DISTINCT player_id FROM device_fingerprints WHERE ip_address = $1
		ORDER BY player_id
	`

	rows, err := s.db.QueryContext(ctx, query, ip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPlayerIDs(rows)
}

// FindAccountsByNetwork finds accounts in a network range
func (s *FingerprintPostgresStorage) FindAccountsByNetwork(ctx context.Context, networkPrefix string) ([]string, error) {
	query := `
		SELECT DISTINCT player_id FROM device_fingerprints
		WHERE ip_address LIKE $1
		ORDER BY player_id
	`

	rows, err := s.db.QueryContext(ctx, query, networkPrefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPlayerIDs(rows)
}

// GetLatestFingerprint gets the most recent fingerprint for a player
func (s *FingerprintPostgresStorage) GetLatestFingerprint(ctx context.Context, playerID string) (*fraud.DeviceFingerprint, error) {
	query := `
		SELECT player_id, fingerprint, user_agent, screen_resolution,
			   color_depth, timezone, language, platform,
			   hardware_concurrency, device_memory, touch_support,
			   webgl_renderer, ip_address, first_seen, last_seen
		FROM device_fingerprints
		WHERE player_id = $1
		ORDER BY last_seen DESC
		LIMIT 1
	`

	fp := &fraud.DeviceFingerprint{}
	var userAgent, screenResolution, timezone, language, platform, webGLRenderer, ipAddress sql.NullString
	var colorDepth, hardwareConcurrency int
	var deviceMemory float64
	var touchSupport bool

	err := s.db.QueryRowContext(ctx, query, playerID).Scan(
		&fp.PlayerID,
		&fp.Fingerprint,
		&userAgent,
		&screenResolution,
		&colorDepth,
		&timezone,
		&language,
		&platform,
		&hardwareConcurrency,
		&deviceMemory,
		&touchSupport,
		&webGLRenderer,
		&ipAddress,
		&fp.FirstSeen,
		&fp.LastSeen,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if userAgent.Valid {
		fp.UserAgent = userAgent.String
	}
	if screenResolution.Valid {
		fp.ScreenResolution = screenResolution.String
	}
	fp.ColorDepth = colorDepth
	if timezone.Valid {
		fp.Timezone = timezone.String
	}
	if language.Valid {
		fp.Language = language.String
	}
	if platform.Valid {
		fp.Platform = platform.String
	}
	fp.HardwareConcurrency = hardwareConcurrency
	fp.DeviceMemory = deviceMemory
	fp.TouchSupport = touchSupport
	if webGLRenderer.Valid {
		fp.WebGLRenderer = webGLRenderer.String
	}
	if ipAddress.Valid {
		fp.IPAddress = ipAddress.String
	}

	return fp, nil
}

// FingerprintExists checks if a fingerprint exists
func (s *FingerprintPostgresStorage) FingerprintExists(ctx context.Context, fingerprint string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM device_fingerprints WHERE fingerprint = $1)
	`, fingerprint).Scan(&exists)
	return exists, err
}

// CreateFingerprintTable creates the fingerprints table if it doesn't exist
func (s *FingerprintPostgresStorage) CreateFingerprintTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS device_fingerprints (
			id SERIAL PRIMARY KEY,
			player_id VARCHAR(64) NOT NULL,
			fingerprint VARCHAR(128) NOT NULL,
			user_agent TEXT,
			screen_resolution VARCHAR(16),
			color_depth INTEGER,
			timezone VARCHAR(64),
			language VARCHAR(8),
			platform VARCHAR(32),
			hardware_concurrency INTEGER,
			device_memory FLOAT,
			touch_support BOOLEAN,
			webgl_renderer TEXT,
			ip_address VARCHAR(45),
			first_seen TIMESTAMP NOT NULL,
			last_seen TIMESTAMP NOT NULL,
			UNIQUE (fingerprint)
		);

		CREATE INDEX IF NOT EXISTS idx_fingerprints_player_id ON device_fingerprints(player_id);
		CREATE INDEX IF NOT EXISTS idx_fingerprints_fingerprint ON device_fingerprints(fingerprint);
		CREATE INDEX IF NOT EXISTS idx_fingerprints_ip ON device_fingerprints(ip_address);
	`

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// scanFingerprints is a helper function to scan fingerprints from rows
func scanFingerprints(rows *sql.Rows) ([]fraud.DeviceFingerprint, error) {
	var fps []fraud.DeviceFingerprint
	for rows.Next() {
		fp := fraud.DeviceFingerprint{}
		var userAgent, screenResolution, timezone, language, platform, webGLRenderer, ipAddress sql.NullString
		var colorDepth, hardwareConcurrency int
		var deviceMemory float64
		var touchSupport bool

		err := rows.Scan(
			&fp.PlayerID,
			&fp.Fingerprint,
			&userAgent,
			&screenResolution,
			&colorDepth,
			&timezone,
			&language,
			&platform,
			&hardwareConcurrency,
			&deviceMemory,
			&touchSupport,
			&webGLRenderer,
			&ipAddress,
			&fp.FirstSeen,
			&fp.LastSeen,
		)
		if err != nil {
			return nil, err
		}

		if userAgent.Valid {
			fp.UserAgent = userAgent.String
		}
		if screenResolution.Valid {
			fp.ScreenResolution = screenResolution.String
		}
		fp.ColorDepth = colorDepth
		if timezone.Valid {
			fp.Timezone = timezone.String
		}
		if language.Valid {
			fp.Language = language.String
		}
		if platform.Valid {
			fp.Platform = platform.String
		}
		fp.HardwareConcurrency = hardwareConcurrency
		fp.DeviceMemory = deviceMemory
		fp.TouchSupport = touchSupport
		if webGLRenderer.Valid {
			fp.WebGLRenderer = webGLRenderer.String
		}
		if ipAddress.Valid {
			fp.IPAddress = ipAddress.String
		}

		fps = append(fps, fp)
	}
	return fps, rows.Err()
}

// scanPlayerIDs is a helper function to scan player IDs from rows
func scanPlayerIDs(rows *sql.Rows) ([]string, error) {
	var players []string
	for rows.Next() {
		var playerID string
		if err := rows.Scan(&playerID); err != nil {
			return nil, err
		}
		players = append(players, playerID)
	}
	return players, rows.Err()
}
