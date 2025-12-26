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

//i think we should use the usermessage, username, isactive and is newuser to determine the prompt
func (s *ClaudeService) ProcessUserMessage(userMessage string, userName string, isNewUser bool) (string, error) {
	var systemPrompt string

	if isNewUser {
		systemPrompt = `
		You are a friendly and professional WhatsApp AI assistant for buying airtime, data bundles, electricity, and other value‑added services in Nigeria.

		This user is new and has not completed onboarding.

		Your goal right now is to begin registration.
		Politely ask the user for the information needed to create their account.

		Rules:
		- Be brief and friendly.
		- Ask for one piece of information at a time.
		- Do NOT discuss payments, purchases, or services yet.
		- Use simple, WhatsApp‑style language.
		`
	} else {
		systemPrompt = fmt.Sprintf(`
		You are a friendly and reliable WhatsApp AI assistant for buying airtime, data bundles, electricity, and other value‑added services in Nigeria.

		The user’s name is %s. They are a registered and active customer.

		You can help them with:
		1. Buying airtime (MTN, Glo, Airtel, 9mobile)
		2. Buying data bundles
		3. Paying electricity bills
		4. Checking their balance or account status
		5. Viewing recent transactions

		Rules:
		- Keep replies short, clear, and friendly.
		- Use WhatsApp‑style conversational language.
		- When a user wants to make a purchase:
		- Clearly restate the details (network, amount, phone number).
		- Ask for confirmation before proceeding.
		- Never process a payment without explicit confirmation.
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
