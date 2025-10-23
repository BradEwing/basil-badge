package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var (
	rankingURL = "https://data.basil-ladder.net/stats/ranking.json"
)

// Shields.io response format
type ShieldsResponse struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

// Bot ranking data structures
type Bot struct {
	Name   string `json:"botName"`
	Rating int    `json:"rating,omitempty"`
	Rank   string `json:"rank,omitempty"`
}

type RankingData struct {
	Rankings []Bot `json:"rankings,omitempty"`
	// Handle case where data is directly an array
	Bots []Bot `json:"-"`
}

type api struct {
	client *http.Client
	logger *zap.Logger
}

// Badge endpoint handler
func (a *api) BadgeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	botName, err := url.QueryUnescape(vars["botName"])
	if err != nil {
		botName = vars["botName"] // Use as-is if decode fails
	}

	w.Header().Set("Content-Type", "application/json")

	bots, err := a.fetchRankingData()
	if err != nil {
		a.logger.Error("failed to fetch ranking data", zap.Error(err))
		response := ShieldsResponse{
			SchemaVersion: 1,
			Label:         "BASIL",
			Message:       "API Error",
			Color:         "red",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	bot := findBot(bots, botName)
	if bot == nil {
		a.logger.Error("failed to find bot", zap.Any("bots", bots), zap.String("bot_name", botName))
		response := ShieldsResponse{
			SchemaVersion: 1,
			Label:         "BASIL",
			Message:       "Bot not found",
			Color:         "red",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	color := "green"

	response := ShieldsResponse{
		SchemaVersion: 1,
		Label:         "BASIL",
		Message:       fmt.Sprintf("%d ELO (%s)", bot.Rating, bot.Rank),
		Color:         color,
	}
	a.logger.Info("fetched bot", zap.String("name", botName), zap.String("rank", bot.Rank), zap.Int("rating", bot.Rating))

	json.NewEncoder(w).Encode(response)
}

func (a *api) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := map[string]interface{}{
		"status": "ok",
	}

	json.NewEncoder(w).Encode(health)
}

func (a *api) fetchRankingData() ([]Bot, error) {
	resp, err := a.client.Get(rankingURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var bots []Bot
	var rankingData RankingData

	defer resp.Body.Close()

	var rawData json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(rawData, &bots); err != nil {
		if err := json.Unmarshal(rawData, &rankingData); err != nil {
			return nil, err
		}
		bots = rankingData.Rankings
	}

	if len(bots) == 0 {
		return []Bot{}, nil
	}

	return bots, nil
}

// Find bot by name (case-insensitive)
func findBot(bots []Bot, name string) *Bot {
	lowerName := strings.ToLower(name)
	for _, bot := range bots {
		if strings.ToLower(bot.Name) == lowerName {
			return &bot
		}
	}
	return nil
}
