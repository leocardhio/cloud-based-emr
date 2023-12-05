package main

import (
	"fmt"
	"service-auth/config"
	"service-auth/db"
	"service-auth/db/csfle"
	"service-auth/logger"
	"service-auth/router"
)

func main() {
	// skew := 5000
	// dur := time.Duration(skew) * time.Millisecond
	// fmt.Println(dur.Milliseconds(), time.Now().Add(3*time.Second).UnixMilli()-time.Now().UnixMilli())
	cfg := config.Get()
	keyVaultNamespace := "encryption.__keyVault"

	client := db.ConnectDB(&cfg)
	defer db.DisconnectDB(client)

	if err := db.CreateKeyVaultIndex(client, keyVaultNamespace); err != nil {
		logger.LogError.Println(err)
		return
	}

	csfle := csfle.InitCSFLE(&cfg, client)

	err := csfle.CreateClientEncryption(keyVaultNamespace).GetKey()
	defer csfle.CloseClient()
	if err != nil {
		logger.LogInfo.Println("DEK Key is not available, creating DEK Key...")
		if err = csfle.MakeKey(); err != nil {
			logger.LogError.Printf("create key error: %v", err)
			return
		}
	}

	router := router.InitRouter(client, csfle)

	router.Run(fmt.Sprintf("%s:%d", cfg.RESTHost, cfg.RESTPort))

	// Start the server
	logger.LogInfo.Printf("Server started on http://%s:%d\n", cfg.RESTHost, cfg.RESTPort)
	logger.LogFatal.Fatal(router.Run(":8080"))
}
