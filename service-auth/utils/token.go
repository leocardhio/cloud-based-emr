package utils

import (
	"service-auth/datastruct"
	user "service-auth/datastruct/user"
	"strings"
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
	claim := &user.Claim{
		Role: j.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    j.Issuer,
			Subject:   j.Subject,
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

func VerifyToken(tokenString, jwtPublicKey string) (*user.Claim, error) {
	publicKey, err := jwt.ParseEdPublicKeyFromPEM([]byte(jwtPublicKey))
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &user.Claim{}, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if err := token.Claims.Valid(); err != nil {
		return nil, err
	}

	// note: possibly need error checking
	claim, _ := token.Claims.(*user.Claim)
	if claim.Issuer != "13519220@oauth.std.stei.itb.ac.id" {
		return claim, user.UnauthorizedIssuerError
	}

	return claim, nil
}

func ExtractBearerToken(header string) (string, error) {
	if header == "" {
		return "", user.AuthorizationHeaderError
	}

	token := strings.Split(header, " ")
	if len(token) != 2 {
		return "", user.AuthorizationHeaderError
	}

	return token[1], nil
}
