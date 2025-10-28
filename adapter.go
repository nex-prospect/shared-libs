package observability

import (
	"context"

	"github.com/gomessguii/logger"
)

// LoggerAdapter adapta o logger existente para interfaces que precisam de campos estruturados
type LoggerAdapter struct {
	logger *logger.Logger
}

func NewLoggerAdapter(l *logger.Logger) *LoggerAdapter {
	return &LoggerAdapter{logger: l}
}

func (a *LoggerAdapter) LogInfo(ctx context.Context, message string, fields map[string]interface{}) {
	// Converter fields para string formatado
	a.logger.LogInfo("%s | fields=%v", message, fields)
}

func (a *LoggerAdapter) LogWarn(ctx context.Context, message string, fields map[string]interface{}) {
	a.logger.LogWarn("%s | fields=%v", message, fields)
}

func (a *LoggerAdapter) LogError(ctx context.Context, message string, fields map[string]interface{}) {
	a.logger.LogError("%s | fields=%v", message, fields)
}
