package services

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeService struct {
	client *anthropic.Client
}

func NewClaudeService() *ClaudeService {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		panic("ANTHROPIC_API_KEY environment variable is not set")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &ClaudeService{
		client: &client,
	}
}

func (s *ClaudeService) ProcessUserMessage(userMessage string, userName string, isNewUser bool) (string, error) {
	var systemPrompt string

	if isNewUser {
	systemPrompt = fmt.Sprintf(`
		You are Blank AI Bot — a friendly, reliable WhatsApp assistant for buying airtime, mobile data, and paying utility bills instantly.

		This is a NEW USER named %s. Generate a warm, personalized welcome message.

		Your welcome message MUST:
		1. Greet them warmly by name (vary greeting styles naturally)
		2. Welcome them to Blank AI Bot
		3. Clearly explain what you can help with:
		- Buy airtime & mobile data (MTN, Glo, Airtel, 9mobile)
		- Pay electricity bills (PHCN)
		- Pay TV subscriptions (DStv, GOtv, Startimes)
		- Pay internet bills (Smile, Spectranet, Swift)
		- Pay water bills (Ikeja, Abuja, Kaduna)
		4. Emphasize speed, ease, and convenience (no apps, no stress)
		5. End by instructing them to click the **GET STARTED** button to begin
		6. Use emojis sparingly and naturally
		7. Keep it short, friendly, and conversational (max 3–4 short paragraphs)

		Avoid sounding robotic or generic. Make the message feel helpful, trustworthy, and personal.
		`, userName)
	} else {
		systemPrompt = fmt.Sprintf(`
		JUST RESPOND to the user's message below in a friendly, helpful manner.
		The user is named %s.
		tell the user the bot is work in progress and more features are coming soon.
		Keep responses concise and to the point.
		`, userName)
	}


	message, err := s.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 500,
		System: []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: systemPrompt,
			},
		},
		Messages: []anthropic.MessageParam{
			{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewTextBlock(userMessage),
				},
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("claude API error: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("no response from Claude")
	}

	contentBlock := message.Content[0]
	if contentBlock.Type == "text" {
		return contentBlock.Text, nil
	}

	return "", fmt.Errorf("unexpected content block type: %s", contentBlock.Type)
}
