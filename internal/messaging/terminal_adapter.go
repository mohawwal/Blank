package messaging

import (
	"fmt"
	"strings"
)

// TerminalAdapter implements MessagingService for terminal output
type TerminalAdapter struct{}

// NewTerminalAdapter creates a new terminal messaging adapter
func NewTerminalAdapter() *TerminalAdapter {
	return &TerminalAdapter{}
}

// SendMessage prints message to terminal
func (t *TerminalAdapter) SendMessage(to string, body string) error {
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("BOT:", body)
	fmt.Println(strings.Repeat("─", 60) + "\n")
	return nil
}

// SendTemplateMessage prints template message to terminal
func (t *TerminalAdapter) SendTemplateMessage(to string, templateData TemplateData) error {
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Printf("BOT (Template: %s):\n", templateData.TemplateName)

	// For terminal, just print the body variable
	if body, ok := templateData.Variables["1"]; ok {
		fmt.Println(body)
	}

	// Print button/link if available
	if link, ok := templateData.Variables["2"]; ok {
		fmt.Printf("\n🔗 Link: https://checkout.paystack.com/%s\n", link)
	}

	fmt.Println(strings.Repeat("─", 60) + "\n")
	return nil
}
