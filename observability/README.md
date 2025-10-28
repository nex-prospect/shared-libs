# Observability Library

Biblioteca compartilhada de observabilidade para rastreamento e logging estruturado em microserviços.

## Componentes

### 1. LoggerAdapter
Adapta o `github.com/gomessguii/logger` para interfaces que exigem logs estruturados com campos.

### 2. SchedulerTracer
Sistema de rastreamento para operações de agendamento (scheduling), permitindo:
- Tracking de sessões de scheduling
- Snapshots de decisões de agendamento
- Métricas de qualidade (time collapse, distribuição, etc.)
- Persistência em local storage, S3, ou PostgreSQL

## Uso

### Adicionar ao seu serviço

```bash
cd /path/to/your-service
go mod edit -replace github.com/nex-prospect/shared-libs/observability=/home/is_me/go/src/github.com/nex-prospect/shared-libs/observability
go get github.com/nex-prospect/shared-libs/observability
```

### Exemplo: LoggerAdapter

```go
import (
    "github.com/gomessguii/logger"
    "github.com/nex-prospect/shared-libs/observability"
)

func main() {
    myLogger := logger.New()
    adapter := observability.NewLoggerAdapter(myLogger)

    adapter.LogInfo(ctx, "Processing started", map[string]interface{}{
        "user_id": "123",
        "action": "create",
    })
}
```

### Exemplo: SchedulerTracer

```go
import "github.com/nex-prospect/shared-libs/observability"

func main() {
    // Configurar storage (local, S3, ou desabilitado)
    storageConfig := observability.StorageConfig{
        Enabled:   true,
        Type:      "local",
        LocalPath: "/tmp/scheduler_sessions",
    }

    // Criar tracer
    loggerAdapter := observability.NewLoggerAdapter(myLogger)
    tracer := observability.NewSchedulerTracerWithConfig(loggerAdapter, storageConfig)

    // Iniciar sessão de rastreamento
    session := tracer.StartSession(ctx, cadenceID, "My Cadence", tenantID)

    // ... realizar operações de scheduling ...

    // Registrar decisões
    tracer.RecordDecision(ctx, session, observability.SchedulingDecision{
        DecisionType:      "schedule_lead",
        LeadID:            leadID,
        FinalScheduleTime: scheduleTime,
    })

    // Completar sessão com resultados
    tracer.CompleteSession(ctx, session, observability.SchedulingResults{
        TotalLeadsProcessed: 100,
        LeadsScheduledToday: 80,
    })
}
```

## Interfaces

### Logger Interface
```go
type Logger interface {
    LogInfo(ctx context.Context, message string, fields map[string]interface{})
    LogWarn(ctx context.Context, message string, fields map[string]interface{})
    LogError(ctx context.Context, message string, fields map[string]interface{})
}
```

## Serviços que usam

- `campaign-service` - Rastreamento de scheduling de cadências

## Estrutura de arquivos

```
observability/
├── adapter.go           # LoggerAdapter
├── scheduler_tracer.go  # SchedulerTracer + modelos
├── go.mod              # Dependências
└── README.md           # Documentação
```

## Por que está em shared-libs?

Para permitir reutilização entre múltiplos microserviços sem criar dependências entre eles. Qualquer serviço que precise de observabilidade estruturada pode importar esta biblioteca.

## Próximos passos

Quando pronto para publicar no GitHub:

1. Criar repositório público `github.com/nex-prospect/shared-libs`
2. Fazer push do código
3. Remover o `replace` do go.mod dos serviços
4. Usar `go get github.com/nex-prospect/shared-libs/observability@latest`
