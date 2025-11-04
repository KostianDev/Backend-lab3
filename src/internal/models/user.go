package models

import "gorm.io/gorm"

// User represents an application user owning a financial account.
type User struct {
	BaseModel

	Email        string `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string `gorm:"size:255;not null"`

	DefaultCurrency string `gorm:"size:3;not null"`

	Account Account `gorm:"constraint:OnDelete:CASCADE"`

	Expenses []Expense `gorm:"constraint:OnDelete:CASCADE"`
	Incomes  []Income  `gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate ensures default currency is set when missing.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.DefaultCurrency == "" {
		u.DefaultCurrency = "UAH"
	}
	return nil
}
