package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"poker-platform/internal/fraud"
	"poker-platform/internal/storage"
)

// AlertPostgresStorage implements AlertStorage for PostgreSQL
type AlertPostgresStorage struct {
	db *sql.DB
}

// NewAlertPostgresStorage creates a new PostgreSQL alert storage
func NewAlertPostgresStorage(db *sql.DB) *AlertPostgresStorage {
	return &AlertPostgresStorage{db: db}
}

// CreateAlert creates a new alert in the database
func (s *AlertPostgresStorage) CreateAlert(ctx context.Context, alert *fraud.AntiCheatAlert) error {
	query := `
		INSERT INTO fraud_alerts (
			id, player_id, agent_id, club_id, table_id, hand_id,
			alert_type, severity, score, evidence, status,
			reviewer_id, review_notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := s.db.ExecContext(ctx, query,
		alert.ID,
		alert.PlayerID,
		alert.AgentID,
		alert.ClubID,
		alert.TableID,
		alert.HandID,
		alert.AlertType,
		alert.Severity,
		alert.Score,
		strings.Join(alert.Evidence, "||"),
		alert.Status,
		nil, // reviewer_id
		nil, // review_notes
		alert.CreatedAt,
		alert.CreatedAt,
	)

	return err
}

// GetAlert retrieves an alert by ID
func (s *AlertPostgresStorage) GetAlert(ctx context.Context, alertID string) (*fraud.AntiCheatAlert, error) {
	query := `
		SELECT id, player_id, agent_id, club_id, table_id, hand_id,
			   alert_type, severity, score, evidence, status,
			   reviewer_id, review_notes, created_at, updated_at
		FROM fraud_alerts
		WHERE id = $1
	`

	alert := &fraud.AntiCheatAlert{}
	var evidence, reviewerID, reviewNotes sql.NullString

	err := s.db.QueryRowContext(ctx, query, alertID).Scan(
		&alert.ID,
		&alert.PlayerID,
		&alert.AgentID,
		&alert.ClubID,
		&alert.TableID,
		&alert.HandID,
		&alert.AlertType,
		&alert.Severity,
		&alert.Score,
		&evidence,
		&alert.Status,
		&reviewerID,
		&reviewNotes,
		&alert.CreatedAt,
		&alert.ReviewedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if evidence.Valid {
		alert.Evidence = strings.Split(evidence.String, "||")
	}
	if reviewerID.Valid {
		alert.ReviewedBy = reviewerID.String
	}
	if reviewNotes.Valid {
		alert.Notes = reviewNotes.String
	}

	return alert, nil
}

// GetPlayerAlerts retrieves all alerts for a player
func (s *AlertPostgresStorage) GetPlayerAlerts(ctx context.Context, playerID string, limit int) ([]*fraud.AntiCheatAlert, error) {
	query := `
		SELECT id, player_id, agent_id, club_id, table_id, hand_id,
			   alert_type, severity, score, evidence, status,
			   reviewer_id, review_notes, created_at, updated_at
		FROM fraud_alerts
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*fraud.AntiCheatAlert
	for rows.Next() {
		alert := &fraud.AntiCheatAlert{}
		var evidence, reviewerID, reviewNotes sql.NullString

		err := rows.Scan(
			&alert.ID,
			&alert.PlayerID,
			&alert.AgentID,
			&alert.ClubID,
			&alert.TableID,
			&alert.HandID,
			&alert.AlertType,
			&alert.Severity,
			&alert.Score,
			&evidence,
			&alert.Status,
			&reviewerID,
			&reviewNotes,
			&alert.CreatedAt,
			&alert.ReviewedAt,
		)
		if err != nil {
			return nil, err
		}

		if evidence.Valid {
			alert.Evidence = strings.Split(evidence.String, "||")
		}
		if reviewerID.Valid {
			alert.ReviewedBy = reviewerID.String
		}
		if reviewNotes.Valid {
			alert.Notes = reviewNotes.String
		}

		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

// GetAlertsByTimeRange retrieves alerts within a time range
func (s *AlertPostgresStorage) GetAlertsByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*fraud.AntiCheatAlert, error) {
	query := `
		SELECT id, player_id, agent_id, club_id, table_id, hand_id,
			   alert_type, severity, score, evidence, status,
			   reviewer_id, review_notes, created_at, updated_at
		FROM fraud_alerts
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := s.db.QueryContext(ctx, query, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// GetAlertsByType retrieves alerts by type
func (s *AlertPostgresStorage) GetAlertsByType(ctx context.Context, alertType string, limit int) ([]*fraud.AntiCheatAlert, error) {
	query := `
		SELECT id, player_id, agent_id, club_id, table_id, hand_id,
			   alert_type, severity, score, evidence, status,
			   reviewer_id, review_notes, created_at, updated_at
		FROM fraud_alerts
		WHERE alert_type = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, alertType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// GetAlertsBySeverity retrieves alerts by severity
func (s *AlertPostgresStorage) GetAlertsBySeverity(ctx context.Context, severity string, limit int) ([]*fraud.AntiCheatAlert, error) {
	query := `
		SELECT id, player_id, agent_id, club_id, table_id, hand_id,
			   alert_type, severity, score, evidence, status,
			   reviewer_id, review_notes, created_at, updated_at
		FROM fraud_alerts
		WHERE severity = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, severity, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// UpdateAlertStatus updates an alert's status
func (s *AlertPostgresStorage) UpdateAlertStatus(ctx context.Context, alertID, status, reviewerID, notes string) error {
	query := `
		UPDATE fraud_alerts
		SET status = $1, reviewer_id = $2, review_notes = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := s.db.ExecContext(ctx, query, status, reviewerID, notes, time.Now(), alertID)
	return err
}

// GetPendingAlerts retrieves all pending alerts
func (s *AlertPostgresStorage) GetPendingAlerts(ctx context.Context, limit int) ([]*fraud.AntiCheatAlert, error) {
	query := `
		SELECT id, player_id, agent_id, club_id, table_id, hand_id,
			   alert_type, severity, score, evidence, status,
			   reviewer_id, review_notes, created_at, updated_at
		FROM fraud_alerts
		WHERE status = 'pending'
		ORDER BY 
			CASE severity 
				WHEN 'critical' THEN 1 
				WHEN 'high' THEN 2 
				WHEN 'medium' THEN 3 
				WHEN 'low' THEN 4 
			END,
			created_at DESC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// GetAlertStats retrieves aggregated alert statistics
func (s *AlertPostgresStorage) GetAlertStats(ctx context.Context, start, end time.Time) (*storage.AlertStats, error) {
	stats := &storage.AlertStats{
		ByType:     make(map[string]int),
		BySeverity: make(map[string]int),
		ByAgent:    make(map[string]int),
	}

	// Total count
	var total int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM fraud_alerts WHERE created_at BETWEEN $1 AND $2
	`, start, end).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats.TotalAlerts = total

	// By type
	rows, err := s.db.QueryContext(ctx, `
		SELECT alert_type, COUNT(*) FROM fraud_alerts 
		WHERE created_at BETWEEN $1 AND $2 
		GROUP BY alert_type
	`, start, end)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var alertType string
		var count int
		rows.Scan(&alertType, &count)
		stats.ByType[alertType] = count
	}
	rows.Close()

	// By severity
	rows, err = s.db.QueryContext(ctx, `
		SELECT severity, COUNT(*) FROM fraud_alerts 
		WHERE created_at BETWEEN $1 AND $2 
		GROUP BY severity
	`, start, end)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var severity string
		var count int
		rows.Scan(&severity, &count)
		stats.BySeverity[severity] = count
	}
	rows.Close()

	// Pending count
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM fraud_alerts 
		WHERE status = 'pending' AND created_at BETWEEN $1 AND $2
	`, start, end).Scan(&stats.PendingReview)
	if err != nil {
		return nil, err
	}

	// Confirmed fraud
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM fraud_alerts 
		WHERE status = 'confirmed' AND created_at BETWEEN $1 AND $2
	`, start, end).Scan(&stats.ConfirmedFraud)
	if err != nil {
		return nil, err
	}

	// Dismissed
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM fraud_alerts 
		WHERE status = 'dismissed' AND created_at BETWEEN $1 AND $2
	`, start, end).Scan(&stats.Dismissed)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// DeleteOldAlerts removes alerts older than the specified time
func (s *AlertPostgresStorage) DeleteOldAlerts(ctx context.Context, before time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM fraud_alerts WHERE created_at < $1
	`, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// scanAlerts is a helper function to scan alerts from rows
func scanAlerts(rows *sql.Rows) ([]*fraud.AntiCheatAlert, error) {
	var alerts []*fraud.AntiCheatAlert
	for rows.Next() {
		alert := &fraud.AntiCheatAlert{}
		var evidence, reviewerID, reviewNotes sql.NullString

		err := rows.Scan(
			&alert.ID,
			&alert.PlayerID,
			&alert.AgentID,
			&alert.ClubID,
			&alert.TableID,
			&alert.HandID,
			&alert.AlertType,
			&alert.Severity,
			&alert.Score,
			&evidence,
			&alert.Status,
			&reviewerID,
			&reviewNotes,
			&alert.CreatedAt,
			&alert.ReviewedAt,
		)
		if err != nil {
			return nil, err
		}

		if evidence.Valid {
			alert.Evidence = strings.Split(evidence.String, "||")
		}
		if reviewerID.Valid {
			alert.ReviewedBy = reviewerID.String
		}
		if reviewNotes.Valid {
			alert.Notes = reviewNotes.String
		}

		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}

// CreateAlertTable creates the alerts table if it doesn't exist
func (s *AlertPostgresStorage) CreateAlertTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS fraud_alerts (
			id VARCHAR(64) PRIMARY KEY,
			player_id VARCHAR(64) NOT NULL,
			agent_id VARCHAR(64),
			club_id VARCHAR(64),
			table_id VARCHAR(64),
			hand_id VARCHAR(64),
			alert_type VARCHAR(32) NOT NULL,
			severity VARCHAR(16) NOT NULL,
			score DECIMAL(5,4) NOT NULL,
			evidence TEXT,
			status VARCHAR(16) NOT NULL DEFAULT 'pending',
			reviewer_id VARCHAR(64),
			review_notes TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_fraud_alerts_player_id ON fraud_alerts(player_id);
		CREATE INDEX IF NOT EXISTS idx_fraud_alerts_agent_id ON fraud_alerts(agent_id);
		CREATE INDEX IF NOT EXISTS idx_fraud_alerts_created_at ON fraud_alerts(created_at);
		CREATE INDEX IF NOT EXISTS idx_fraud_alerts_status ON fraud_alerts(status);
		CREATE INDEX IF NOT EXISTS idx_fraud_alerts_severity ON fraud_alerts(severity);
	`

	_, err := s.db.ExecContext(ctx, query)
	return err
}
