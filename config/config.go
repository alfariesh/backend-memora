package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type (
	// Config -.
	Config struct {
		App       app
		HTTP      http
		Log       log
		PG        pg
		GRPC      grpc
		RMQ       rmq
		NATS      nats
		JWT       jwt
		Metrics   metrics
		Swagger   swagger
		Resend    resend
		OneSignal oneSignal
		Worker    worker
	}

	// App -.
	app struct {
		Name    string `env:"APP_NAME,required"`
		Version string `env:"APP_VERSION,required"`
	}

	// HTTP -.
	http struct {
		Port                          string `env:"HTTP_PORT,required"`
		UsePreforkMode                bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
		CORSEnabled                   bool   `env:"HTTP_CORS_ENABLED" envDefault:"true"`
		AllowedOrigins                string `env:"HTTP_ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://127.0.0.1:3000,http://localhost:5173,http://127.0.0.1:5173"`
		AllowedMethods                string `env:"HTTP_ALLOWED_METHODS" envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
		AllowedHeaders                string `env:"HTTP_ALLOWED_HEADERS" envDefault:"Authorization,Content-Type,Accept,X-Request-ID,X-Correlation-ID"`
		ExposedHeaders                string `env:"HTTP_EXPOSED_HEADERS" envDefault:"X-Request-ID,X-Correlation-ID"`
		CORSAllowCredentials          bool   `env:"HTTP_CORS_ALLOW_CREDENTIALS" envDefault:"false"`
		CORSMaxAge                    int    `env:"HTTP_CORS_MAX_AGE" envDefault:"600"`
		SecurityHeadersEnabled        bool   `env:"HTTP_SECURITY_HEADERS_ENABLED" envDefault:"true"`
		SecurityHSTSEnabled           bool   `env:"HTTP_SECURITY_HSTS_ENABLED" envDefault:"false"`
		SecurityHSTSMaxAge            int    `env:"HTTP_SECURITY_HSTS_MAX_AGE" envDefault:"31536000"`
		SecurityHSTSIncludeSubdomains bool   `env:"HTTP_SECURITY_HSTS_INCLUDE_SUBDOMAINS" envDefault:"true"`
		SecurityHSTSPreload           bool   `env:"HTTP_SECURITY_HSTS_PRELOAD" envDefault:"false"`
	}

	// Log -.
	log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	// PG -.
	pg struct {
		PoolMax int    `env:"PG_POOL_MAX,required"`
		URL     string `env:"PG_URL,required"`
	}

	// GRPC -.
	grpc struct {
		Port string `env:"GRPC_PORT,required"`
	}

	// RMQ -.
	rmq struct {
		ServerExchange string `env:"RMQ_RPC_SERVER,required"`
		ClientExchange string `env:"RMQ_RPC_CLIENT,required"`
		URL            string `env:"RMQ_URL,required"`
	}

	// NATS -.
	nats struct {
		ServerExchange string `env:"NATS_RPC_SERVER,required"`
		URL            string `env:"NATS_URL,required"`
	}

	// JWT -.
	jwt struct {
		Secret             string        `env:"JWT_SECRET,required"`
		TokenExpiry        time.Duration `env:"JWT_TOKEN_EXPIRY" envDefault:"15m"`
		RefreshTokenExpiry time.Duration `env:"JWT_REFRESH_TOKEN_EXPIRY" envDefault:"720h"`
	}

	// Metrics -.
	metrics struct {
		Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	}

	// Swagger -.
	swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}

	// Resend Email Service -.
	resend struct {
		APIKey    string `env:"RESEND_API_KEY"    envDefault:""`
		FromEmail string `env:"RESEND_FROM_EMAIL" envDefault:""`
	}

	// OneSignal Push Service -.
	oneSignal struct {
		AppID      string `env:"ONESIGNAL_APP_ID"       envDefault:""`
		RESTAPIKey string `env:"ONESIGNAL_REST_API_KEY" envDefault:""`
	}

	// Worker -.
	worker struct {
		BatchSize    int           `env:"REMINDER_WORKER_BATCH_SIZE" envDefault:"50"`
		PollInterval time.Duration `env:"REMINDER_WORKER_POLL_INTERVAL" envDefault:"1m"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
