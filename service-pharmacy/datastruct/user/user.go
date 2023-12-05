package user

import (
	"errors"
	"service-pharmacy/datastruct"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	DuplicateEmailError = errors.New("email has already been taken")
	UserNotFoundError   = errors.New("user record not found")
)

type CreateUserData struct {
	ID             any               `json:"_id" bson:"_id,omitempty"`
	Email          *string           `json:"email" binding:"required" bson:"email,omitempty"`
	EmailEncrypted *primitive.Binary `json:"encrypted_email" bson:"encrypted_email"`

	Password          *string           `json:"password" binding:"required" bson:"password,omitempty"`
	PasswordEncrypted *primitive.Binary `json:"encrypted_password" bson:"encrypted_password"`

	Name          *string           `json:"name" binding:"required" bson:"name,omitempty"`
	NameEncrypted *primitive.Binary `json:"encrypted_name" bson:"encrypted_name"`

	Role datastruct.RoleType `json:"role" binding:"required" bson:"role"`

	CreatedAt *time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-" bson:"deleted_at"`
}

type GetUserData struct {
	Email          *string           `json:"email" binding:"required" bson:"email,omitempty"`
	EmailEncrypted *primitive.Binary `json:"encrypted_email" bson:"encrypted_email"`

	Name          *string           `json:"name" binding:"required" bson:"name,omitempty"`
	NameEncrypted *primitive.Binary `json:"encrypted_name" bson:"encrypted_name"`

	Role datastruct.RoleType `json:"role" binding:"required" bson:"role"`

	CreatedAt *time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-" bson:"deleted_at"`
}

func (u *CreateUserData) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	strHashed := string(hashedPassword)
	u.Password = &strHashed
	return nil
}

func (u *CreateUserData) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(password))
}
