package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	ssogrpc "simple-chat/internal/clients/sso/grpc"
	"simple-chat/internal/config"
	"simple-chat/internal/handlers/auth"
	chatHandler "simple-chat/internal/handlers/chat"
	messageHandler "simple-chat/internal/handlers/message"
	"simple-chat/internal/lib/logger/sl"
	mwLogger "simple-chat/internal/lib/middleware"
	"simple-chat/internal/logger"
	chat_service "simple-chat/internal/services/chat"
	message_service "simple-chat/internal/services/message"
	"simple-chat/internal/storage/chat"
	"simple-chat/internal/storage/message"
	"simple-chat/internal/storage/postgresql"
	"syscall"
	"time"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.Env)
	log.Debug("debug messages are available")
	log.Info("info messages are available")
	log.Warn("warn messages are available")
	log.Error("error messages are available")

	dbPool, err := postgresql.NewConection(context.TODO(), log, cfg.Database)
	if err != nil {
		log.Error("failed connect to database", sl.Err(err))
		os.Exit(1)
	}

	chatDB := chat.NewChatDB(log)
	messageDB := message.NewMessageDB(log)

	chatService := chat_service.NewChatService(log, chatDB, dbPool)
	messageService := message_service.NewMessageServices(log, messageDB, dbPool)
	ssoClient, err := ssogrpc.NewClient(log, cfg.SSOClient)
	if err != nil {
		log.Error("failed to create sso client", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(mwLogger.New(log))
	log.Info("middleware successfully conected")

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	log.Info("cors successfully conected")

	router.Route("/auth", auth.AddAuthHandler(ssoClient, log, cfg.AppID))
	router.Route("/chat", chatHandler.AddChatHandler(log, chatService, messageService, ssoClient, cfg.AppID))
	router.Route("/message", messageHandler.AddMessageHandler(log, messageService, ssoClient, cfg.AppID))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Info("starting server", slog.String("addr", fmt.Sprintf("%s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port)))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to listen and serve", sl.Err(err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSignal := <-stop
	log.Info("stoppping server", slog.String("signal", stopSignal.String()))
	ctx, close := context.WithTimeout(context.Background(), time.Minute)
	defer close()
	srv.Shutdown(ctx)
	log.Info("server was stopped")
}
