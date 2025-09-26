package models

import "time"

type User struct {
	ID        uint       `gorm:"primaryKey"`
	Username  string     `gorm:"type:text;unique;not null"`
	Status    string     `gorm:"type:text;not null;check:status IN ('user','visitor')"`
	Phone     *string    `gorm:"type:text;unique"`
	Password  string     `gorm:"type:text;not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	Pages     []*Page    `gorm:"foreignKey:UserID"`
	Segments  []*Segment `gorm:"foreignKey:UserID"`
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

type Segment struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	Overview  string    `gorm:"type:text"`
	Visit     int       `gorm:"type:int;default:0"`
	LastVisit time.Time `gorm:"autoUpdateTime"`
	Pages     []*Page   `gorm:"foreignKey:SegmentID"`
}

type LongTermMemory struct {
	ID      uint   `gorm:"primaryKey"`
	UserID  uint   `gorm:"index;not null"`
	Content string `gorm:"type:text;not null"`
}

type MateMessage struct {
	UserID uint `json:"user_id"`
}

type Correlation struct {
	Page      *Page   `json:"page"`
	Score     float32 `json:"score"`
	SegmentID uint    `json:"segment_id"`
}
