package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Playlist struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name           string          `gorm:"not null"`
	FolderID       uuid.UUID       `gorm:"type:uuid;index"`
	PlaylistFolder *PlaylistFolder `gorm:"foreignKey:FolderID"`
	UserID         uuid.UUID       `gorm:"type:uuid;index;not null"`

	Songs []Song `gorm:"many2many:playlist_songs;constraint:OnDelete:CASCADE;"`
}
