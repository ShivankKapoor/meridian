package main

import (
	"log/slog"
	"meridian/internal/config"
	"meridian/internal/handler"
	"meridian/internal/repository"
	"meridian/internal/service"
	"net/http"
	"os"
)

func main() {
	eVariable := config.Load()

	redisAddr := eVariable.RedisHost + ":" + eVariable.RedisPort
	redisRepo, err := repository.NewRedisRepository(redisAddr, eVariable.RedisPassword)
	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	locationService := service.NewLocationService(redisRepo)
	locationHandler := handler.NewLocationHandler(locationService)

	homeHandler := handler.NewHomeHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", homeHandler.Home)
	mux.HandleFunc("GET /location/{ip}", locationHandler.GetLocation)

	slog.Info("Meridian web engine starting...", "port", eVariable.Port)

	if err := http.ListenAndServe(eVariable.Port, mux); err != nil {
		slog.Error("Failed to start the web server", "error", err)
		os.Exit(1)
	}
}
