package domain

type Product struct {
	ID    string `json:"id" gorm:"primaryKey;type:text"`
	Name  string `json:"name" gorm:"type:text;not null"`
	Price int    `json:"price" gorm:"not null"`
}
