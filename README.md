# Observability Library

Go library for structured logging and operation tracing in microservices.

## Components

### 1. LoggerAdapter
Adapts `github.com/gomessguii/logger` to interfaces that require structured logging with field maps.

### 2. SchedulerTracer
Tracing system for scheduling operations, providing:
- Session tracking
- Decision snapshots
- Quality metrics (time distribution, collision detection, etc.)
- Persistence options: local storage, S3, or PostgreSQL

## Installation

```bash
go get github.com/nex-prospect/shared-libs/observability
```

## Usage

### LoggerAdapter

Provides a structured logging interface with context and field support:

```go
import (
    "context"
    "github.com/gomessguii/logger"
    "github.com/nex-prospect/shared-libs/observability"
)

func main() {
    log := logger.New()
    adapter := observability.NewLoggerAdapter(log)

    ctx := context.Background()
    adapter.LogInfo(ctx, "Operation completed", map[string]interface{}{
        "duration_ms": 150,
        "status": "success",
    })
}
```

### SchedulerTracer

Tracks scheduling operations with detailed metrics and snapshots:

```go
import "github.com/nex-prospect/shared-libs/observability"

func main() {
    // Configure storage
    config := observability.StorageConfig{
        Enabled:   true,
        Type:      "local",
        LocalPath: "/tmp/sessions",
    }

    // Create tracer
    adapter := observability.NewLoggerAdapter(log)
    tracer := observability.NewSchedulerTracerWithConfig(adapter, config)

    // Start tracking session
    session := tracer.StartSession(ctx, jobID, "Job Name", tenantID)

    // Record scheduling decisions
    tracer.RecordDecision(ctx, session, observability.SchedulingDecision{
        DecisionType:      "schedule_item",
        ItemID:            itemID,
        FinalScheduleTime: scheduleTime,
    })

    // Complete session with results
    tracer.CompleteSession(ctx, session, observability.SchedulingResults{
        TotalItemsProcessed: 100,
        ItemsScheduled:      85,
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

## Storage Options

The `SchedulerTracer` supports multiple storage backends:

- **Local**: Stores snapshots as JSON files
- **S3**: Stores snapshots in AWS S3 bucket
- **PostgreSQL**: Stores snapshots in database (coming soon)
- **Disabled**: Structured logs only, no persistence

## Configuration

```go
// Local storage (default)
config := observability.DefaultStorageConfig()

// Disabled storage (logs only)
config := observability.DisabledStorageConfig()

// Custom configuration
config := observability.StorageConfig{
    Enabled:   true,
    Type:      "local",
    LocalPath: "/var/logs/sessions",
}
```

## License

MIT
