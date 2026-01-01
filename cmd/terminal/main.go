package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"whatsapp-bot-app/cmd/api/services"
	"whatsapp-bot-app/common"
	"whatsapp-bot-app/internal/messaging"
	"whatsapp-bot-app/internal/models"
	"whatsapp-bot-app/internal/processor"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to database
	db, err := common.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the database
	if err := db.AutoMigrate(&models.UserModel{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize services
	aiService := services.NewAIService()
	paystackService := services.NewPaystackService()

	// Initialize terminal messaging adapter
	terminalAdapter := messaging.NewTerminalAdapter()

	// Initialize message processor with terminal adapter
	messageProcessor := processor.NewMessageProcessor(db, aiService, paystackService, terminalAdapter)

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║          Welcome to Blank AI Bot (Terminal)           ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  /help     - Show this help message")
	fmt.Println("  /user     - Show current user info")
	fmt.Println("  /reset    - Reset/create new user")
	fmt.Println("  /activate - Mark account as active (skip payment)")
	fmt.Println("  /quit     - Exit the application")
	fmt.Println()

	// Get or create test user
	phoneNumber := "+2348159124775" // Default test phone number
	var currentUser models.UserModel

	// Check if user exists
	result := db.Where("phone_number = ?", phoneNumber).First(&currentUser)
	if result.Error != nil {
		// Create new user
		fmt.Println("No existing user found. Creating new user...")
		fmt.Print("Enter your name: ")
		reader := bufio.NewReader(os.Stdin)
		userName, _ := reader.ReadString('\n')
		userName = strings.TrimSpace(userName)
		if userName == "" {
			userName = "User"
		}

		// Process new user through message processor
		if err := messageProcessor.ProcessMessage(phoneNumber, userName, ""); err != nil {
			log.Printf("Error processing new user: %v", err)
		}

		// Reload user
		db.Where("phone_number = ?", phoneNumber).First(&currentUser)
	} else {
		fmt.Printf("\n👋 Welcome back, %s!\n", currentUser.UserName)
		if currentUser.Status != "active" && currentUser.AuthorizationCode == "" {
			fmt.Println("⚠️  You haven't completed payment yet. Use /activate to skip payment for testing.")
		} else if currentUser.Status == "active" {
			fmt.Println("✅ Your account is active and ready to use!")
		}
		fmt.Println()
	}

	// Start chat loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		// Handle commands
		switch message {
		case "/quit", "/exit":
			fmt.Println("\n👋 Goodbye!")
			return

		case "/help":
			fmt.Println("\nCommands:")
			fmt.Println("  /help     - Show this help message")
			fmt.Println("  /user     - Show current user info")
			fmt.Println("  /reset    - Reset/create new user")
			fmt.Println("  /activate - Mark account as active (skip payment for testing)")
			fmt.Println("  /quit     - Exit the application")
			fmt.Println()
			continue

		case "/user":
			// Refresh user data
			db.Where("phone_number = ?", phoneNumber).First(&currentUser)
			fmt.Println("\n" + strings.Repeat("─", 60))
			fmt.Printf("Phone: %s\n", currentUser.PhoneNumber)
			fmt.Printf("Name: %s\n", currentUser.UserName)
			fmt.Printf("Status: %s\n", currentUser.Status)
			fmt.Printf("Onboarding Step: %s\n", currentUser.OnboardingStep)
			fmt.Printf("Payment Status: %s\n", currentUser.OnboardingTxnStatus)
			fmt.Println(strings.Repeat("─", 60) + "\n")
			continue

		case "/reset":
			// Delete and recreate user
			db.Unscoped().Where("phone_number = ?", phoneNumber).Delete(&models.UserModel{})
			fmt.Println("\n🔄 User reset. Please restart the application.\n")
			return

		case "/activate":
			// Activate user for testing
			currentUser.Status = "active"
			currentUser.OnboardingStep = "completed"
			currentUser.OnboardingTxnStatus = "success"
			currentUser.AuthorizationCode = "test_auth_code"
			db.Save(&currentUser)
			fmt.Println("\n✅ Account activated! You can now chat with the bot.\n")
			continue
		}

		// Process message through message processor
		if err := messageProcessor.ProcessMessage(phoneNumber, currentUser.UserName, message); err != nil {
			log.Printf("Error processing message: %v", err)
		}

		// Reload user data in case status changed
		db.Where("phone_number = ?", phoneNumber).First(&currentUser)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}
