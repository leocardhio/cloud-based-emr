package utils

import (
	"service-auth-client/datastruct"
	admin_credential "service-auth-client/datastruct/client"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTPayload struct {
	Subject  string
	Audience []string
	Role     datastruct.RoleType
	Issuer   string
}

func (j *JWTPayload) GenerateToken(jwtPrivateKey string, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)
	claim := &admin_credential.Claim{
		Role: j.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Subject:   j.Subject,
			Issuer:    j.Issuer,
			Audience:  j.Audience,
		},
	}

	privateKey, err := jwt.ParseEdPrivateKeyFromPEM([]byte(jwtPrivateKey))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claim)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
