package models

import "time"

// Expense represents an outgoing transaction debited from a user's account.
type Expense struct {
	BaseModel

	AccountID uint `gorm:"not null;index"`
	UserID    uint `gorm:"not null;index"`

	AmountCents int64     `gorm:"not null"`
	Category    string    `gorm:"size:120;not null"`
	IncurredAt  time.Time `gorm:"not null"`

	Description string `gorm:"size:512"`

	Account *Account `gorm:"constraint:OnDelete:CASCADE"`
	User    *User    `gorm:"constraint:OnDelete:CASCADE"`
}
