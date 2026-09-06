package domain

import "time"

type Location string

const (
	LocationIndoor  Location = "indoor"
	LocationOutdoor Location = "outdoor"
)

type Table struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TableNumber string    `gorm:"uniqueIndex;size:50;not null" json:"table_number"`
	Capacity    int       `gorm:"not null" json:"capacity"`
	Location    Location  `gorm:"size:20;not null" json:"location"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Table) TableName() string {
	return "tables"
}

func (l Location) Valid() bool {
	return l == LocationIndoor || l == LocationOutdoor
}
