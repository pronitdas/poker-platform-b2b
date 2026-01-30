package fraud

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// KafkaAlertProducerConfig holds Kafka producer configuration
type KafkaAlertProducerConfig struct {
	Brokers        []string
	Topic          string
	ProducerGroup  string
	MaxRetries     int
	RetryBackoff   time.Duration
	FlushFrequency time.Duration
	FlushMessages  int
	RequiredAcks   sarama.RequiredAcks
	Compression    sarama.CompressionCodec
	BatchSize      int
	BatchTimeout   time.Duration
	AsyncMode      bool
}

// KafkaAlertProducer publishes fraud alerts to Kafka
type KafkaAlertProducer struct {
	producer sarama.SyncProducer
	async    sarama.AsyncProducer
	topic    string
	mu       sync.RWMutex
	closed   bool
	stats    *ProducerStats
}

// ProducerStats tracks Kafka producer statistics
type ProducerStats struct {
	MessagesSent     int64
	MessagesFailed   int64
	BytesSent        int64
	LastMessageTime  time.Time
	Errors           []ProducerError
	lastErrorCleanup time.Time
}

type ProducerError struct {
	Time    time.Time
	Error   error
	Message *AlertMessage
}

// AlertMessage is the message format for Kafka
type AlertMessage struct {
	ID            string             `json:"id"`
	PlayerID      string             `json:"player_id"`
	AlertType     string             `json:"alert_type"`
	Severity      string             `json:"severity"`
	Score         float64            `json:"score"`
	TableID       string             `json:"table_id,omitempty"`
	HandID        string             `json:"hand_id,omitempty"`
	AgentID       string             `json:"agent_id"`
	ClubID        string             `json:"club_id"`
	Evidence      []string           `json:"evidence"`
	Metadata      json.RawMessage    `json:"metadata,omitempty"`
	Timestamp     time.Time          `json:"timestamp"`
	DetectedAt    time.Time          `json:"detected_at"`
	RiskBreakdown map[string]float64 `json:"risk_breakdown,omitempty"`
}

// NewKafkaAlertProducer creates a new Kafka alert producer
func NewKafkaAlertProducer(config KafkaAlertProducerConfig) (*KafkaAlertProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Retry.Max = config.MaxRetries
	saramaConfig.Producer.Retry.Backoff = config.RetryBackoff
	saramaConfig.Producer.Flush.Frequency = config.FlushFrequency
	saramaConfig.Producer.Flush.Messages = config.FlushMessages
	saramaConfig.Producer.RequiredAcks = config.RequiredAcks
	saramaConfig.Producer.Compression = config.Compression
	saramaConfig.Producer.Flush.MaxMessages = config.BatchSize

	// Enable idempotent producer for exactly-once semantics
	if config.RequiredAcks == sarama.WaitForAll {
		saramaConfig.Producer.Idempotent = true
		saramaConfig.Net.MaxOpenRequests = 1
	}

	var producer sarama.SyncProducer
	var async sarama.AsyncProducer
	var err error

	if config.AsyncMode {
		async, err = sarama.NewAsyncProducer(config.Brokers, saramaConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create async Kafka producer: %w", err)
		}
	} else {
		producer, err = sarama.NewSyncProducer(config.Brokers, saramaConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create sync Kafka producer: %w", err)
		}
	}

	p := &KafkaAlertProducer{
		producer: producer,
		async:    async,
		topic:    config.Topic,
		stats:    &ProducerStats{},
	}

	// Start error handler for async producer
	if async != nil {
		go p.handleErrors()
	}

	return p, nil
}

// handleErrors processes errors from async producer
func (p *KafkaAlertProducer) handleErrors() {
	for err := range p.async.Errors() {
		p.mu.Lock()
		p.stats.MessagesFailed++
		p.stats.Errors = append(p.stats.Errors, ProducerError{
			Time:  time.Now(),
			Error: err,
		})
		// Clean up old errors periodically
		if time.Since(p.stats.lastErrorCleanup) > 5*time.Minute {
			if len(p.stats.Errors) > 100 {
				p.stats.Errors = p.stats.Errors[len(p.stats.Errors)-100:]
			}
			p.stats.lastErrorCleanup = time.Now()
		}
		p.mu.Unlock()
	}
}

// PublishAlert sends an alert to Kafka synchronously
func (p *KafkaAlertProducer) PublishAlert(ctx context.Context, alert *AntiCheatAlert, riskBreakdown map[string]float64) error {
	if p.producer == nil {
		return fmt.Errorf("producer is not configured for sync mode")
	}

	msg := p.buildMessage(alert, riskBreakdown)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(alert.PlayerID),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{Key: []byte("alert_type"), Value: []byte(alert.AlertType)},
			{Key: []byte("severity"), Value: []byte(alert.Severity)},
			{Key: []byte("agent_id"), Value: []byte(alert.AgentID)},
			{Key: []byte("club_id"), Value: []byte(alert.ClubID)},
		},
		Timestamp: time.Now(),
	}

	partition, offset, err := p.producer.SendMessage(kafkaMsg)
	if err != nil {
		p.mu.Lock()
		p.stats.MessagesFailed++
		p.stats.Errors = append(p.stats.Errors, ProducerError{
			Time:    time.Now(),
			Error:   err,
			Message: &msg,
		})
		p.mu.Unlock()
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	p.mu.Lock()
	p.stats.MessagesSent++
	p.stats.BytesSent += int64(len(data))
	p.stats.LastMessageTime = time.Now()
	p.mu.Unlock()

	// Note: In production, you might want to log partition/offset for debugging
	_ = partition
	_ = offset

	return nil
}

// PublishAlertAsync sends an alert to Kafka asynchronously
func (p *KafkaAlertProducer) PublishAlertAsync(alert *AntiCheatAlert, riskBreakdown map[string]float64) error {
	if p.async == nil {
		return fmt.Errorf("producer is not configured for async mode")
	}

	msg := p.buildMessage(alert, riskBreakdown)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(alert.PlayerID),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{Key: []byte("alert_type"), Value: []byte(alert.AlertType)},
			{Key: []byte("severity"), Value: []byte(alert.Severity)},
			{Key: []byte("agent_id"), Value: []byte(alert.AgentID)},
			{Key: []byte("club_id"), Value: []byte(alert.ClubID)},
		},
		Timestamp: time.Now(),
	}

	p.async.Input() <- kafkaMsg

	p.mu.Lock()
	p.stats.MessagesSent++
	p.stats.BytesSent += int64(len(data))
	p.stats.LastMessageTime = time.Now()
	p.mu.Unlock()

	return nil
}

// buildMessage constructs an AlertMessage from an AntiCheatAlert
func (p *KafkaAlertProducer) buildMessage(alert *AntiCheatAlert, riskBreakdown map[string]float64) AlertMessage {
	return AlertMessage{
		ID:            alert.ID,
		PlayerID:      alert.PlayerID,
		AlertType:     alert.AlertType,
		Severity:      alert.Severity,
		Score:         alert.Score,
		TableID:       alert.TableID,
		HandID:        alert.HandID,
		AgentID:       alert.AgentID,
		ClubID:        alert.ClubID,
		Evidence:      alert.Evidence,
		Metadata:      alert.Metadata,
		Timestamp:     alert.CreatedAt,
		DetectedAt:    time.Now(),
		RiskBreakdown: riskBreakdown,
	}
}

// PublishBatch sends multiple alerts to Kafka
func (p *KafkaAlertProducer) PublishBatch(ctx context.Context, alerts []*AntiCheatAlert, riskBreakdowns []map[string]float64) error {
	if len(alerts) == 0 {
		return nil
	}

	if p.producer == nil {
		return fmt.Errorf("producer is not configured for sync mode")
	}

	for i, alert := range alerts {
		var breakdown map[string]float64
		if i < len(riskBreakdowns) {
			breakdown = riskBreakdowns[i]
		}

		if err := p.PublishAlert(ctx, alert, breakdown); err != nil {
			return fmt.Errorf("failed to publish alert %d: %w", i, err)
		}
	}

	return nil
}

// GetStats returns current producer statistics
func (p *KafkaAlertProducer) GetStats() ProducerStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return *p.stats
}

// GetErrors returns recent producer errors
func (p *KafkaAlertProducer) GetErrors() []ProducerError {
	p.mu.RLock()
	defer p.mu.RUnlock()

	recentErrors := make([]ProducerError, 0, len(p.stats.Errors))
	cutoff := time.Now().Add(-1 * time.Hour)
	for _, err := range p.stats.Errors {
		if err.Time.After(cutoff) {
			recentErrors = append(recentErrors, err)
		}
	}
	return recentErrors
}

// Close shuts down the producer gracefully
func (p *KafkaAlertProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true

	var err error
	if p.producer != nil {
		err = p.producer.Close()
	}
	if p.async != nil {
		asyncErr := p.async.Close()
		if err == nil {
			err = asyncErr
		}
	}

	return err
}

// EnsureTopic creates the topic if it doesn't exist
func EnsureTopic(brokers []string, topic string, partitions int32, replicationFactor int16) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0

	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return fmt.Errorf("failed to create cluster admin: %w", err)
	}
	defer admin.Close()

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	err = admin.CreateTopic(topic, topicDetail, false)
	if err != nil {
		// Ignore error if topic already exists
		if topicErr, ok := err.(*sarama.TopicError); ok && topicErr.Err == sarama.ErrTopicAlreadyExists {
			return nil
		}
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}
