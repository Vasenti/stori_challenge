package domain

type User struct {
	Email string `gorm:"primaryKey;uniqueIndex;size:320"`
	Transactions []Transaction `gorm:"foreignKey:UserEmail;references:Email"`
}

func (User) TableName() string { return "users" }