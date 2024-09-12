package logger

import (
	"fmt"
	"log/slog"
	"os"
	"simple-chat/internal/lib/logger/slogpretty"

	"github.com/grafana/loki-client-go/loki"
	slogloki "github.com/samber/slog-loki/v3"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

func New(env string) *slog.Logger {
	var log *slog.Logger

	config, _ := loki.NewDefaultConfig("http://loki:3100/loki/api/v1/push")
	client, err := loki.New(config)
	if err != nil {
		println("loki client error: ", err.Error())
		os.Exit(1)
	}

	switch env {
	case EnvLocal:
		log = SetupPrettyLogger()
		return log

	case EnvDev:
		log = slog.New(
			slogloki.Option{Level: slog.LevelDebug, Client: client}.NewLokiHandler(),
		)
		return log
	case EnvProd:
		log = slog.New(
			slogloki.Option{Level: slog.LevelInfo, Client: client}.NewLokiHandler(),
		)
		return log
	default:
		fmt.Printf("Invalid environment: %s. Supported environments are: %s, %s, %s", env, EnvLocal, EnvDev, EnvProd)
		os.Exit(1)
		return nil
	}
}

func SetupPrettyLogger() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOptions: &slog.HandlerOptions{Level: slog.LevelDebug},
	}

	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}
