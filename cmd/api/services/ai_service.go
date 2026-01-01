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
		systemPrompt = fmt.Sprintf(`You are Blank AI Bot — a friendly WhatsApp assistant for airtime, data, and bill payments.
		Generate a SHORT welcome for %s (new user).

		Requirements:
		- Greet them by name warmly
		- Welcome to Blank AI Bot in ONE sentence
		- List services briefly: airtime, data, electricity, TV bills
		- Emphasize speed, ease, and convenience (no apps, no stress)
		- Maximum 3 short sentences total
		- Use 1-2 emojis only
		- Keep ENTIRE message under 60 words, no emdashes

		Be warm but VERY brief.`, userName)
	} else {
		systemPrompt = fmt.Sprintf(`JUST RESPOND to the user's message below in a friendly, helpful manner.
		The user is named %s.

		Tell the user what we do currently, -
		- Buy airtime & mobile data (MTN, Glo, Airtel, 9mobile)
		- Pay electricity bills (PHCN)
		- Pay TV subscriptions (DStv, GOtv, Startimes)
		- Pay internet bills (Smile, Spectranet, Swift)
		- Pay water bills (Ikeja, Abuja, Kaduna)
		4. Emphasize speed, ease, and convenience (no apps, no stress)
   		- The bot is a work in progress and more features are coming soon.
		Keep responses concise and to the point.`, userName)
	}

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT4oMini,
			MaxTokens: 200,
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
