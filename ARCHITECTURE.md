# Architecture Overview

This document provides an overview of the AI Interview Trainer backend architecture.

## High-Level Architecture

```
┌─────────────┐
│   Client    │
│  (Frontend) │
└──────┬──────┘
       │ HTTP/HTTPS
       │
┌──────▼─────────────────────────────────────┐
│         Fiber HTTP Server                  │
│  ┌──────────────────────────────────────┐  │
│  │  Middleware Layer                    │  │
│  │  - Request ID                        │  │
│  │  - Recovery                          │  │
│  │  - Logging                           │  │
│  │  - API Key Auth                      │  │
│  │  - CORS                              │  │
│  └──────────────────────────────────────┘  │
│  ┌──────────────────────────────────────┐  │
│  │  Route Handlers                      │  │
│  │  - User Handler                      │  │
│  │  - AI Handler                        │  │
│  │  - Debug Handler                     │  │
│  └──────────────────────────────────────┘  │
└──────┬─────────────────────────────────────┘
       │
┌──────▼─────────────────────────────────────┐
│         Service Layer                      │
│  ┌──────────────────────────────────────┐  │
│  │  User Service                        │  │
│  │  - User CRUD                         │  │
│  │  - Credit Management                 │  │
│  └──────────────────────────────────────┘  │
│  ┌──────────────────────────────────────┐  │
│  │  AI Service                          │  │
│  │  - Text Generation                   │  │
│  │  - Text-to-Speech                    │  │
│  │  - Speech-to-Text                    │  │
│  └──────────────────────────────────────┘  │
└──────┬─────────────────────────────────────┘
       │
┌──────▼─────────────────────────────────────┐
│         Data Layer                         │
│  ┌──────────────────────────────────────┐  │
│  │  GORM (ORM)                          │  │
│  │  - User Model                        │  │
│  │  - UserSettings Model                │  │
│  └──────────────────────────────────────┘  │
│  ┌──────────────────────────────────────┐  │
│  │  MySQL Database                      │  │
│  └──────────────────────────────────────┘  │
└────────────────────────────────────────────┘

┌────────────────────────────────────────────┐
│  External Services                         │
│  - OpenAI API                              │
│  - Vertex AI                               │
│  - ElevenLabs                              │
│  - Unreal Speech                           │
└────────────────────────────────────────────┘
```

## Component Details

### 1. HTTP Layer (Fiber)

The application uses [Fiber](https://gofiber.io/), a fast HTTP framework for Go.

**Key Features:**
- Fast request handling
- Middleware support
- Built-in JSON parsing
- Streaming support for audio

### 2. Middleware Stack

Middleware is applied in order:

1. **Recovery** - Catches panics and logs them
2. **Request ID** - Adds unique ID to each request for tracing
3. **Logging** - Structured logging with Zap
4. **API Key Auth** - Validates API key on protected routes
5. **CORS** - Handles cross-origin requests
6. **Error Handler** - Centralized error response formatting

### 3. Route Handlers

Handlers are thin - they parse requests, call services, and format responses.

**User Handler:**
- User CRUD operations
- Settings management
- Token/credit operations

**AI Handler:**
- Message processing
- Audio generation
- Speech transcription

### 4. Service Layer

Services contain business logic and coordinate between handlers and data layer.

**User Service:**
- Manages user data
- Handles credit/token system
- User settings management

**AI Service:**
- Integrates with multiple AI providers
- Handles model selection based on user preferences
- Manages API calls to external services

### 5. Data Layer

**Models:**
- `User` - User account information
- `UserSettings` - User preferences (AI models, etc.)

**Database:**
- MySQL (via GORM)
- Connection pooling configured
- Auto-migrations on startup

## Request Flow Example

### Creating a User

```
1. Client → POST /api/users
2. Middleware → API Key validation
3. Handler → Parse request body
4. Service → Validate & create user in DB
5. Service → Return created user
6. Handler → Format JSON response
7. Client ← 200 OK with user data
```

### AI Message Request

```
1. Client → POST /api/ai/message?email=user@example.com
2. Middleware → API Key validation
3. Handler → Parse message from body
4. Service → Get user by email
5. Service → Check user credits
6. Service → Select AI model based on user settings
7. Service → Call external AI API (OpenAI/Vertex/etc.)
8. Service → Decrease user credits
9. Handler → Return AI response
10. Client ← 200 OK with AI message
```

## Configuration Management

Configuration is loaded from environment variables via `pkg/config`:

- Server settings (port, timeouts)
- Database connection
- API keys for external services
- Authentication secrets
- CORS settings

All critical config is validated on startup.

## Error Handling

Errors flow through layers:

1. **Service Layer** - Returns errors (no panics)
2. **Handler Layer** - Converts errors to HTTP responses
3. **Error Middleware** - Catches unhandled errors and formats responses

Error types:
- `AppError` - Custom application errors with HTTP status codes
- Standard Go errors - Wrapped in AppError when needed

## Concurrency

- Fiber handles concurrent requests automatically
- Database connections are pooled (GORM manages this)
- AI service uses goroutines for parallel audio chunk processing
- No shared mutable state (stateless services)

## Security

- API key authentication on all protected routes
- JWT validation for user authentication
- CORS protection
- Secure session cookies
- Input validation
- SQL injection protection (via GORM parameterized queries)

## Logging

Structured logging with Zap:
- Request/response logging
- Error logging with context
- Performance metrics
- Request ID tracking

## Testing Strategy

- **Unit Tests** - Test individual functions/services
- **Integration Tests** - Test API endpoints with test database
- **Smoke Tests** - Basic health checks

Tests use:
- In-memory SQLite for database tests
- HTTP test client for API tests
- Mock external services where possible

## Deployment

The application is designed to be:
- Stateless (can scale horizontally)
- Containerized (Docker)
- Cloud-ready (works on GCP, AWS, etc.)

Key considerations:
- Environment variables for configuration
- Health check endpoint
- Graceful shutdown
- Connection pooling

## Future Improvements

- Add caching layer (Redis) for frequently accessed data
- Implement rate limiting
- Add metrics/monitoring (Prometheus)
- WebSocket support for real-time features
- Background job processing for long-running tasks

