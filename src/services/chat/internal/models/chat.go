package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type User struct {
	ID            uint           `gorm:"primaryKey"`
	Username      string         `gorm:"type:text;unique;not null"`
	Status        string         `gorm:"type:text;not null;check:status IN ('user','visitor')"`
	Phone         *string        `gorm:"type:text;unique"`
	Password      string         `gorm:"type:text;not null"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	Conversations []Conversation `gorm:"foreignKey:UserID"`
}

type Conversation struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Title     string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Messages  []Message `gorm:"foreignKey:ConversationID"`
	User      User      `gorm:"foreignKey:UserID"`
}

type Message struct {
	ID             uint         `gorm:"primaryKey"`
	ConversationID uint         `gorm:"not null;index"`
	Role           string       `gorm:"type:text;not null"`
	Content        JSONContent  `gorm:"type:jsonb;not null"`
	CreatedAt      time.Time    `gorm:"autoCreateTime"`
	Conversation   Conversation `gorm:"foreignKey:ConversationID"`
}

type Content struct {
	Text string `json:"text"`
}

type JSONContent Content

func (c JSONContent) Value() (driver.Value, error) {
	if c.Text == "" {
		return nil, nil
	}

	return json.Marshal(c)
}

func (c *JSONContent) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, c)
}
