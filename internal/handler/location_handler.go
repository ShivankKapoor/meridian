package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"meridian/internal/service"
)

type LocationHandler struct {
	locationService *service.LocationService
}

func NewLocationHandler(locationService *service.LocationService) *LocationHandler {
	return &LocationHandler{locationService: locationService}
}

func (h *LocationHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	ip := r.PathValue("ip")

	if net.ParseIP(ip) == nil {
		http.Error(w, "invalid IP address", http.StatusBadRequest)
		return
	}

	loc, err := h.locationService.GetLocation(r.Context(), ip)
	if err != nil {
		if errors.Is(err, service.ErrPrivateIP) {
			http.Error(w, "location unavailable for private/reserved IP addresses", http.StatusBadRequest)
			return
		}
		slog.Error("failed to get location", "ip", ip, "error", err)
		http.Error(w, "failed to get location", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loc)
}
