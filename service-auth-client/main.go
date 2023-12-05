package main

import (
	"fmt"
	"service-auth-client/config"
	"service-auth-client/db"
	"service-auth-client/logger"
	"service-auth-client/router"
)

func main() {
	cfg := config.Get()

	client := db.ConnectDB(&cfg)
	defer db.DisconnectDB(client)

	router := router.InitRouter(client)

	router.Run(fmt.Sprintf("%s:%d", cfg.RESTHost, cfg.RESTPort))

	// Start the server
	logger.LogInfo.Printf("Server started on http://%s:%d\n", cfg.RESTHost, cfg.RESTPort)
	logger.LogFatal.Fatal(router.Run(":8080"))
}
