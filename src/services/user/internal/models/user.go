package models

import (
	"gorm.io/gorm"
)

type User struct {
	ID       int32  `gorm:"primaryKey;autoIncrement" json:"id"`
	Phone    string `gorm:"uniqueIndex;not null" json:"phone"`
	Password string `gorm:"not null" json:"password"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	// TODO: Implement any pre-create logic here (e.g., password hashing)
	return nil
}