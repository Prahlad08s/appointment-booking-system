package main

import (
	"fmt"

	"appointment-booking-system/config"
	"appointment-booking-system/repositories"
	"appointment-booking-system/router"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting Appointment Booking System")

	cfg := config.LoadConfig(logger)

	db := repositories.InitDB(cfg, logger)

	r := router.SetupRouter(db, logger)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	logger.Info("Server starting", zap.String("address", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
