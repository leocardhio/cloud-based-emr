package user

import (
	"service-radiology/datastruct"
	"time"
)

type ConsentData struct {
	ClientID     string `json:"client_id" binding:"required" bson:"client_id"`
	ConsentGiver string `json:"consent_giver" binding:"required" bson:"consent_giver"`
}

type PatientConsent struct {
	Signature *string `json:"signature" binding:"required" bson:"signature"`
	NoIHS     string  `json:"no_ihs" binding:"required" bson:"no_ihs"`

	// list dari organizationID (ClientID)
	ConsentTo []ConsentData `json:"consent_to" binding:"required" bson:"consent_to"`

	CreatedAt *time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-" bson:"deleted_at"`
}

type ConsentBody struct {
	NoIHS        string                    `json:"no_ihs" binding:"required" bson:"no_ihs"`
	ConsentType  datastruct.PatientConsent `json:"consent_type" bson:"consent_type"`
	ConsentGiver string                    `json:"consent_giver" binding:"required" bson:"consent_giver"`
}
