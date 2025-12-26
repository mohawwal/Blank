# WhatsApp Bot with Claude AI

A WhatsApp bot powered by Claude AI for selling recharge cards and data bundles.

## Features

- **User Registration**: Automatically registers new users and asks for their name
- **User Status Check**: Verifies if users are registered and active before processing requests
- **Claude AI Integration**: Uses Claude AI to intelligently process user messages
- **WhatsApp Integration**: Uses Twilio for WhatsApp messaging
- **Payment Integration**: Paystack integration for payments (future feature)

## How It Works

1. **New User Flow**:
   - User sends a message to the WhatsApp bot
   - Bot checks if user exists in database
   - If not registered, bot creates a new user record with status "new"
   - Bot asks user for their full name using Claude AI
   - Once user provides name (minimum 3 characters), user status is set to "active"

2. **Active User Flow**:
   - User sends a message
   - Bot verifies user is registered and active
   - Claude AI processes the message and responds intelligently
   - User can request:
     - Purchase recharge cards (MTN, Glo, Airtel, 9Mobile)
     - Buy data bundles
     - Check balance
     - View transaction history

## Setup

### Prerequisites

- Go 1.25.5 or higher
- PostgreSQL database
- Twilio account with WhatsApp enabled
- Anthropic API key (for Claude AI)
- Paystack account (optional, for payments)

### Installation

1. Clone the repository
2. Copy `.env.example` to `.env` and fill in your credentials:

```bash
cp .env.example .env
```

3. Update the `.env` file with your credentials:

```env
APP_PORT=2342
DATABASE_URL=postgresql://username:password@host:port/database
ANTHROPIC_API_KEY=your_anthropic_api_key_here
TWILIO_ACCOUNT_SID=your_twilio_account_sid
TWILIO_AUTH_TOKEN=your_twilio_auth_token
TWILIO_WHATSAPP_NUMBER=whatsapp:+14155238886
PAYSTACK_SECRET_KEY=your_paystack_secret_key
```

4. Install dependencies:

```bash
go mod download
```

5. Run database migrations:

```bash
go run internal/database/migrate_up.go
```

6. Build and run the application:

```bash
go build -o bin/app ./cmd/api
./bin/app
```

## API Endpoints

- `GET /` - Health check endpoint
- `POST /webhook/whatsapp` - WhatsApp webhook endpoint (configured in Twilio)

## Project Structure

```
whatsapp-bot-app/
├── cmd/api/
│   ├── handlers/        # HTTP handlers
│   │   ├── webhook_handler.go
│   │   ├── payment_handler.go
│   │   └── handler.go
│   ├── services/        # Business logic services
│   │   ├── claude_service.go
│   │   ├── twilio_service.go
│   │   └── paystack_service.go
│   ├── middlewares/     # Custom middlewares
│   ├── main.go
│   ├── routes.go
│   └── server.go
├── internal/
│   ├── database/        # Database migrations
│   └── models/          # Data models
│       └── user.model.go
├── common/              # Shared utilities
│   └── connection.go
└── .env                 # Environment variables
```

## User Model

The user model includes:

- `phone_number`: WhatsApp phone number (unique)
- `name`: User's full name
- `email`: User's email
- `status`: User status (new, active, suspended)
- `onboarding_step`: Current onboarding step (awaiting_name, completed)
- Bank and payment information for Paystack integration

## Development

### Testing the Webhook

Use ngrok to expose your local server:

```bash
ngrok http 2342
```

Then configure the ngrok URL in your Twilio WhatsApp sandbox settings:
```
https://your-ngrok-url.ngrok.io/webhook/whatsapp
```

## Next Steps

- Implement actual recharge card purchase logic
- Implement data bundle purchase logic
- Add wallet balance tracking
- Add transaction history
- Implement Paystack payment webhook
- Add more intelligent conversation handling with Claude AI

## License

MIT
# Blank
