package main

import (
	"InternalAssetManagement/database"
	"InternalAssetManagement/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

const shutDownTimeOut = 10 * time.Second

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	srv := server.SetupRoutes()
	if err := database.ConnectAndMigrate(
		host,
		port,
		dbName,
		user,
		password,
		database.SSLModeDisable); err != nil {
		logrus.Panicf("Failed to initialize and migrate database with error: %+v", err)
	}
	logrus.Print("migration successful!!")

	go func() {
		if err := srv.Run(":8080"); err != nil && err != http.ErrServerClosed {
			logrus.Panicf("Failed to run server with error: %+v", err)
		}
	}()
	logrus.Print("Server started at :8080")

	<-done

	logrus.Info("shutting down server")
	if err := database.ShutdownDatabase(); err != nil {
		logrus.WithError(err).Error("failed to close database connection")
	}
	if err := srv.Shutdown(shutDownTimeOut); err != nil {
		logrus.WithError(err).Panic("failed to gracefully shutdown server")
	}
}
