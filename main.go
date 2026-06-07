package main

import (
	"context"
	"log/slog"
	"meridian/internal/config"
	"meridian/internal/handler"
	"meridian/internal/repository"
	"meridian/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	eVariable := config.Load()

	redisAddr := eVariable.RedisHost + ":" + eVariable.RedisPort
	redisRepo := repository.NewRedisRepository(redisAddr, eVariable.RedisPassword)

	discordService := service.NewDiscordService(eVariable.DiscordWebHook)
	locationService := service.NewLocationService(redisRepo, discordService)
	locationHandler := handler.NewLocationHandler(locationService)

	homeHandler := handler.NewHomeHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", homeHandler.Home)
	mux.HandleFunc("GET /location/{ip}", locationHandler.GetLocation)

	srv := &http.Server{Addr: eVariable.Port, Handler: mux}

	go func() {
		slog.Info("Meridian web engine starting...", "port", eVariable.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start the web server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Server stopped")
}
