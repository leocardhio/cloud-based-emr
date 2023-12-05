package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"service-auth-client/logger"
	"strings"

	"github.com/kelseyhightower/envconfig"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
)

var (
	JWTPrivateKey string
	JWTDuration   int

	TimestampSkew int
)

type Config struct {
	RESTHost string `envconfig:"REST_HOST" default:"localhost"`
	RESTPort int    `envconfig:"REST_PORT" default:"8079"`

	DBUser       string `envconfig:"DB_USER" default:""`
	DBPassword   string `envconfig:"DB_PASSWORD" default:""` //aws
	DBClusterURL string `envconfig:"DB_CLUSTER_URL" default:""`

	JWTPrivateKey string `envconfig:"JWT_PRIVATE_KEY" default:""` // base64 format
	JWTDuration   int    `envconfig:"JWT_DURATION" default:"900"`

	SMProjectId   string `envconfig:"SM_PROJECT_ID" default:""`
	SecretVersion string `envconfig:"SECRET_VERSION" default:"1"`

	TimestampSkew int `envconfig:"TIMESTAMP_SKEW" default:"5000"` //ms
}

func Get() Config {
	cfg := Config{}
	envconfig.MustProcess("", &cfg)
	AccessSecret(&cfg)

	JWTDuration = cfg.JWTDuration
	JWTPrivateKey = strings.ReplaceAll(cfg.JWTPrivateKey, "\\n", "\n")

	TimestampSkew = cfg.TimestampSkew

	cfg.DBUser = url.QueryEscape(cfg.DBUser)
	cfg.DBPassword = url.QueryEscape(cfg.DBPassword)

	return cfg
}

func AccessSecret(cfg *Config) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		logger.LogFatal.Fatalf("failed to setup client: %v", err)
	}
	defer client.Close()

	secretJwtPrivate, err := InitSecretConfig(&ctx, cfg.SMProjectId, cfg.JWTPrivateKey, cfg.SecretVersion).
		AccessSecretResource(client)
	if err != nil {
		logger.LogFatal.Fatalf("failed to access jwt secret: %v", err)
	}

	secretDbPrivate, err := InitSecretConfig(&ctx, cfg.SMProjectId, cfg.DBPassword, cfg.SecretVersion).
		AccessSecretResource(client)
	if err != nil {
		logger.LogFatal.Fatalf("failed to access db secret: %v", err)
	}

	cfg.JWTPrivateKey = string(secretJwtPrivate.Payload.Data)
	cfg.DBPassword = string(secretDbPrivate.Payload.Data)
}

func AccessKeyFromFile(keyName string) string {
	path, _ := filepath.Rel("..", fmt.Sprintf("../key/%s", keyName))
	file, err := os.ReadFile(path)
	if err != nil {
		logger.LogError.Fatalf("fail to open local key file: %s", keyName)
	}

	return string(file)
}
