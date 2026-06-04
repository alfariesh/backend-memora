package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/repo/persistent"
	"github.com/evrone/go-clean-template/internal/repo/webapi"
	"github.com/evrone/go-clean-template/internal/usecase/reminder"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	l := logger.New(cfg.Log.Level)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("worker - postgres.New: %w", err))
	}
	defer pg.Close()

	if err = waitForPostgres(ctx, l, pg); err != nil {
		l.Fatal(fmt.Errorf("worker - waitForPostgres: %w", err))
	}

	userRepo := persistent.NewUserRepo(pg)
	importantDayRepo := persistent.NewImportantDayRepo(pg)
	reminderJobRepo := persistent.NewReminderJobRepo(pg)
	userSettingsRepo := persistent.NewUserSettingsRepo(pg)
	notificationRepo := persistent.NewNotificationRepo(pg)
	deviceTokenRepo := persistent.NewDeviceTokenRepo(pg)

	uc := reminder.New(
		reminderJobRepo,
		importantDayRepo,
		userRepo,
		userSettingsRepo,
		notificationRepo,
		deviceTokenRepo,
		webapi.NewCloudflareEmailSender(cfg.Email.AccountID, cfg.Email.APIToken, cfg.Email.FromEmail),
		webapi.NewExpoPushSender(cfg.Expo.PushAccessToken),
	)

	ticker := time.NewTicker(cfg.Worker.PollInterval)
	defer ticker.Stop()

	l.Info("worker - started")

	for {
		processed, runErr := uc.RunOnce(ctx, time.Now().UTC(), cfg.Worker.BatchSize)
		if runErr != nil {
			l.Error(fmt.Errorf("worker - RunOnce: %w", runErr))
		} else if processed > 0 {
			l.Info("worker - processed reminder jobs: %d", processed)
		}

		select {
		case <-ctx.Done():
			l.Info("worker - stopped")

			return
		case <-ticker.C:
		}
	}
}

func waitForPostgres(ctx context.Context, l logger.Interface, pg *postgres.Postgres) error {
	const (
		attempts = 20
		delay    = time.Second
	)

	for attempt := 1; attempt <= attempts; attempt++ {
		if err := pg.Pool.Ping(ctx); err == nil {
			return nil
		}

		if attempt == attempts {
			break
		}

		l.Info("worker - waiting for postgres, attempts left: %d", attempts-attempt)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return pg.Pool.Ping(ctx)
}
