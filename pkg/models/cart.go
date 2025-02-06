package models

import "gorm.io/gorm"

type Cart struct {
	ID         uint       `gorm:"primaryKey"`
	UserID     int64      `gorm:"index;not null"`
	ProductID  uint64     `gorm:"not null"`
	Items      []CartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE"`
	TotalPrice float64    `gorm:"not null;default:0"`
}

type CartItem struct {
	ID          uint    `gorm:"primaryKey"`
	CartID      uint    `gorm:"index;not null"`
	ProductID   int64   `gorm:"not null"`
	ProductName string  `gorm:"not null"`
	Quantity    int     `gorm:"not null;default:1"`
	Price       float64 `gorm:"not null"`
	Deleted_At  gorm.DeletedAt
}
