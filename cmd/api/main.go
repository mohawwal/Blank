package main

import (
	"fmt"
	"whatsapp-bot-app/cmd/api/handlers"
	"whatsapp-bot-app/cmd/api/middlewares"
	"whatsapp-bot-app/cmd/api/services"
	"whatsapp-bot-app/common"

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

	twilioService := services.NewTwilioService()
	paystackService := services.NewPaystackService()
	claudeService := services.NewClaudeService()
	// vtpassService := services.NewVTPassService()

	h := handlers.Handler{
		DB:              db,
		TwilioService:   twilioService,
		PaystackService: paystackService,
		ClaudeService:   claudeService,
		// VTPassService:   vtpassService,
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
