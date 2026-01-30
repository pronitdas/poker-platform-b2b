package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// ClickHouseConfig holds ClickHouse connection configuration
type ClickHouseConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Database     string        `yaml:"database"`
	Username     string        `yaml:"username"`
	Password     string        `yaml:"password"`
	Secure       bool          `yaml:"secure"`
	MaxOpenConns int           `yaml:"max_open_conns"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
	ConnTimeout  time.Duration `yaml:"conn_timeout"`
}

// ClickHouseAnalytics implements AnalyticsRepository for ClickHouse
type ClickHouseAnalytics struct {
	db clickhouse.Conn
}

// NewClickHouseAnalytics creates a new ClickHouse analytics repository
func NewClickHouseAnalytics(ctx context.Context, config ClickHouseConfig) (*ClickHouseAnalytics, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		TLS: &tls.Config{InsecureSkipVerify: config.Secure},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &ClickHouseAnalytics{db: conn}, nil
}

// CreateTables creates the analytics tables if they don't exist
func (ch *ClickHouseAnalytics) CreateTables(ctx context.Context) error {
	queries := []string{
		// Hand analytics table
		`CREATE TABLE IF NOT EXISTS hand_analytics (
			event_id String,
			event_type String,
			hand_id String,
			table_id String,
			game_type String,
			betting_type String,
			player_id String,
			seat_number Int32,
			position String,
			chips_before Int64,
			chips_after Int64,
			total_pot Int64,
			rake_amount Int64,
			action_type String,
			action_amount Int64,
			action_time_ms Int64,
			timestamp DateTime64(3),
			session_id String,
			agent_id String,
			club_id String,
			duration_ms Int64,
			num_players Int32,
			street_reached String
		) ENGINE = ReplacingMergeTree(timestamp)
		ORDER BY (hand_id, player_id, timestamp)`,

		// Fraud alerts table
		`CREATE TABLE IF NOT EXISTS fraud_alerts_analytics (
			event_id String,
			event_type String,
			alert_id String,
			alert_type String,
			severity String,
			player_id String,
			table_id String,
			hand_id String,
			detection_type String,
			risk_score Float64,
			signal_strength Float64,
			details String,
			timestamp DateTime64(3),
			agent_id String,
			club_id String,
			resolved Bool,
			resolution_time Nullable(DateTime64(3))
		) ENGINE = ReplacingMergeTree(timestamp)
		ORDER BY (alert_id, player_id, timestamp)`,

		// Session analytics table
		`CREATE TABLE IF NOT EXISTS session_analytics (
			event_id String,
			event_type String,
			session_id String,
			player_id String,
			agent_id String,
			club_id String,
			table_id String,
			device_id String,
			ip_address String,
			country String,
			platform String,
			chips_deposited Int64,
			chips_withdrawn Int64,
			net_profit Int64,
			hands_played Int32,
			duration_ms Int64,
			timestamp DateTime64(3)
		) ENGINE = ReplacingMergeTree(timestamp)
		ORDER BY (session_id, player_id, timestamp)`,

		// Table stats table
		`CREATE TABLE IF NOT EXISTS table_stats_analytics (
			event_id String,
			event_type String,
			table_id String,
			game_type String,
			betting_type String,
			stake_level String,
			agent_id String,
			club_id String,
			avg_pot_size Int64,
			avg_hands_per_hour Float64,
			avg_players_active Float64,
			total_rake Int64,
			timestamp DateTime64(3),
			period_start DateTime64(3),
			period_end DateTime64(3)
		) ENGINE = ReplacingMergeTree(timestamp)
		ORDER BY (table_id, timestamp)`,
	}

	for _, query := range queries {
		if err := ch.db.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// RecordHandEvent records a hand analytics event
func (ch *ClickHouseAnalytics) RecordHandEvent(ctx context.Context, event *HandAnalyticsEvent) error {
	query := `
		INSERT INTO hand_analytics (
			event_id, event_type, hand_id, table_id, game_type, betting_type,
			player_id, seat_number, position, chips_before, chips_after,
			total_pot, rake_amount, action_type, action_amount, action_time_ms,
			timestamp, session_id, agent_id, club_id, duration_ms, num_players, street_reached
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	return ch.db.Exec(ctx, query,
		event.EventID, event.EventType, event.HandID, event.TableID,
		event.GameType, event.BettingType, event.PlayerID, event.SeatNumber,
		event.Position, event.ChipsBefore, event.ChipsAfter, event.TotalPot,
		event.RakeAmount, event.ActionType, event.ActionAmount,
		event.ActionTime.Milliseconds(), event.Timestamp, event.SessionID,
		event.AgentID, event.ClubID, event.DurationMS, event.NumPlayers,
		event.StreetReached,
	)
}

// RecordHandEvents records multiple hand analytics events in batch
func (ch *ClickHouseAnalytics) RecordHandEvents(ctx context.Context, events []*HandAnalyticsEvent) error {
	if len(events) == 0 {
		return nil
	}

	for _, event := range events {
		if err := ch.RecordHandEvent(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

// RecordFraudEvent records a fraud analytics event
func (ch *ClickHouseAnalytics) RecordFraudEvent(ctx context.Context, event *FraudAnalyticsEvent) error {
	query := `
		INSERT INTO fraud_alerts_analytics (
			event_id, event_type, alert_id, alert_type, severity, player_id,
			table_id, hand_id, detection_type, risk_score, signal_strength,
			details, timestamp, agent_id, club_id, resolved, resolution_time
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	var resolutionTime interface{}
	if event.ResolutionTime != nil {
		resolutionTime = *event.ResolutionTime
	}

	return ch.db.Exec(ctx, query,
		event.EventID, event.EventType, event.AlertID, event.AlertType,
		event.Severity, event.PlayerID, event.TableID, event.HandID,
		event.DetectionType, event.RiskScore, event.SignalStrength,
		event.Details, event.Timestamp, event.AgentID, event.ClubID,
		event.Resolved, resolutionTime,
	)
}

// RecordFraudEvents records multiple fraud analytics events in batch
func (ch *ClickHouseAnalytics) RecordFraudEvents(ctx context.Context, events []*FraudAnalyticsEvent) error {
	if len(events) == 0 {
		return nil
	}

	for _, event := range events {
		if err := ch.RecordFraudEvent(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

// RecordSessionEvent records a session analytics event
func (ch *ClickHouseAnalytics) RecordSessionEvent(ctx context.Context, event *SessionAnalyticsEvent) error {
	query := `
		INSERT INTO session_analytics (
			event_id, event_type, session_id, player_id, agent_id, club_id,
			table_id, device_id, ip_address, country, platform,
			chips_deposited, chips_withdrawn, net_profit, hands_played,
			duration_ms, timestamp
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	return ch.db.Exec(ctx, query,
		event.EventID, event.EventType, event.SessionID, event.PlayerID,
		event.AgentID, event.ClubID, event.TableID, event.DeviceID,
		event.IPAddress, event.Country, event.Platform, event.ChipsDeposited,
		event.ChipsWithdrawn, event.NetProfit, event.HandsPlayed,
		event.Duration.Milliseconds(), event.Timestamp,
	)
}

// RecordTableStats records table statistics
func (ch *ClickHouseAnalytics) RecordTableStats(ctx context.Context, event *TableAnalyticsEvent) error {
	query := `
		INSERT INTO table_stats_analytics (
			event_id, event_type, table_id, game_type, betting_type,
			stake_level, agent_id, club_id, avg_pot_size, avg_hands_per_hour,
			avg_players_active, total_rake, timestamp, period_start, period_end
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	return ch.db.Exec(ctx, query,
		event.EventID, event.EventType, event.TableID, event.GameType,
		event.BettingType, event.StakeLevel, event.AgentID, event.ClubID,
		event.AvgPotSize, event.AvgHandsPerHour, event.AvgPlayersActive,
		event.TotalRake, event.Timestamp, event.PeriodStart, event.PeriodEnd,
	)
}

// GetHandAnalytics retrieves hand analytics based on query
func (ch *ClickHouseAnalytics) GetHandAnalytics(ctx context.Context, query HandAnalyticsQuery) ([]HandAnalyticsEvent, error) {
	sql := `
		SELECT event_id, event_type, hand_id, table_id, game_type, betting_type,
			   player_id, seat_number, position, chips_before, chips_after,
			   total_pot, rake_amount, action_type, action_amount, action_time_ms,
			   timestamp, session_id, agent_id, club_id, duration_ms, num_players, street_reached
		FROM hand_analytics
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	if query.AgentID != "" {
		sql += " AND agent_id = ?"
		args = append(args, query.AgentID)
	}
	if query.ClubID != "" {
		sql += " AND club_id = ?"
		args = append(args, query.ClubID)
	}
	if query.PlayerID != "" {
		sql += " AND player_id = ?"
		args = append(args, query.PlayerID)
	}
	if query.TableID != "" {
		sql += " AND table_id = ?"
		args = append(args, query.TableID)
	}
	if !query.StartTime.IsZero() {
		sql += " AND timestamp >= ?"
		args = append(args, query.StartTime)
	}
	if !query.EndTime.IsZero() {
		sql += " AND timestamp <= ?"
		args = append(args, query.EndTime)
	}

	sql += " ORDER BY timestamp DESC"
	if query.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", query.Limit)
		if query.Offset > 0 {
			sql += fmt.Sprintf(" OFFSET %d", query.Offset)
		}
	}

	rows, err := ch.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []HandAnalyticsEvent
	for rows.Next() {
		var event HandAnalyticsEvent
		var actionTimeMs int64
		var durationMs int64

		err := rows.Scan(
			&event.EventID, &event.EventType, &event.HandID, &event.TableID,
			&event.GameType, &event.BettingType, &event.PlayerID, &event.SeatNumber,
			&event.Position, &event.ChipsBefore, &event.ChipsAfter, &event.TotalPot,
			&event.RakeAmount, &event.ActionType, &event.ActionAmount, &actionTimeMs,
			&event.Timestamp, &event.SessionID, &event.AgentID, &event.ClubID,
			&durationMs, &event.NumPlayers, &event.StreetReached,
		)
		if err != nil {
			return nil, err
		}

		event.ActionTime = time.Duration(actionTimeMs) * time.Millisecond
		event.DurationMS = durationMs
		events = append(events, event)
	}

	return events, rows.Err()
}

// GetFraudAnalytics retrieves fraud analytics based on query
func (ch *ClickHouseAnalytics) GetFraudAnalytics(ctx context.Context, query FraudAnalyticsQuery) ([]FraudAnalyticsEvent, error) {
	sql := `
		SELECT event_id, event_type, alert_id, alert_type, severity, player_id,
			   table_id, hand_id, detection_type, risk_score, signal_strength,
			   details, timestamp, agent_id, club_id, resolved, resolution_time
		FROM fraud_alerts_analytics
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	if query.AgentID != "" {
		sql += " AND agent_id = ?"
		args = append(args, query.AgentID)
	}
	if query.PlayerID != "" {
		sql += " AND player_id = ?"
		args = append(args, query.PlayerID)
	}
	if query.Severity != "" {
		sql += " AND severity = ?"
		args = append(args, query.Severity)
	}
	if query.Resolved != nil {
		sql += " AND resolved = ?"
		args = append(args, *query.Resolved)
	}
	if !query.StartTime.IsZero() {
		sql += " AND timestamp >= ?"
		args = append(args, query.StartTime)
	}
	if !query.EndTime.IsZero() {
		sql += " AND timestamp <= ?"
		args = append(args, query.EndTime)
	}

	sql += " ORDER BY timestamp DESC"
	if query.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", query.Limit)
	}

	rows, err := ch.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []FraudAnalyticsEvent
	for rows.Next() {
		var event FraudAnalyticsEvent
		var resolutionTime *time.Time

		err := rows.Scan(
			&event.EventID, &event.EventType, &event.AlertID, &event.AlertType,
			&event.Severity, &event.PlayerID, &event.TableID, &event.HandID,
			&event.DetectionType, &event.RiskScore, &event.SignalStrength,
			&event.Details, &event.Timestamp, &event.AgentID, &event.ClubID,
			&event.Resolved, &resolutionTime,
		)
		if err != nil {
			return nil, err
		}

		event.ResolutionTime = resolutionTime
		events = append(events, event)
	}

	return events, rows.Err()
}

// GetFraudTrend retrieves fraud trend data for analysis
func (ch *ClickHouseAnalytics) GetFraudTrend(ctx context.Context, query FraudTrendQuery) ([]FraudTrendPoint, error) {
	var interval string
	switch query.GroupBy {
	case time.Hour:
		interval = "toStartOfHour"
	case 24 * time.Hour:
		interval = "toStartOfDay"
	case 7 * 24 * time.Hour:
		interval = "toStartOfWeek"
	case 30 * 24 * time.Hour:
		interval = "toStartOfMonth"
	default:
		interval = "toStartOfDay"
	}

	sql := fmt.Sprintf(`
		SELECT
			%s(timestamp) as time_bucket,
			count() as total_alerts,
			sum(case when severity = 'high' then 1 else 0 end) as high_severity,
			sum(case when severity = 'medium' then 1 else 0 end) as medium_severity,
			sum(case when severity = 'low' then 1 else 0 end) as low_severity,
			sum(case when resolved = true then 1 else 0 end) as resolved_count,
			avg(risk_score) as avg_risk_score
		FROM fraud_alerts_analytics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY %s(timestamp)
		ORDER BY time_bucket
	`, interval, interval)

	rows, err := ch.db.Query(ctx, sql, query.StartTime, query.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []FraudTrendPoint
	for rows.Next() {
		var trend FraudTrendPoint
		err := rows.Scan(
			&trend.TimeBucket, &trend.TotalAlerts, &trend.HighSeverity,
			&trend.MediumSeverity, &trend.LowSeverity, &trend.ResolvedCount,
			&trend.AvgRiskScore,
		)
		if err != nil {
			return nil, err
		}
		trends = append(trends, trend)
	}

	return trends, rows.Err()
}

// GetPlayerStats retrieves player statistics
func (ch *ClickHouseAnalytics) GetPlayerStats(ctx context.Context, playerID string, period time.Duration) (*PlayerAnalyticsStats, error) {
	sql := `
		SELECT
			player_id,
			count() as total_hands,
			sum(chips_after - chips_before) as total_profit,
			0 as total_rake_paid,
			avg(duration_ms) as avg_session_duration,
			0 as win_rate,
			0 as avg_pot_size,
			max(timestamp) as last_active,
			min(timestamp) as first_seen
		FROM hand_analytics
		WHERE player_id = ? AND timestamp >= now() - interval ?
		GROUP BY player_id
	`

	rows, err := ch.db.Query(ctx, sql, playerID, period.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		stats := &PlayerAnalyticsStats{}
		var avgSessionDurationMs int64
		err := rows.Scan(
			&stats.PlayerID, &stats.TotalHandsPlayed, &stats.TotalProfit,
			&stats.TotalRakePaid, &avgSessionDurationMs, &stats.WinRate,
			&stats.AvgPotSize, &stats.LastActive, &stats.FirstSeen,
		)
		if err != nil {
			return nil, err
		}
		stats.AvgSessionDuration = time.Duration(avgSessionDurationMs) * time.Millisecond
		return stats, nil
	}

	return nil, rows.Err()
}

// GetRevenueStats retrieves revenue statistics
func (ch *ClickHouseAnalytics) GetRevenueStats(ctx context.Context, query RevenueQuery) (*RevenueStats, error) {
	sql := `
		SELECT
			sum(rake_amount) as total_rake,
			0 as total_deposits,
			0 as total_withdrawals,
			min(timestamp) as period_start,
			max(timestamp) as period_end
		FROM hand_analytics
		WHERE timestamp >= ? AND timestamp <= ?
	`

	rows, err := ch.db.Query(ctx, sql, query.StartTime, query.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		stats := &RevenueStats{}
		err := rows.Scan(
			&stats.TotalRake, &stats.TotalDeposits, &stats.TotalWithdrawals,
			&stats.PeriodStart, &stats.PeriodEnd,
		)
		if err != nil {
			return nil, err
		}
		stats.NetRevenue = stats.TotalRake
		return stats, nil
	}

	return nil, rows.Err()
}

// GetSessionAnalytics retrieves session analytics based on query
func (ch *ClickHouseAnalytics) GetSessionAnalytics(ctx context.Context, query SessionAnalyticsQuery) ([]SessionAnalyticsEvent, error) {
	sql := `
		SELECT event_id, event_type, session_id, player_id, agent_id, club_id,
			   table_id, device_id, ip_address, country, platform,
			   chips_deposited, chips_withdrawn, net_profit, hands_played,
			   duration_ms, timestamp
		FROM session_analytics
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	if query.PlayerID != "" {
		sql += " AND player_id = ?"
		args = append(args, query.PlayerID)
	}
	if !query.StartTime.IsZero() {
		sql += " AND timestamp >= ?"
		args = append(args, query.StartTime)
	}
	if !query.EndTime.IsZero() {
		sql += " AND timestamp <= ?"
		args = append(args, query.EndTime)
	}

	sql += " ORDER BY timestamp DESC"
	if query.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", query.Limit)
	}

	rows, err := ch.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []SessionAnalyticsEvent
	for rows.Next() {
		var event SessionAnalyticsEvent
		var durationMs int64

		err := rows.Scan(
			&event.EventID, &event.EventType, &event.SessionID, &event.PlayerID,
			&event.AgentID, &event.ClubID, &event.TableID, &event.DeviceID,
			&event.IPAddress, &event.Country, &event.Platform, &event.ChipsDeposited,
			&event.ChipsWithdrawn, &event.NetProfit, &event.HandsPlayed,
			&durationMs, &event.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		event.Duration = time.Duration(durationMs) * time.Millisecond
		events = append(events, event)
	}

	return events, rows.Err()
}

// GetTableAnalytics retrieves table analytics based on query
func (ch *ClickHouseAnalytics) GetTableAnalytics(ctx context.Context, query TableAnalyticsQuery) ([]TableAnalyticsEvent, error) {
	sql := `
		SELECT event_id, event_type, table_id, game_type, betting_type,
			   stake_level, agent_id, club_id, avg_pot_size, avg_hands_per_hour,
			   avg_players_active, total_rake, timestamp, period_start, period_end
		FROM table_stats_analytics
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	if query.TableID != "" {
		sql += " AND table_id = ?"
		args = append(args, query.TableID)
	}
	if !query.StartTime.IsZero() {
		sql += " AND timestamp >= ?"
		args = append(args, query.StartTime)
	}
	if !query.EndTime.IsZero() {
		sql += " AND timestamp <= ?"
		args = append(args, query.EndTime)
	}

	sql += " ORDER BY timestamp DESC"
	if query.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", query.Limit)
	}

	rows, err := ch.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []TableAnalyticsEvent
	for rows.Next() {
		var event TableAnalyticsEvent
		err := rows.Scan(
			&event.EventID, &event.EventType, &event.TableID, &event.GameType,
			&event.BettingType, &event.StakeLevel, &event.AgentID, &event.ClubID,
			&event.AvgPotSize, &event.AvgHandsPerHour, &event.AvgPlayersActive,
			&event.TotalRake, &event.Timestamp, &event.PeriodStart, &event.PeriodEnd,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

// GetPlayerActivityStats retrieves player activity statistics
func (ch *ClickHouseAnalytics) GetPlayerActivityStats(ctx context.Context, query ActivityQuery) ([]PlayerActivityStat, error) {
	sql := `
		SELECT
			player_id,
			count() as hands_played,
			sum(chips_after - chips_before) as total_profit,
			0 as avg_session_time,
			max(timestamp) as last_active
		FROM hand_analytics
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY player_id
		ORDER BY hands_played DESC
		LIMIT ?
	`

	rows, err := ch.db.Query(ctx, sql, query.StartTime, query.EndTime, query.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []PlayerActivityStat
	for rows.Next() {
		var stat PlayerActivityStat
		var avgSessionTimeMs int64
		err := rows.Scan(
			&stat.PlayerID, &stat.HandsPlayed, &stat.TotalProfit,
			&avgSessionTimeMs, &stat.LastActive,
		)
		if err != nil {
			return nil, err
		}
		stat.AvgSessionTime = time.Duration(avgSessionTimeMs) * time.Millisecond
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// Close closes the ClickHouse connection
func (ch *ClickHouseAnalytics) Close() error {
	return ch.db.Close()
}

// Ping checks if the connection is alive
func (ch *ClickHouseAnalytics) Ping(ctx context.Context) error {
	return ch.db.Ping(ctx)
}
