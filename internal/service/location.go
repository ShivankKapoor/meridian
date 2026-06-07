package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"meridian/internal/models"
	"meridian/internal/repository"
)

type LocationService struct {
	repo *repository.RedisRepository
}

func NewLocationService(repo *repository.RedisRepository) *LocationService {
	return &LocationService{repo: repo}
}

func getLocationViaIPAPI(ip string) (*models.Location, error) {
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var loc models.Location
	if err := json.NewDecoder(resp.Body).Decode(&loc); err != nil {
		return nil, err
	}

	return &loc, nil
}

func (s *LocationService) GetLocation(ctx context.Context, ip string) (*models.Location, error) {
	slog.Info("fetching location", "ip", ip)

	loc, err := s.getLocationViaRedis(ctx, ip)
	if err == nil {
		slog.Info("cache hit", "ip", ip)
		return loc, nil
	}

	slog.Info("cache miss, calling ip-api", "ip", ip)
	loc, err = getLocationViaIPAPI(ip)
	if err != nil {
		slog.Error("failed to fetch location from ip-api", "ip", ip, "error", err)
		return nil, err
	}

	go func() {
		slog.Info("storing location in redis", "ip", ip)
		s.repo.SetCache(context.Background(), ip, loc)
	}()

	return loc, nil
}

func (s *LocationService) getLocationViaRedis(ctx context.Context, ip string) (*models.Location, error) {
	var loc models.Location
	if err := s.repo.GetCache(ctx, ip, &loc); err != nil {
		return nil, err
	}

	return &loc, nil
}
