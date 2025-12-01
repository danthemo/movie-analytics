package models

import (
	"time"

	"gorm.io/datatypes"
)

type Movie struct {
	ID          uint `gorm:"primaryKey"`
	Title       string
	Year        uint
	Description string       `gorm:"type:text"`
	PosterUrl   string       `gorm:"type:text"`
	TrailerUrl  string       `gorm:"type:text"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	Comments    []RawComment `gorm:"foreignKey:MovieID;constraint:OnDelete:CASCADE"`
}

type RawComment struct {
	ID        uint `gorm:"primaryKey"`
	MovieID   uint
	Movie     Movie     `gorm:"foreignKey:MovieID;references:ID"`
	Source    string    `gorm:"type:text"`
	Author    string    `gorm:"type:text"`
	Text      string    `gorm:"type:text"`
	ScrapedAt time.Time `gorm:"autoCreateTime"`
}

type MovieInsight struct {
	MovieID      uint  `gorm:"primaryKey"`
	Movie        Movie `gorm:"foreignKey:MovieID;references:ID"`
	AvgSentiment float64
	Summary      string `gorm:"type:text"`
	TopPositive  datatypes.JSON
	TopNegative  datatypes.JSON
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
