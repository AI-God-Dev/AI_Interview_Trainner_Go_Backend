# AI Interview Trainer - Backend API

Go backend for an AI-powered interview training platform. Built with Fiber.

## Features

- Multiple AI providers: OpenAI, Vertex AI, ElevenLabs, Unreal Speech
- Text generation: GPT-3.5, GPT-4, Gemini Pro, custom assistants
- Text-to-Speech: Multiple TTS providers with streaming
- Speech-to-Text: OpenAI Whisper and Vertex AI
- User management with credit system
- JWT auth with Google OAuth
- API key protection

## Setup

1. Install Go 1.21+
2. Copy `env.example` to `.env` and fill in your values
3. Run `go mod download`
4. Run `go run .`

The server will start on `http://localhost:8080`

## Configuration

Required env vars:
- `DSN` - Database connection string
- `JWT_SECRET` - Secret for JWT signing (use a strong value!)
- `API_KEY` - API key for endpoint protection

See `env.example` for all options.

## API Docs

Swagger docs available at `/swagger/index.html` when running.

## Docker

```bash
docker-compose up
```

Or build manually:
```bash
docker build -t ai-interview-trainer .
docker run -p 8080:8080 --env-file .env ai-interview-trainer
```

## Project Structure

```
app/
  handlers/    # HTTP handlers
  models/      # Data models
  services/    # Business logic
pkg/
  config/      # Config management
  errors/      # Error handling
  logger/      # Logging
  middleware/  # HTTP middleware
  routes/      # Route definitions
platform/
  database/    # DB connection
```

## Development

```bash
make run       # Run the app
make test      # Run tests
make lint      # Run linters
make format    # Format code
```

## Deployment

For Google Cloud Run:
```bash
gcloud run deploy --source .
```

Make sure all env vars are set in your deployment environment.

## License

[Add your license]
