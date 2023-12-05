package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"service-pharmacy/logger"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/kelseyhightower/envconfig"
)

var (
	JWTPublicKey  string
	RSAPrivateKey string
	RSAPublicKey  string

	TimestampSkew int
)

type Config struct {
	RESTHost string `envconfig:"REST_HOST" default:"localhost"`
	RESTPort int    `envconfig:"REST_PORT" default:"8083"`

	DBUser       string `envconfig:"DB_USER" default:""`
	DBPassword   string `envconfig:"DB_PASSWORD" default:""` //aws
	DBClusterURL string `envconfig:"DB_CLUSTER_URL" default:""`

	SAEmail      string `envconfig:"SA_EMAIL" default:""`
	SAPrivateKey string `envconfig:"SA_PRIVATE_KEY" default:""`

	KMSProjectId string `envconfig:"KMS_PROJECT_ID" default:""`
	KMSLocation  string `envconfig:"KMS_LOCATION" default:""`
	KMSKeyRing   string `envconfig:"KMS_KEY_RING" default:""`
	KMSKeyName   string `envconfig:"KMS_KEY_NAME" default:""`

	JWTPublicKey string `envconfig:"JWT_PUBLIC_KEY" default:""` // base64 format

	SMProjectId   string `envconfig:"SM_PROJECT_ID" default:""`
	SecretVersion string `envconfig:"SECRET_VERSION" default:"1"`

	RSAPrivateKey string `envconfig:"RSA_PRIVATE_KEY" default:""`
	RSAPublicKey  string `envconfig:"RSA_PUBLIC_KEY" default:""`

	TimestampSkew int `envconfig:"TIMESTAMP_SKEW" default:"5000"` //ms
}

func Get() Config {
	cfg := Config{}
	envconfig.MustProcess("", &cfg)
	AccessSecret(&cfg)

	JWTPublicKey = AccessKeyFromFile("jwt_public.pem")
	RSAPublicKey = AccessKeyFromFile("rsa_sign.pub")
	RSAPrivateKey = strings.ReplaceAll(cfg.RSAPrivateKey, "\\n", "\n")

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

	secretSaPrivate, err := InitSecretConfig(&ctx, cfg.SMProjectId, cfg.SAPrivateKey, cfg.SecretVersion).
		AccessSecretResource(client)
	if err != nil {
		logger.LogFatal.Fatalf("failed to access sa secret: %v", err)
	}

	secretRsaPrivate, err := InitSecretConfig(&ctx, cfg.SMProjectId, cfg.RSAPrivateKey, cfg.SecretVersion).
		AccessSecretResource(client)
	if err != nil {
		logger.LogFatal.Fatalf("failed to access rsa secret: %v", err)
	}

	secretDbPrivate, err := InitSecretConfig(&ctx, cfg.SMProjectId, cfg.DBPassword, cfg.SecretVersion).
		AccessSecretResource(client)
	if err != nil {
		logger.LogFatal.Fatalf("failed to access db secret: %v", err)
	}

	cfg.SAPrivateKey = string(secretSaPrivate.Payload.Data)
	cfg.RSAPrivateKey = string(secretRsaPrivate.Payload.Data)
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
