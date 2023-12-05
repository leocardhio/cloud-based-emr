package client_credential

import (
	"errors"
	"service-auth-client/datastruct"

	"github.com/golang-jwt/jwt/v4"
)

var (
	IncorrectCredentialError = errors.New("incorrect credentials")
	AuthorizationHeaderError = errors.New("error extracting authorization header")
	NotAuthorizedError       = errors.New("forbidden access")
)

type Credential struct {
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
	AdminName    string `json:"admin_name" binding:"required"`
}

type Claim struct {
	Role datastruct.RoleType `json:"role" binding:"required"`
	jwt.RegisteredClaims
}
