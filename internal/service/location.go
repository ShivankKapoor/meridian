package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"meridian/internal/models"
	"meridian/internal/repository"
)

var ipAPIClient = &http.Client{Timeout: 30 * time.Second}

type LocationService struct {
	repo    *repository.RedisRepository
	logRepo *repository.MariaDBRepository
	discord *DiscordService
}

func NewLocationService(repo *repository.RedisRepository, logRepo *repository.MariaDBRepository, discord *DiscordService) *LocationService {
	return &LocationService{repo: repo, logRepo: logRepo, discord: discord}
}

func (s *LocationService) logLookup(ip, status string, durationMs *int, loc *models.Location) {
	if err := s.logRepo.LogLookup(context.Background(), ip, status, durationMs, loc); err != nil {
		slog.Error("failed to write lookup log", "ip", ip, "error", err)
		s.discord.NotifyDBDown(err)
	}
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

var ErrPrivateIP = errors.New("ip is a private/reserved address")

var privateRanges = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"127.0.0.0/8",
	"::1/128",
	"fc00::/7",
}

func isPrivateIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(parsed) {
			return true
		}
	}
	return false
}

func (s *LocationService) GetLocation(ctx context.Context, ip string) (*models.Location, error) {
	slog.Info("fetching location", "ip", ip)

	if isPrivateIP(ip) {
		slog.Warn("location requested for private/reserved IP", "ip", ip)
		go s.logLookup(ip, "private_ip", nil, nil)
		return nil, ErrPrivateIP
	}

	loc, err := s.getLocationViaRedis(ctx, ip)
	if err == nil {
		slog.Info("cache hit", "ip", ip)
		go s.logLookup(ip, "cache_hit", nil, loc)
		return loc, nil
	}

	if !errors.Is(err, redis.Nil) {
		slog.Error("redis error", "error", err)
		go s.discord.NotifyRedisDown(err)
	}

	slog.Info("cache miss, calling ip-api", "ip", ip)
	go s.discord.NotifyCacheMiss(ip)
	start := time.Now()
	loc, err = getLocationViaIPAPI(ctx, ip)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		slog.Error("failed to fetch location from ip-api", "ip", ip, "error", err)
		go s.logLookup(ip, "cache_miss", &durationMs, nil)
		return nil, err
	}

	go s.logLookup(ip, "cache_miss", &durationMs, loc)

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
