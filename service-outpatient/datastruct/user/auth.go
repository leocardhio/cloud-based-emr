package user

import (
	"errors"
	"service-outpatient/datastruct"

	"github.com/golang-jwt/jwt/v4"
)

var (
	IncorrectCredentialError = errors.New("incorrect email or password")
	AuthorizationHeaderError = errors.New("error extracting authorization header")
	NotAuthorizedError       = errors.New("forbidden access")
	UnauthorizedIssuerError  = errors.New("unauthorized token issuer")
)

type Credential struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Claim struct {
	Role datastruct.RoleType `json:"role" binding:"required"`
	jwt.RegisteredClaims
}
