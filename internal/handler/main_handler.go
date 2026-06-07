package handler

import (
	"log/slog"
	"net/http"
)

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

func (h *HomeHandler) Home(w http.ResponseWriter, r *http.Request) {
	slog.Info("Home endpoint called")

	resp := "<h1>Welcome to Meridian</h1>"

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte(resp))
}
