package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"meridian/internal/models"
	"meridian/internal/repository"
)

var ipAPIClient = &http.Client{Timeout: 30 * time.Second}

type LocationService struct {
	repo    *repository.RedisRepository
	discord *DiscordService
}

func NewLocationService(repo *repository.RedisRepository, discord *DiscordService) *LocationService {
	return &LocationService{repo: repo, discord: discord}
}

type ipAPIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	models.Location
}

func getLocationViaIPAPI(ctx context.Context, ip string) (*models.Location, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://ip-api.com/json/%s", ip), nil)
	if err != nil {
		return nil, err
	}

	resp, err := ipAPIClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("ip-api: %s", result.Message)
	}

	return &result.Location, nil
}

func (s *LocationService) GetLocation(ctx context.Context, ip string) (*models.Location, error) {
	slog.Info("fetching location", "ip", ip)

	loc, err := s.getLocationViaRedis(ctx, ip)
	if err == nil {
		slog.Info("cache hit", "ip", ip)
		return loc, nil
	}

	if !errors.Is(err, redis.Nil) {
		slog.Error("redis error", "error", err)
		go s.discord.NotifyRedisDown(err)
	}

	slog.Info("cache miss, calling ip-api", "ip", ip)
	go s.discord.NotifyCacheMiss(ip)
	loc, err = getLocationViaIPAPI(ctx, ip)
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
