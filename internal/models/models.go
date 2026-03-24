// package models

// import (
// 	"time"

// 	"gorm.io/datatypes"
// )

// type Movie struct {
// 	ID          uint `gorm:"primaryKey"`
// 	Title       string
// 	Year        uint
// 	Description string       `gorm:"type:text"`
// 	PosterUrl   string       `gorm:"type:text"`
// 	TrailerUrl  string       `gorm:"type:text"`
// 	CreatedAt   time.Time    `gorm:"autoCreateTime"`
// 	Comments    []RawComment `gorm:"foreignKey:MovieID;constraint:OnDelete:CASCADE"`
// }

// type RawComment struct {
// 	ID        uint `gorm:"primaryKey"`
// 	MovieID   uint
// 	Movie     Movie     `gorm:"foreignKey:MovieID;references:ID"`
// 	Source    string    `gorm:"type:text"`
// 	Author    string    `gorm:"type:text"`
// 	Text      string    `gorm:"type:text"`
// 	ScrapedAt time.Time `gorm:"autoCreateTime"`
// }

// type MovieInsight struct {
// 	MovieID      uint  `gorm:"primaryKey"`
// 	Movie        Movie `gorm:"foreignKey:MovieID;references:ID"`
// 	AvgSentiment float64
// 	Summary      string `gorm:"type:text"`
// 	TopPositive  datatypes.JSON
// 	TopNegative  datatypes.JSON
// 	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
// }

package models

import (
	"time"

	"gorm.io/datatypes"
)

type Movie struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Title       string       `json:"title"`
	Year        uint         `json:"year"`
	Description string       `gorm:"type:text" json:"description"`
	Directors   string       `gorm:"type:text" json:"directors"`
	Actors      string       `gorm:"type:text" json:"actors"`
	PosterUrl   string       `gorm:"type:text" json:"poster_url"`
	TrailerUrl  string       `gorm:"type:text" json:"trailer_url"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	Comments    []RawComment `gorm:"foreignKey:MovieID;constraint:OnDelete:CASCADE" json:"comments"`
}

type RawComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MovieID   uint      `json:"movie_id"`
	Movie     Movie     `gorm:"foreignKey:MovieID;references:ID;constraint:OnDelete:CASCADE" json:"movie"`
	Source    string    `gorm:"type:text" json:"source"`
	Author    string    `gorm:"type:text" json:"author"`
	Text      string    `gorm:"type:text" json:"text"`
	ScrapedAt time.Time `gorm:"autoCreateTime" json:"scraped_at"`
}

type MovieInsight struct {
	MovieID      uint           `gorm:"primaryKey" json:"movie_id"`
	Movie        Movie          `gorm:"foreignKey:MovieID;references:ID;constraint:OnDelete:CASCADE" json:"movie"`
	AvgSentiment float64        `json:"avg_sentiment"`
	Summary      string         `gorm:"type:text" json:"summary"`
	Rating       float64        `json:"rating"`
	TopPositive  datatypes.JSON `json:"top_positive"`
	TopNegative  datatypes.JSON `json:"top_negative"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}
