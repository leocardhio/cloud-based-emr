package utils

import (
	user "service-radiology/datastruct/user"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

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
	if claim.Issuer != "13519220@auth.std.stei.itb.ac.id" {
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
