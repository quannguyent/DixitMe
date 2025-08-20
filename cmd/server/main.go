// @title DixitMe API
// @version 1.0
// @description API for DixitMe - Online Dixit Card Game
// @description This API provides endpoints for managing players, games, and real-time gameplay through WebSocket connections.
// @contact.name DixitMe API Support
// @contact.email support@dixitme.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
package main

import (
	_ "dixitme/docs" // Import docs for swagger
	"dixitme/internal/app"
	"dixitme/internal/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create and initialize the application
	application, err := app.NewApp()
	if err != nil {
		log := logger.GetLogger()
		log.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		application.Cleanup()
		os.Exit(0)
	}()

	// Start the server
	if err := application.Run(); err != nil {
		log := logger.GetLogger()
		log.Error("Failed to start server", "error", err)
		application.Cleanup()
		os.Exit(1)
	}
}
