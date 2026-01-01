package main

import (
	"fmt"
	"whatsapp-bot-app/cmd/api/handlers"
	"whatsapp-bot-app/cmd/api/middlewares"
	"whatsapp-bot-app/cmd/api/services"
	"whatsapp-bot-app/common"
	"whatsapp-bot-app/internal/messaging"
	"whatsapp-bot-app/internal/processor"

	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Application struct {
	logger  echo.Logger
	server  *echo.Echo
	handler handlers.Handler
}

func main() {
	e := echo.New()
	err := godotenv.Load()
	if err != nil {
		e.Logger.Fatal(err.Error())
	}
	db, err := common.ConnectDB()
	if err != nil {
		e.Logger.Fatal(err.Error())
		fmt.Println("Failed to connect to database")
	}

	claudeService := services.NewAIService()
	paystackService := services.NewPaystackService()

	// Initialize WhatsApp messaging adapter
	whatsappAdapter := messaging.NewWhatsAppAdapter()

	// Initialize message processor with WhatsApp adapter
	messageProcessor := processor.NewMessageProcessor(db, claudeService, paystackService, whatsappAdapter)

	h := handlers.Handler{
		DB:               db,
		ClaudeService:    claudeService,
		PaystackService:  paystackService,
		MessageProcessor: messageProcessor,
	}

	app := Application{
		logger:  e.Logger,
		server:  e,
		handler: h,
	}

	fmt.Println(app)
	e.Use(middleware.Logger(), middlewares.CustomMiddleware)
	app.routes(h)
	port := os.Getenv("APP_PORT")
	portAddress := fmt.Sprintf("localhost:%s", port)
	e.Logger.Fatal(e.Start(portAddress))
}
