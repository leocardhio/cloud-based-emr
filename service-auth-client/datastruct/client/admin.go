package client_credential

import (
	"errors"
)

var (
	ClientNotFoundError = errors.New("client record not found")
)

type GetClientData struct {
	ID           any    `json:"_id" bson:"_id"`
	ClientID     string `json:"client_id" binding:"required" bson:"client_id"`
	ClientSecret string `json:"client_secret" binding:"required" bson:"client_secret"`
}

func (u *GetClientData) CheckSecret(secret string) error {
	if secret != u.ClientSecret {
		return IncorrectCredentialError
	}

	return nil
}
