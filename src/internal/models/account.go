package models

// Account keeps aggregated balance for a specific user.
type Account struct {
	BaseModel

	UserID uint `gorm:"uniqueIndex;not null"`

	BalanceCents    int64  `gorm:"not null;default:0"`
	CurrencyISOCode string `gorm:"size:3;not null;default:'UAH'"`

	User *User `gorm:"constraint:OnDelete:CASCADE"`

	Expenses []Expense `gorm:"constraint:OnDelete:CASCADE"`
	Incomes  []Income  `gorm:"constraint:OnDelete:CASCADE"`
}
