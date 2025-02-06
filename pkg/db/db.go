package db

import (
	"log"

	"github.com/AthulKrishna2501/go-grpc-cart-service/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func Init(url string) Handler {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&models.Cart{})
	db.AutoMigrate(&models.CartItem{})

	return Handler{db}
}
