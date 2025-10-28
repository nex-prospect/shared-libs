package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// StorageConfig configura onde os snapshots serão persistidos
type StorageConfig struct {
	Enabled   bool   // Se false, apenas logs estruturados (STDOUT)
	Type      string // "local", "s3", "postgres", "disabled"
	LocalPath string // Caminho para storage local (padrão: /tmp/scheduler_sessions)
	S3Bucket  string // Bucket S3 para snapshots
	S3Region  string // Região AWS
}

// DefaultStorageConfig retorna config padrão (local, habilitado)
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Enabled:   true,
		Type:      "local",
		LocalPath: "/tmp/scheduler_sessions",
	}
}

// DisabledStorageConfig retorna config com snapshots desabilitados (apenas logs)
func DisabledStorageConfig() StorageConfig {
	return StorageConfig{
		Enabled: false,
		Type:    "disabled",
	}
}

// SchedulerTracer rastreia CADA decisão do agendamento para debug profundo
type SchedulerTracer struct {
	logger Logger
	config StorageConfig
}

// NewSchedulerTracer cria tracer com config padrão (local storage)
func NewSchedulerTracer(logger Logger) *SchedulerTracer {
	return &SchedulerTracer{
		logger: logger,
		config: DefaultStorageConfig(),
	}
}

// NewSchedulerTracerWithConfig cria tracer com config customizada
func NewSchedulerTracerWithConfig(logger Logger, config StorageConfig) *SchedulerTracer {
	return &SchedulerTracer{
		logger: logger,
		config: config,
	}
}

// SchedulingSession representa UMA execução de ProcessActiveCadences
type SchedulingSession struct {
	SessionID   string    `json:"session_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CadenceID   uuid.UUID `json:"cadence_id"`
	CadenceName string    `json:"cadence_name"`
	TenantID    uuid.UUID `json:"tenant_id"`

	// Inputs
	Inputs SchedulingInputs `json:"inputs"`

	// Decisões tomadas
	Decisions []SchedulingDecision `json:"decisions"`

	// Resultados
	Results SchedulingResults `json:"results"`

	// Problemas detectados
	Issues []SchedulingIssue `json:"issues"`

	// Metadata para correlação
	Metadata map[string]interface{} `json:"metadata"`
}

type SchedulingInputs struct {
	CadenceConfig     CadenceConfigSnapshot   `json:"cadence_config"`
	ChannelConfigs    []ChannelConfigSnapshot `json:"channel_configs"`
	LeadsAvailable    int                     `json:"leads_available"`
	LeadsFiltered     []LeadFilterResult      `json:"leads_filtered"`
	LastScheduledTime *time.Time              `json:"last_scheduled_time"`
	DeliveryWindow    DeliveryWindowSnapshot  `json:"delivery_window"`
	RateLimit         RateLimitSnapshot       `json:"rate_limit"`
	CurrentTime       time.Time               `json:"current_time"`
}

type CadenceConfigSnapshot struct {
	ID                uuid.UUID   `json:"id"`
	Name              string      `json:"name"`
	Status            string      `json:"status"`
	RequiredTagIDs    []uuid.UUID `json:"required_tag_ids"`
	RequiredStatusIDs []uuid.UUID `json:"required_status_ids"`
	StepCount         int         `json:"step_count"`
}

type ChannelConfigSnapshot struct {
	ID                     uuid.UUID  `json:"id"`
	Name                   string     `json:"name"`
	Status                 string     `json:"status"`
	CadenceIntervalSeconds int        `json:"cadence_interval_seconds"`
	LastCadenceTime        *time.Time `json:"last_cadence_time"`
}

type LeadFilterResult struct {
	LeadID     uuid.UUID   `json:"lead_id"`
	Included   bool        `json:"included"`
	Reason     string      `json:"reason"`
	Tags       []uuid.UUID `json:"tags"`
	Status     uuid.UUID   `json:"status"`
	PhoneValid *bool       `json:"phone_valid,omitempty"`
}

type DeliveryWindowSnapshot struct {
	Enabled    bool                        `json:"enabled"`
	Windows    map[string]WindowDefinition `json:"windows"`
	TimeZone   string                      `json:"timezone"`
	NextWindow *NextWindowInfo             `json:"next_window,omitempty"`
}

type WindowDefinition struct {
	Weekday   string `json:"weekday"`
	StartHour int    `json:"start_hour"`
	EndHour   int    `json:"end_hour"`
}

type NextWindowInfo struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

type RateLimitSnapshot struct {
	Enabled        bool   `json:"enabled"`
	LimitType      string `json:"limit_type"`
	MaxLeads       int    `json:"max_leads"`
	CurrentPeriod  string `json:"current_period"`
	LeadsProcessed int    `json:"leads_processed"`
	RemainingSlots int    `json:"remaining_slots"`
}

// SchedulingDecision captura UMA decisão de agendamento
type SchedulingDecision struct {
	DecisionID   string    `json:"decision_id"`
	Timestamp    time.Time `json:"timestamp"`
	DecisionType string    `json:"decision_type"`
	LeadID       uuid.UUID `json:"lead_id,omitempty"`

	// Cálculos realizados
	Calculation CalculationDetails `json:"calculation"`

	// Ajustes aplicados
	Adjustments []Adjustment `json:"adjustments"`

	// Resultado final
	FinalScheduleTime time.Time `json:"final_schedule_time"`

	// Validações
	Validations []ValidationResult `json:"validations"`
}

type CalculationDetails struct {
	// Step 1: Intervalo efetivo
	ChannelMinInterval   int    `json:"channel_min_interval_seconds"`
	ProportionalInterval *int   `json:"proportional_interval_seconds,omitempty"`
	EffectiveInterval    int    `json:"effective_interval_seconds"`
	IntervalReason       string `json:"interval_reason"`

	// Step 2: Tempo base
	BaseTime            time.Time `json:"base_time"`
	BaseTimeCalculation string    `json:"base_time_calculation"`
	BaseTimeWasPast     bool      `json:"base_time_was_past"`

	// Step 3: Position offset
	LeadIndexInPeriod int `json:"lead_index_in_period"`
	PeriodIndex       int `json:"period_index"`
	PositionOffset    int `json:"position_offset_seconds"`

	// Step 4: Initial calculated time
	InitialCalculatedTime time.Time `json:"initial_calculated_time"`
}

type Adjustment struct {
	Type     string                 `json:"type"`
	Reason   string                 `json:"reason"`
	Before   time.Time              `json:"before"`
	After    time.Time              `json:"after"`
	Delta    int                    `json:"delta_seconds"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ValidationResult struct {
	CheckType string    `json:"check_type"`
	Passed    bool      `json:"passed"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type SchedulingResults struct {
	TotalLeadsProcessed     int `json:"total_leads_processed"`
	LeadsScheduledToday     int `json:"leads_scheduled_today"`
	LeadsDeferredNextPeriod int `json:"leads_deferred_next_period"`
	LeadsSkipped            int `json:"leads_skipped"`
	LeadsFailed             int `json:"leads_failed"`

	// Distribuição temporal
	Distribution TemporalDistribution `json:"distribution"`

	// Performance
	ExecutionTimeMs int64 `json:"execution_time_ms"`
	DatabaseQueries int   `json:"database_queries"`

	// Scheduling quality metrics
	QualityMetrics QualityMetrics `json:"quality_metrics"`
}

type TemporalDistribution struct {
	EarliestSchedule time.Time      `json:"earliest_schedule"`
	LatestSchedule   time.Time      `json:"latest_schedule"`
	TotalSpan        int            `json:"total_span_seconds"`
	AverageInterval  float64        `json:"average_interval_seconds"`
	MinInterval      int            `json:"min_interval_seconds"`
	MaxInterval      int            `json:"max_interval_seconds"`
	PeriodsUsed      int            `json:"periods_used"`
	LeadsPerPeriod   map[string]int `json:"leads_per_period"`
}

type QualityMetrics struct {
	// Detecção de problemas
	TimeCollapseDetected bool `json:"time_collapse_detected"`
	CollapseCount        int  `json:"collapse_count"`

	// Conformidade
	MinIntervalViolations int `json:"min_interval_violations"`
	WindowViolations      int `json:"window_violations"`

	// Distribuição
	DistributionScore    float64 `json:"distribution_score"`    // 0-100
	ProportionalityScore float64 `json:"proportionality_score"` // 0-100
	RandomnessScore      float64 `json:"randomness_score"`      // 0-100
}

type SchedulingIssue struct {
	Severity  string                 `json:"severity"` // ERROR, WARNING, INFO
	IssueType string                 `json:"issue_type"`
	Message   string                 `json:"message"`
	LeadID    *uuid.UUID             `json:"lead_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

// StartSession inicia uma sessão de rastreamento
func (st *SchedulerTracer) StartSession(ctx context.Context, cadenceID uuid.UUID, cadenceName string, tenantID uuid.UUID) *SchedulingSession {
	session := &SchedulingSession{
		SessionID:   uuid.New().String(),
		StartTime:   time.Now(),
		CadenceID:   cadenceID,
		CadenceName: cadenceName,
		TenantID:    tenantID,
		Decisions:   make([]SchedulingDecision, 0),
		Issues:      make([]SchedulingIssue, 0),
		Metadata:    make(map[string]interface{}),
	}

	st.logger.LogInfo(ctx, "scheduling_session_started", map[string]interface{}{
		"session_id":   session.SessionID,
		"cadence_id":   cadenceID,
		"cadence_name": cadenceName,
		"tenant_id":    tenantID,
	})

	return session
}

// RecordInputs captura estado inicial
func (st *SchedulerTracer) RecordInputs(ctx context.Context, session *SchedulingSession, inputs SchedulingInputs) {
	session.Inputs = inputs

	st.logger.LogInfo(ctx, "scheduling_inputs_recorded", map[string]interface{}{
		"session_id":         session.SessionID,
		"leads_available":    inputs.LeadsAvailable,
		"channels":           len(inputs.ChannelConfigs),
		"rate_limit_enabled": inputs.RateLimit.Enabled,
		"last_scheduled":     inputs.LastScheduledTime,
	})
}

// RecordDecision captura uma decisão de agendamento
func (st *SchedulerTracer) RecordDecision(ctx context.Context, session *SchedulingSession, decision SchedulingDecision) {
	decision.DecisionID = fmt.Sprintf("%s-%d", session.SessionID, len(session.Decisions))
	decision.Timestamp = time.Now()

	session.Decisions = append(session.Decisions, decision)

	st.logger.LogInfo(ctx, "scheduling_decision", map[string]interface{}{
		"session_id":          session.SessionID,
		"decision_id":         decision.DecisionID,
		"lead_id":             decision.LeadID,
		"decision_type":       decision.DecisionType,
		"final_schedule_time": decision.FinalScheduleTime,
		"effective_interval":  decision.Calculation.EffectiveInterval,
		"adjustments_count":   len(decision.Adjustments),
	})
}

// RecordIssue registra um problema detectado
func (st *SchedulerTracer) RecordIssue(ctx context.Context, session *SchedulingSession, severity, issueType, message string, leadID *uuid.UUID, context map[string]interface{}) {
	issue := SchedulingIssue{
		Severity:  severity,
		IssueType: issueType,
		Message:   message,
		LeadID:    leadID,
		Timestamp: time.Now(),
		Context:   context,
	}

	session.Issues = append(session.Issues, issue)

	st.logger.LogWarn(ctx, "scheduling_issue_detected", map[string]interface{}{
		"session_id": session.SessionID,
		"severity":   severity,
		"issue_type": issueType,
		"message":    message,
		"lead_id":    leadID,
		"context":    context,
	})
}

// CompleteSession finaliza sessão e gera relatório completo
func (st *SchedulerTracer) CompleteSession(ctx context.Context, session *SchedulingSession, results SchedulingResults) {
	session.EndTime = time.Now()
	session.Results = results
	session.Results.ExecutionTimeMs = session.EndTime.Sub(session.StartTime).Milliseconds()

	// Calcular quality metrics
	session.Results.QualityMetrics = st.calculateQualityMetrics(session)

	// Log resumo estruturado
	st.logger.LogInfo(ctx, "scheduling_session_completed", map[string]interface{}{
		"session_id":             session.SessionID,
		"execution_time_ms":      session.Results.ExecutionTimeMs,
		"leads_scheduled":        session.Results.LeadsScheduledToday,
		"leads_deferred":         session.Results.LeadsDeferredNextPeriod,
		"issues_count":           len(session.Issues),
		"time_collapse_detected": session.Results.QualityMetrics.TimeCollapseDetected,
		"distribution_score":     session.Results.QualityMetrics.DistributionScore,
	})

	// Persistir snapshot completo em JSON (para análise posterior)
	st.persistSessionSnapshot(ctx, session)

	// Alertar se problemas críticos detectados
	if session.Results.QualityMetrics.TimeCollapseDetected {
		st.logger.LogError(ctx, "CRITICAL: Time collapse detected in scheduling session", map[string]interface{}{
			"session_id":     session.SessionID,
			"collapse_count": session.Results.QualityMetrics.CollapseCount,
		})
	}
}

// calculateQualityMetrics analisa qualidade do agendamento
func (st *SchedulerTracer) calculateQualityMetrics(session *SchedulingSession) QualityMetrics {
	metrics := QualityMetrics{}

	if len(session.Decisions) == 0 {
		return metrics
	}

	// Detectar collapses (múltiplos leads no mesmo timestamp)
	timestampCounts := make(map[time.Time]int)
	for _, decision := range session.Decisions {
		timestampCounts[decision.FinalScheduleTime]++
	}

	collapseCount := 0
	for _, count := range timestampCounts {
		if count > 1 {
			collapseCount++
		}
	}

	metrics.TimeCollapseDetected = collapseCount > 0
	metrics.CollapseCount = collapseCount

	// Validar intervalos mínimos
	sortedDecisions := make([]SchedulingDecision, len(session.Decisions))
	copy(sortedDecisions, session.Decisions)
	// Sort by FinalScheduleTime (simplificado - assumindo já ordenado)

	for i := 1; i < len(sortedDecisions); i++ {
		diff := sortedDecisions[i].FinalScheduleTime.Sub(sortedDecisions[i-1].FinalScheduleTime)
		if diff.Seconds() < 60 {
			metrics.MinIntervalViolations++
		}
	}

	// Score de distribuição (0-100)
	// Ideal: leads uniformemente distribuídos
	if session.Results.Distribution.AverageInterval > 0 {
		// Calcular variância dos intervalos
		// Score alto = baixa variância (distribuição uniforme)
		metrics.DistributionScore = 100.0 - float64(metrics.MinIntervalViolations)*10
		if metrics.DistributionScore < 0 {
			metrics.DistributionScore = 0
		}
	}

	// Score de proporcionalidade (0-100)
	// Se rate limit ativo, verificar se intervalos estão próximos do proporcional esperado
	if session.Inputs.RateLimit.Enabled {
		// Implementar cálculo de desvio do intervalo proporcional ideal
		metrics.ProportionalityScore = 85.0 // Placeholder
	}

	// Score de aleatoriedade (0-100)
	// Verificar se há jitter aplicado
	hasJitter := false
	for _, decision := range session.Decisions {
		for _, adj := range decision.Adjustments {
			if adj.Type == "jitter" {
				hasJitter = true
				break
			}
		}
		if hasJitter {
			break
		}
	}
	if hasJitter {
		metrics.RandomnessScore = 90.0
	} else {
		metrics.RandomnessScore = 0.0
	}

	return metrics
}

// persistSessionSnapshot salva snapshot completo baseado na configuração
func (st *SchedulerTracer) persistSessionSnapshot(ctx context.Context, session *SchedulingSession) {
	// SEMPRE logar resumo estruturado no STDOUT (coletado por sistema de logs)
	st.logStructuredSummary(ctx, session)

	// Snapshot completo: apenas se habilitado
	if !st.config.Enabled {
		return
	}

	jsonData, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		st.logger.LogError(ctx, "Failed to marshal session snapshot", map[string]interface{}{
			"session_id": session.SessionID,
			"error":      err.Error(),
		})
		return
	}

	// Persistir baseado no tipo de storage configurado
	switch st.config.Type {
	case "local":
		st.persistToLocal(ctx, session.SessionID, jsonData)
	case "s3":
		st.persistToS3(ctx, session.SessionID, jsonData)
	case "postgres":
		st.persistToPostgres(ctx, session.SessionID, jsonData)
	default:
		// disabled ou tipo desconhecido: apenas logs estruturados
	}
}

// logStructuredSummary SEMPRE executado: log estruturado no STDOUT
func (st *SchedulerTracer) logStructuredSummary(ctx context.Context, session *SchedulingSession) {
	st.logger.LogInfo(ctx, "scheduling_session_summary", map[string]interface{}{
		"session_id":             session.SessionID,
		"cadence_id":             session.CadenceID,
		"cadence_name":           session.CadenceName,
		"tenant_id":              session.TenantID,
		"leads_scheduled":        session.Results.LeadsScheduledToday,
		"leads_deferred":         session.Results.LeadsDeferredNextPeriod,
		"leads_skipped":          session.Results.LeadsSkipped,
		"time_collapse_detected": session.Results.QualityMetrics.TimeCollapseDetected,
		"collapse_count":         session.Results.QualityMetrics.CollapseCount,
		"execution_time_ms":      session.Results.ExecutionTimeMs,
		"issues_count":           len(session.Issues),
		"distribution_score":     session.Results.QualityMetrics.DistributionScore,
	})
}

// persistToLocal salva snapshot em arquivo local
func (st *SchedulerTracer) persistToLocal(ctx context.Context, sessionID string, jsonData []byte) {
	// Criar diretório se não existir
	if st.config.LocalPath != "" {
		if err := os.MkdirAll(st.config.LocalPath, 0755); err != nil {
			st.logger.LogError(ctx, "Failed to create snapshot directory", map[string]interface{}{
				"path":  st.config.LocalPath,
				"error": err.Error(),
			})
			return
		}
	}

	// Determinar caminho do arquivo
	path := st.config.LocalPath
	if path == "" {
		path = "/tmp/scheduler_sessions"
	}
	filename := filepath.Join(path, fmt.Sprintf("session_%s.json", sessionID))

	// Escrever arquivo
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		st.logger.LogError(ctx, "Failed to write session snapshot", map[string]interface{}{
			"filename": filename,
			"error":    err.Error(),
		})
		return
	}

	st.logger.LogInfo(ctx, "session_snapshot_persisted", map[string]interface{}{
		"session_id":   sessionID,
		"storage_type": "local",
		"filename":     filename,
		"size_bytes":   len(jsonData),
	})
}

// persistToS3 salva snapshot no S3 (implementação básica)
func (st *SchedulerTracer) persistToS3(ctx context.Context, sessionID string, jsonData []byte) {
	// TODO: Implementar upload para S3
	// Requer: AWS SDK, credenciais, etc.
	// Por enquanto, apenas logar que seria salvo no S3

	st.logger.LogInfo(ctx, "session_snapshot_would_be_persisted_to_s3", map[string]interface{}{
		"session_id": sessionID,
		"bucket":     st.config.S3Bucket,
		"region":     st.config.S3Region,
		"size_bytes": len(jsonData),
		"note":       "S3 implementation pending - install AWS SDK and configure credentials",
	})

	// Implementação real requer:
	// import "github.com/aws/aws-sdk-go/service/s3"
	//
	// sess, _ := session.NewSession(&aws.Config{
	//     Region: aws.String(st.config.S3Region),
	// })
	// svc := s3.New(sess)
	// key := fmt.Sprintf("scheduler-sessions/%s/session_%s.json",
	//                    time.Now().Format("2006-01-02"), sessionID)
	// _, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
	//     Bucket: aws.String(st.config.S3Bucket),
	//     Key:    aws.String(key),
	//     Body:   bytes.NewReader(jsonData),
	// })
}

// persistToPostgres salva snapshot no PostgreSQL
func (st *SchedulerTracer) persistToPostgres(ctx context.Context, sessionID string, jsonData []byte) {
	// TODO: Implementar INSERT no PostgreSQL
	// Requer: gorm ou pgx, tabela scheduler_sessions

	st.logger.LogInfo(ctx, "session_snapshot_would_be_persisted_to_postgres", map[string]interface{}{
		"session_id": sessionID,
		"size_bytes": len(jsonData),
		"note":       "PostgreSQL implementation pending - create table and add insert logic",
	})

	// Implementação real requer:
	// CREATE TABLE scheduler_sessions (
	//     id UUID PRIMARY KEY,
	//     session_id VARCHAR(255) UNIQUE,
	//     cadence_id UUID,
	//     tenant_id UUID,
	//     created_at TIMESTAMP,
	//     snapshot JSONB
	// );
	//
	// INSERT INTO scheduler_sessions (id, session_id, snapshot, created_at)
	// VALUES ($1, $2, $3, NOW())
}

// Logger interface (simplificado)
type Logger interface {
	LogInfo(ctx context.Context, message string, fields map[string]interface{})
	LogWarn(ctx context.Context, message string, fields map[string]interface{})
	LogError(ctx context.Context, message string, fields map[string]interface{})
}
