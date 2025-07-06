package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	api := api{
		client: &http.Client{
			Timeout: time.Second * 5,
		},
		logger: logger,
	}

	r := mux.NewRouter()

	// Routes
	r.HandleFunc("/badge/{botName}", api.BadgeHandler).Methods("GET")
	r.HandleFunc("/health", api.HealthHandler).Methods("GET")

	logger.Info("BASIL badge server starting", zap.String("port", port))

	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Fatal("fatal api error", zap.Error(err))
	}
}
