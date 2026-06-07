package service

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

var discordClient = &http.Client{Timeout: 30 * time.Second}

type DiscordService struct {
	webhook string
}

func NewDiscordService(webhook string) *DiscordService {
	return &DiscordService{webhook: webhook}
}

type discordEmbed struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Color       int          `json:"color"`
	Fields      []embedField `json:"fields"`
	Footer      embedFooter  `json:"footer"`
	Timestamp   string       `json:"timestamp"`
}

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type embedFooter struct {
	Text string `json:"text"`
}

type discordPayload struct {
	Embeds []discordEmbed `json:"embeds"`
}

func (d *DiscordService) send(payload discordPayload) {
	if d.webhook == "" {
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("discord: failed to marshal payload", "error", err)
		return
	}

	resp, err := discordClient.Post(d.webhook, "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Error("discord: failed to send notification", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error("discord: unexpected status", "status", resp.StatusCode)
	}
}

func (d *DiscordService) NotifyCacheMiss(ip string) {
	d.send(discordPayload{
		Embeds: []discordEmbed{
			{
				Title:       "Cache Miss",
				Description: "No cached result found — falling back to ip-api lookup.",
				Color:       0xE67E22, // orange
				Fields: []embedField{
					{Name: "IP Address", Value: "`" + ip + "`", Inline: true},
					{Name: "Source", Value: "ip-api.com", Inline: true},
				},
				Footer:    embedFooter{Text: "Meridian"},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	})
}

func (d *DiscordService) NotifyRedisDown(err error) {
	d.send(discordPayload{
		Embeds: []discordEmbed{
			{
				Title:       "Redis Unreachable",
				Description: "Failed to connect to Redis — requests will hit ip-api directly until Redis recovers.",
				Color:       0xE74C3C, // red
				Fields: []embedField{
					{Name: "Error", Value: "`" + err.Error() + "`", Inline: false},
				},
				Footer:    embedFooter{Text: "Meridian"},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	})
}
