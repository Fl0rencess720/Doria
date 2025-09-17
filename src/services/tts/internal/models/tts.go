package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"type:text;unique;not null"`
	Status    string    `gorm:"type:text;not null;check:status IN ('user','visitor')"`
	Phone     *string   `gorm:"type:text;unique"`
	Password  string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
