package fraud

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Bot Detection Metrics
	BotDetectionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_bot_detection_duration_seconds",
		Help:    "Time spent detecting bot behavior",
		Buckets: prometheus.DefBuckets,
	}, []string{"detector_type"})

	BotDetectionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "poker_fraud_bot_detection_total",
		Help: "Total number of bot detections",
	}, []string{"detector_type", "result"})

	BotScore = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_bot_score",
		Help:    "Distribution of bot detection scores",
		Buckets: []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
	}, []string{"detector_type"})

	// Collusion Detection Metrics
	CollusionDetectionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_collusion_detection_duration_seconds",
		Help:    "Time spent detecting collusion",
		Buckets: prometheus.DefBuckets,
	}, []string{"detector_type"})

	CollusionDetectionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "poker_fraud_collusion_detection_total",
		Help: "Total number of collusion detections",
	}, []string{"detector_type", "severity"})

	CollusionScore = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_collusion_score",
		Help:    "Distribution of collusion detection scores",
		Buckets: []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
	}, []string{"detector_type"})

	CollusionRingsDetected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_fraud_collusion_rings_detected_total",
		Help: "Total number of collusion rings detected",
	})

	CollusionRingSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "poker_fraud_collusion_ring_size",
		Help:    "Size of detected collusion rings",
		Buckets: []float64{2, 3, 4, 5, 10, 20, 50},
	})

	// Multi-Account Detection Metrics
	MultiAccountDetectionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_multi_account_detection_duration_seconds",
		Help:    "Time spent detecting multi-account abuse",
		Buckets: prometheus.DefBuckets,
	}, []string{"detector_type"})

	MultiAccountDetectionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "poker_fraud_multi_account_detection_total",
		Help: "Total number of multi-account detections",
	}, []string{"detector_type", "severity"})

	MultiAccountScore = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_multi_account_score",
		Help:    "Distribution of multi-account detection scores",
		Buckets: []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
	}, []string{"detector_type"})

	RelatedAccountsFound = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "poker_fraud_related_accounts_found",
		Help:    "Number of related accounts found per player",
		Buckets: []float64{0, 1, 2, 3, 5, 10, 20, 50},
	})

	// Rule Detection Metrics
	RuleDetectionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_rule_detection_duration_seconds",
		Help:    "Time spent on rule-based detection",
		Buckets: prometheus.DefBuckets,
	}, []string{"rule_name"})

	RuleTriggeredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "poker_fraud_rule_triggered_total",
		Help: "Total number of times each rule was triggered",
	}, []string{"rule_name", "category", "severity"})

	RuleCooldownActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "poker_fraud_rule_cooldown_active",
		Help: "Number of rules currently in cooldown",
	}, []string{"rule_name"})

	// Risk Scoring Metrics
	RiskScoreCalculationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "poker_fraud_risk_score_calculation_duration_seconds",
		Help:    "Time spent calculating overall risk scores",
		Buckets: prometheus.DefBuckets,
	})

	RiskScoreOverall = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_risk_score_overall",
		Help:    "Distribution of overall risk scores",
		Buckets: []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
	}, []string{"player_category"})

	RiskScoreBreakdown = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_risk_score_breakdown",
		Help:    "Distribution of risk score components",
		Buckets: []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
	}, []string{"component"})

	HighRiskPlayers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "poker_fraud_high_risk_players",
		Help: "Number of players with high risk scores",
	})

	ReviewRecommended = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_fraud_review_recommended_total",
		Help: "Total number of players recommended for review",
	})

	// Alert Metrics
	AlertGenerated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "poker_fraud_alert_generated_total",
		Help: "Total number of fraud alerts generated",
	}, []string{"alert_type", "severity"})

	AlertProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "poker_fraud_alert_processed_total",
		Help: "Total number of fraud alerts processed",
	}, []string{"alert_type", "action"})

	AlertResolutionTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_alert_resolution_time_seconds",
		Help:    "Time to resolve fraud alerts",
		Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
	}, []string{"alert_type", "severity"})

	AlertStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "poker_fraud_alert_status",
		Help: "Current number of alerts by status",
	}, []string{"status"})

	FalsePositiveRate = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "poker_fraud_false_positive_rate",
		Help:    "Rate of false positive alerts",
		Buckets: []float64{0, 0.01, 0.02, 0.03, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3},
	})

	// Feature Extraction Metrics
	FeatureExtractionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_fraud_feature_extraction_duration_seconds",
		Help:    "Time spent extracting behavioral features",
		Buckets: prometheus.DefBuckets,
	}, []string{"feature_type"})

	FeatureExtractionErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_fraud_feature_extraction_errors_total",
		Help: "Total number of feature extraction errors",
	})

	// Session/Device Metrics
	DeviceFingerprintsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_fraud_device_fingerprints_total",
		Help: "Total number of unique device fingerprints",
	})

	IPAddressesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_fraud_ip_addresses_total",
		Help: "Total number of unique IP addresses",
	})

	SessionOverlapDetected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_fraud_session_overlap_detected_total",
		Help: "Total number of session overlap detections",
	})

	// Performance Metrics
	FraudDetectionTotalDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "poker_fraud_detection_total_duration_seconds",
		Help:    "Total time for complete fraud detection pipeline",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
	})

	FraudDetectionThroughput = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_fraud_detection_throughput_total",
		Help: "Total number of fraud detection runs",
	})

	ActivePlayersMonitored = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "poker_fraud_active_players_monitored",
		Help: "Number of players currently being monitored",
	})

	TablesMonitored = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "poker_fraud_tables_monitored",
		Help: "Number of tables currently being monitored",
	})

	// Error Metrics
	FraudDetectionErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "poker_fraud_detection_errors_total",
		Help: "Total number of fraud detection errors",
	}, []string{"component", "error_type"})
)

// RecordBotDetection records bot detection metrics
func RecordBotDetection(detectorType string, duration float64, score float64, isBot bool) {
	BotDetectionDuration.WithLabelValues(detectorType).Observe(duration)
	result := "human"
	if isBot {
		result = "bot"
	}
	BotDetectionTotal.WithLabelValues(detectorType, result).Inc()
	BotScore.WithLabelValues(detectorType).Observe(score)
}

// RecordCollusionDetection records collusion detection metrics
func RecordCollusionDetection(detectorType string, duration float64, score float64, severity string) {
	CollusionDetectionDuration.WithLabelValues(detectorType).Observe(duration)
	CollusionDetectionTotal.WithLabelValues(detectorType, severity).Inc()
	CollusionScore.WithLabelValues(detectorType).Observe(score)
}

// RecordCollusionRing records metrics when a collusion ring is detected
func RecordCollusionRing(size int) {
	CollusionRingsDetected.Inc()
	CollusionRingSize.Observe(float64(size))
}

// RecordMultiAccountDetection records multi-account detection metrics
func RecordMultiAccountDetection(detectorType string, duration float64, score float64, severity string, relatedAccounts int) {
	MultiAccountDetectionDuration.WithLabelValues(detectorType).Observe(duration)
	MultiAccountDetectionTotal.WithLabelValues(detectorType, severity).Inc()
	MultiAccountScore.WithLabelValues(detectorType).Observe(score)
	RelatedAccountsFound.Observe(float64(relatedAccounts))
}

// RecordRuleTriggered records when a rule is triggered
func RecordRuleTriggered(ruleName, category, severity string) {
	RuleTriggeredTotal.WithLabelValues(ruleName, category, severity).Inc()
}

// RecordRiskScore records risk score metrics
func RecordRiskScore(overall float64, breakdown map[string]float64, category string) {
	RiskScoreCalculationDuration.Observe(0) // Call this before calculation
	RiskScoreOverall.WithLabelValues(category).Observe(overall)
	for component, score := range breakdown {
		RiskScoreBreakdown.WithLabelValues(component).Observe(score)
	}
	if overall >= 0.8 {
		HighRiskPlayers.Inc()
	}
	if overall >= 0.6 {
		ReviewRecommended.Inc()
	}
}

// RecordAlert records alert metrics
func RecordAlert(alertType, severity string, resolutionTime float64, action string) {
	AlertGenerated.WithLabelValues(alertType, severity).Inc()
	AlertProcessed.WithLabelValues(alertType, action).Inc()
	AlertResolutionTime.WithLabelValues(alertType, severity).Observe(resolutionTime)
}

// UpdateAlertStatus updates the current alert status gauge
func UpdateAlertStatus(status string, count float64) {
	AlertStatus.WithLabelValues(status).Set(count)
}

// RecordFalsePositiveRate records the false positive rate
func RecordFalsePositiveRate(rate float64) {
	FalsePositiveRate.Observe(rate)
}

// RecordFeatureExtraction records feature extraction metrics
func RecordFeatureExtraction(featureType string, duration float64, success bool) {
	FeatureExtractionDuration.WithLabelValues(featureType).Observe(duration)
	if !success {
		FeatureExtractionErrors.Inc()
	}
}

// RecordFraudDetection records overall fraud detection metrics
func RecordFraudDetection(duration float64) {
	FraudDetectionTotalDuration.Observe(duration)
	FraudDetectionThroughput.Inc()
}

// RecordError records error metrics
func RecordError(component, errorType string) {
	FraudDetectionErrors.WithLabelValues(component, errorType).Inc()
}
