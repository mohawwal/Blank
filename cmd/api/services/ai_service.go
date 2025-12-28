package services

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type AIService struct {
	client *openai.Client
}

func NewAIService() *AIService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(apiKey)

	return &AIService{
		client: client,
	}
}

func (s *AIService) ProcessUserMessage(userMessage string, userName string, isNewUser bool) (string, error) {
	var systemPrompt string

	if isNewUser {
		systemPrompt = fmt.Sprintf(`You are Blank AI Bot — a friendly, reliable WhatsApp assistant for buying airtime, mobile data, and paying utility bills instantly.

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

Avoid sounding robotic or generic. Make the message feel helpful, trustworthy, and personal.`, userName)
	} else {
		systemPrompt = fmt.Sprintf(`JUST RESPOND to the user's message below in a friendly, helpful manner.
The user is named %s.
Tell the user the bot is work in progress and more features are coming soon.
Keep responses concise and to the point.`, userName)
	}

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT4oMini,
			MaxTokens: 500,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
