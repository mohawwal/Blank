package main

import (
	"log"
	"whatsapp-bot-app/common"
	"whatsapp-bot-app/internal/models"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	db, err := common.ConnectDB()
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&models.UserModel{}, &models.TransactionModel{})
	if err != nil {
		panic(err)
	}

	log.Println("Migration Completed")
}
