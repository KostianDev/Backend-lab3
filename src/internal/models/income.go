package models

import "time"

// Income represents an incoming transaction credited to a user's account.
type Income struct {
	BaseModel

	AccountID uint `gorm:"not null;index"`
	UserID    uint `gorm:"not null;index"`

	AmountCents int64     `gorm:"not null"`
	Source      string    `gorm:"size:255;not null"`
	ReceivedAt  time.Time `gorm:"not null"`

	Notes string `gorm:"size:512"`

	Account *Account `gorm:"constraint:OnDelete:CASCADE"`
	User    *User    `gorm:"constraint:OnDelete:CASCADE"`
}
