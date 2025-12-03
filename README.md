# AI Interview Trainer - Backend API

A Go backend API for an AI-powered interview training platform. Originally built for the Australian Public Service, but flexible enough to work for any industry.

Built with [Fiber](https://gofiber.io/) - fast, flexible, and fun to work with.

## What it does

This API powers an interview training platform where users can practice with AI interviewers. It handles:
- Real-time AI conversations (multiple providers)
- Text-to-speech generation (streaming audio)
- Speech-to-text transcription
- User management and credit system
- Authentication via Google OAuth

## Quick Start

```bash
# Clone it
git clone https://github.com/AI-God-Dev/AI_Interview_Trainner_Go_Backend.git
cd AI_Interview_Trainner_Go_Backend

# Set up environment
cp env.example .env
# Edit .env with your config (at minimum: DSN, JWT_SECRET, API_KEY)

# Install deps and run
go mod download
go run .
```

Server runs on `http://localhost:8080` by default.

## Features

### AI Providers
- **Text Generation**: OpenAI (GPT-3.5, GPT-4), Vertex AI (PaLM, Gemini Pro), custom assistants
- **Text-to-Speech**: OpenAI TTS, ElevenLabs, Unreal Speech, Vertex AI
- **Speech-to-Text**: OpenAI Whisper, Vertex AI

### Core Features
- User management with credit/token system
- JWT authentication with Google OAuth
- API key protection for endpoints
- Streaming audio responses
- Multiple AI model support per user
- Session management

## Configuration

### Required Environment Variables

You **must** set these:

- `DSN` - MySQL connection string (e.g., `user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local`)
- `JWT_SECRET` - Secret key for JWT signing (generate a strong random string!)
- `API_KEY` - API key for protecting endpoints (use a strong value)

### Optional but Recommended

- `OPEN_AI_API_KEY` - For OpenAI features
- `ELEVEN_LABS_API_KEY` - For ElevenLabs TTS
- `UNREAL_SPEECH_API_KEY` - For Unreal Speech TTS
- `VERTEX_AI_API_KEY` - For Vertex AI features
- `GCLOUD_API_KEY` - For Google Cloud services (get via `gcloud auth print-access-token`)

See `env.example` for the full list with descriptions.

### Tips

- For local dev, you can use a local MySQL instance or [PlanetScale](https://planetscale.com/) for a free cloud DB
- Generate a strong `JWT_SECRET` with: `openssl rand -hex 32`
- The `API_KEY` is checked on every request - make it long and random
- Set `COOKIE_SECURE=false` for local development over HTTP

## API Documentation

Once the server is running, check out the Swagger docs:

```
http://localhost:8080/swagger/index.html
```

Note: Swagger docs are still being improved, some endpoints might not be fully documented yet.

## Development

### Using Make

```bash
make run              # Start the server
make test             # Run unit tests
make test-integration # Run integration tests
make test-all         # Run all tests
make lint             # Check code quality
make format           # Format code
make build            # Build binary
```

### Testing

The project includes unit tests and integration tests:

```bash
# Run unit tests only
go test -v ./pkg/... ./app/services/...

# Run integration tests (requires test database)
go test -v -tags=integration ./...

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for more details on the testing strategy.

### Without Make

```bash
go run .                    # Run
go test ./...               # Test
golangci-lint run           # Lint
go fmt ./...                # Format
go build -o bin/server .    # Build
```

### Hot Reload

If you have [Air](https://github.com/cosmtrek/air) installed:

```bash
make dev
# or
air
```

## Docker

### Quick Start

```bash
docker-compose up
```

This will start the app with all the config from your `.env` file.

### Manual Build

```bash
docker build -t ai-interview-trainer .
docker run -p 8080:8080 --env-file .env ai-interview-trainer
```

The Dockerfile uses a multi-stage build with a distroless base image for security and minimal size.

## Project Structure

```
app/
  handlers/    # HTTP request handlers
  models/      # Data models (User, AI models, etc.)
  services/    # Business logic (AI service, user service, etc.)
pkg/
  config/      # Configuration management
  errors/      # Custom error types
  logger/      # Structured logging (zap)
  middleware/  # HTTP middleware (auth, logging, recovery, etc.)
  routes/      # Route definitions
platform/
  database/    # Database connection and setup
```

For a detailed architecture overview, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Deployment

### Google Cloud Run

The easiest way to deploy:

```bash
gcloud run deploy --source .
```

Make sure you've set all required environment variables in Cloud Run's configuration.

### Other Platforms

The app should work on any platform that supports Go:
- AWS Lambda (with adapter)
- Heroku
- DigitalOcean App Platform
- Railway
- Fly.io

Just make sure to:
1. Set all required environment variables
2. Expose port 8080 (or change `PORT` env var)
3. Have a MySQL database accessible

## Troubleshooting

### Database Connection Issues

- Check your `DSN` format matches: `user:pass@tcp(host:port)/dbname?params`
- Make sure MySQL is running and accessible
- Check firewall rules if connecting to remote DB

### Authentication Not Working

- Verify `JWT_SECRET` is set and matches what your frontend expects
- Check `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` if using OAuth
- Make sure `API_KEY` matches what your frontend sends in `x-api-key` header

### AI Services Not Responding

- Verify API keys are set correctly
- Check API key permissions/quotas
- For Vertex AI, you might need to run `gcloud auth application-default login` first

## Contributing

Found a bug? Have an idea? PRs welcome!

1. Fork the repo
2. Create a feature branch
3. Make your changes
4. Test it
5. Submit a PR

## License

[Add your license here]

## Related

- Frontend: [ai-interview-trainer-frontend](https://github.com/dcrebbin/ai-interview-trainer-frontend)
- Demo: [YouTube](https://www.youtube.com/watch?v=ef2ivitjiBU)
