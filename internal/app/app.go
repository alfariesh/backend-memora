// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alfariesh/backend-memora/config"
	amqprpc "github.com/alfariesh/backend-memora/internal/controller/amqp_rpc"
	"github.com/alfariesh/backend-memora/internal/controller/grpc"
	grpcmw "github.com/alfariesh/backend-memora/internal/controller/grpc/middleware"
	natsrpc "github.com/alfariesh/backend-memora/internal/controller/nats_rpc"
	"github.com/alfariesh/backend-memora/internal/controller/restapi"
	"github.com/alfariesh/backend-memora/internal/repo/persistent"
	"github.com/alfariesh/backend-memora/internal/repo/webapi"
	"github.com/alfariesh/backend-memora/internal/usecase/device"
	"github.com/alfariesh/backend-memora/internal/usecase/importantday"
	"github.com/alfariesh/backend-memora/internal/usecase/notification"
	"github.com/alfariesh/backend-memora/internal/usecase/reminder"
	"github.com/alfariesh/backend-memora/internal/usecase/task"
	"github.com/alfariesh/backend-memora/internal/usecase/user"
	"github.com/alfariesh/backend-memora/internal/usecase/usersettings"
	"github.com/alfariesh/backend-memora/pkg/grpcserver"
	"github.com/alfariesh/backend-memora/pkg/httpserver"
	"github.com/alfariesh/backend-memora/pkg/jwt"
	"github.com/alfariesh/backend-memora/pkg/logger"
	natsRPCServer "github.com/alfariesh/backend-memora/pkg/nats/nats_rpc/server"
	"github.com/alfariesh/backend-memora/pkg/postgres"
	rmqRPCServer "github.com/alfariesh/backend-memora/pkg/rabbitmq/rmq_rpc/server"
	pbgrpc "google.golang.org/grpc"
)

type useCases struct {
	user         *user.UseCase
	userSettings *usersettings.UseCase
	task         *task.UseCase
	importantDay *importantday.UseCase
	notification *notification.UseCase
	device       *device.UseCase
	reminder     *reminder.UseCase
}

type servers struct {
	rmq  *rmqRPCServer.Server
	nats *natsRPCServer.Server
	grpc *grpcserver.Server
	http *httpserver.Server
}

func initUseCases(cfg *config.Config, pg *postgres.Postgres, jwtManager *jwt.Manager) useCases {
	userRepo := persistent.NewUserRepo(pg)
	userSessionRepo := persistent.NewUserSessionRepo(pg)
	userSettingsRepo := persistent.NewUserSettingsRepo(pg)
	taskRepo := persistent.NewTaskRepo(pg)
	importantDayRepo := persistent.NewImportantDayRepo(pg)
	reminderRuleRepo := persistent.NewReminderRuleRepo(pg)
	reminderJobRepo := persistent.NewReminderJobRepo(pg)
	notificationRepo := persistent.NewNotificationRepo(pg)
	deviceTokenRepo := persistent.NewDeviceTokenRepo(pg)
	emailSender := webapi.NewCloudflareEmailSender(cfg.Email.AccountID, cfg.Email.APIToken, cfg.Email.FromEmail)
	pushSender := webapi.NewExpoPushSender(cfg.Expo.PushAccessToken)

	return useCases{
		user:         user.New(userRepo, jwtManager, user.SessionRepo(userSessionRepo), user.RefreshTokenTTL(cfg.JWT.RefreshTokenExpiry)),
		userSettings: usersettings.New(userSettingsRepo),
		task:         task.New(taskRepo),
		importantDay: importantday.New(importantDayRepo, reminderRuleRepo, reminderJobRepo, userSettingsRepo),
		notification: notification.New(notificationRepo),
		device:       device.New(deviceTokenRepo, device.PushSender(pushSender)),
		reminder: reminder.New(
			reminderJobRepo,
			importantDayRepo,
			userRepo,
			userSettingsRepo,
			notificationRepo,
			deviceTokenRepo,
			emailSender,
			pushSender,
		),
	}
}

func initServers(cfg *config.Config, uc useCases, jwtManager *jwt.Manager, l logger.Interface) servers {
	// RabbitMQ RPC Server
	rmqRouter := amqprpc.NewRouter(uc.user, uc.task, uc.importantDay, uc.notification, uc.device, jwtManager, l)

	rmqServer, err := rmqRPCServer.New(cfg.RMQ.URL, cfg.RMQ.ServerExchange, rmqRouter, l)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - rmqServer - server.New: %w", err))
	}

	// NATS RPC Server
	natsRouter := natsrpc.NewRouter(uc.user, uc.task, uc.importantDay, uc.notification, uc.device, jwtManager, l)

	natsServer, err := natsRPCServer.New(cfg.NATS.URL, cfg.NATS.ServerExchange, natsRouter, l)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - natsServer - server.New: %w", err))
	}

	// gRPC Server
	grpcServer := grpcserver.New(l,
		grpcserver.Port(cfg.GRPC.Port),
		grpcserver.ServerOptions(pbgrpc.UnaryInterceptor(grpcmw.AuthInterceptor(jwtManager))),
	)
	grpc.NewRouter(grpcServer.App, uc.user, uc.task, uc.importantDay, uc.notification, uc.device, l)

	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	restapi.NewRouter(httpServer.App, cfg, uc.user, uc.userSettings, uc.task, uc.importantDay, uc.notification, uc.device, jwtManager, l)

	return servers{
		rmq:  rmqServer,
		nats: natsServer,
		grpc: grpcServer,
		http: httpServer,
	}
}

func (s *servers) startServers() {
	s.rmq.Start()
	s.nats.Start()
	s.grpc.Start()
	s.http.Start()
}

func (s *servers) waitForShutdown(l logger.Interface) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error

	select {
	case sig := <-interrupt:
		l.Info("app - Run - signal: %s", sig.String())
	case err = <-s.http.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	case err = <-s.grpc.Notify():
		l.Error(fmt.Errorf("app - Run - grpcServer.Notify: %w", err))
	case err = <-s.rmq.Notify():
		l.Error(fmt.Errorf("app - Run - rmqServer.Notify: %w", err))
	case err = <-s.nats.Notify():
		l.Error(fmt.Errorf("app - Run - natsServer.Notify: %w", err))
	}

	s.shutdownServers(l)
}

func (s *servers) shutdownServers(l logger.Interface) {
	if err := s.http.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}

	if err := s.grpc.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - grpcServer.Shutdown: %w", err))
	}

	if err := s.rmq.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - rmqServer.Shutdown: %w", err))
	}

	if err := s.nats.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - natsServer.Shutdown: %w", err))
	}
}

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// JWT
	jwtManager := jwt.New(cfg.JWT.Secret, cfg.JWT.TokenExpiry)

	uc := initUseCases(cfg, pg, jwtManager)
	s := initServers(cfg, uc, jwtManager, l)
	s.startServers()
	s.waitForShutdown(l)
}
