package caddy

import (
	"context"
	"log/slog"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/pocketbase/pocketbase/tools/logger"
	"github.com/pocketbase/pocketbase/tools/types"
	"go.uber.org/zap"
)

func (h *SitePodHandler) installPocketBaseLogger() error {
	if h.app == nil || h.logger == nil {
		return nil
	}

	pbLogger := h.logger.Named("pocketbase")
	handler := logger.NewBatchHandler(logger.BatchOptions{
		Level:     pocketBaseLogMinLevel(h.app),
		BatchSize: 200,
		BeforeAddFunc: func(_ context.Context, logEntry *logger.Log) bool {
			logPocketBaseEntry(pbLogger, logEntry)
			settings := h.app.Settings()
			return settings != nil && settings.Logs.MaxDays > 0
		},
		WriteFunc: func(_ context.Context, logs []*logger.Log) error {
			settings := h.app.Settings()
			if !h.app.IsBootstrapped() || settings == nil || settings.Logs.MaxDays == 0 {
				return nil
			}

			return h.app.AuxRunInTransaction(func(txApp core.App) error {
				model := &core.Log{}
				for _, entry := range logs {
					model.MarkAsNew()
					model.Id = core.GenerateDefaultRandomId()
					model.Level = int(entry.Level)
					model.Message = entry.Message
					model.Data = entry.Data
					model.Created, _ = types.ParseDateTime(entry.Time)
					if err := txApp.AuxSave(model); err != nil {
						pbLogger.Debug("Failed to write pocketbase log", zap.Error(err))
					}
				}
				return nil
			})
		},
	})

	*h.app.Logger() = *slog.New(handler)
	h.startPocketBaseLogFlusher(handler)

	return nil
}

func pocketBaseLogMinLevel(app *pocketbase.PocketBase) slog.Level {
	if app == nil {
		return slog.LevelInfo
	}
	if app.IsDev() {
		return slog.Level(-99999)
	}
	if app.Settings() != nil {
		return slog.Level(app.Settings().Logs.MinLevel)
	}
	return slog.LevelInfo
}

func (h *SitePodHandler) startPocketBaseLogFlusher(handler *logger.BatchHandler) {
	if handler == nil || h.app == nil {
		return
	}

	ticker := time.NewTicker(3 * time.Second)
	done := make(chan struct{}, 1)

	go func() {
		ctx := context.Background()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				_ = handler.WriteAll(ctx)
			}
		}
	}()

	h.app.OnTerminate().Bind(&hook.Handler[*core.TerminateEvent]{
		Id: "__sitepodPBLoggerOnTerminate__",
		Func: func(e *core.TerminateEvent) error {
			_ = handler.WriteAll(context.Background())
			ticker.Stop()
			select {
			case done <- struct{}{}:
			default:
			}
			return e.Next()
		},
		Priority: -998,
	})
}

func logPocketBaseEntry(zapLogger *zap.Logger, entry *logger.Log) {
	if zapLogger == nil || entry == nil {
		return
	}

	fields := make([]zap.Field, 0, len(entry.Data)+1)
	for key, value := range entry.Data {
		fields = append(fields, zap.Any(key, value))
	}
	if !entry.Time.IsZero() {
		fields = append(fields, zap.Time("time", entry.Time))
	}

	switch {
	case entry.Level >= slog.LevelError:
		zapLogger.Error(entry.Message, fields...)
	case entry.Level >= slog.LevelWarn:
		zapLogger.Warn(entry.Message, fields...)
	case entry.Level >= slog.LevelInfo:
		zapLogger.Info(entry.Message, fields...)
	default:
		zapLogger.Debug(entry.Message, fields...)
	}
}
