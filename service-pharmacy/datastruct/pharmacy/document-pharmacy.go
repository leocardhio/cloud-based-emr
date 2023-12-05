package pharmacy

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Pharmacy struct {
	ID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`

	ClientID  string     `json:"client_id" bson:"client_id"`
	Signature *string    `json:"signature" bson:"signature"`
	Peresepan DrugRecipe `json:"peresepan" binding:"required" bson:"peresepan"`

	Dispensing          *Dispensing       `json:"dispensing" binding:"required" bson:"dispensing,omitempty"`
	DispensingEncrypted *primitive.Binary `json:"encrypted_dispensing" bson:"encrypted_dispensing"`

	CreatedAt *time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-" bson:"deleted_at"`
}
