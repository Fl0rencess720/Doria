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
	Pages     []Page    `gorm:"foreignKey:UserID"`
}

type Page struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"index;not null"`
	SegmentID   uint      `gorm:"index"`
	UserInput   string    `gorm:"type:text"`
	AgentOutput string    `gorm:"type:text"`
	Status      string    `gorm:"type:text;not null;check:status IN ('in_stm','in_mtm','invalid')"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

type MateMessage struct {
	UserID uint `json:"user_id"`
}
