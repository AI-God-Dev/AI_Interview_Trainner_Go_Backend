# Setup Guide

A more detailed guide for getting started with the AI Interview Trainer backend.

## Prerequisites

- Go 1.21 or higher ([install](https://go.dev/doc/install))
- MySQL 8.0+ (or compatible database)
- Git

Optional but recommended:
- Docker & Docker Compose (for containerized setup)
- Make (for easier commands)

## Step-by-Step Setup

### 1. Clone the Repository

```bash
git clone https://github.com/AI-God-Dev/AI_Interview_Trainner_Go_Backend.git
cd AI_Interview_Trainner_Go_Backend
```

### 2. Set Up Database

You have a few options:

#### Option A: Local MySQL

```bash
# Install MySQL (varies by OS)
# macOS: brew install mysql
# Ubuntu: sudo apt-get install mysql-server
# Windows: Download from mysql.com

# Start MySQL
mysql.server start  # macOS
sudo systemctl start mysql  # Linux

# Create database
mysql -u root -p
CREATE DATABASE interview_trainer;
```

#### Option B: PlanetScale (Free Cloud MySQL)

1. Sign up at [planetscale.com](https://planetscale.com)
2. Create a new database
3. Get your connection string from the dashboard
4. Use it as your `DSN` (make sure to add `&tls=true`)

#### Option C: Docker MySQL

```bash
docker run --name mysql-dev -e MYSQL_ROOT_PASSWORD=rootpass -e MYSQL_DATABASE=interview_trainer -p 3306:3306 -d mysql:8.0
```

### 3. Configure Environment

```bash
cp env.example .env
```

Edit `.env` and set at minimum:

```env
DSN=root:rootpass@tcp(localhost:3306)/interview_trainer?charset=utf8mb4&parseTime=True&loc=Local
JWT_SECRET=$(openssl rand -hex 32)  # Generate a random secret
API_KEY=$(openssl rand -hex 32)     # Generate a random API key
```

**Important**: Generate real secrets! Don't use the example values.

### 4. Install Dependencies

```bash
go mod download
```

### 5. Run the Application

```bash
go run .
```

You should see:
```
starting app env=development
db connected
server starting addr=0.0.0.0:8080
```

### 6. Verify It's Working

```bash
curl http://localhost:8080/health
```

Should return:
```json
{"status":"ok","time":"2024-01-01T12:00:00Z"}
```

## Getting API Keys

### OpenAI

1. Go to [platform.openai.com](https://platform.openai.com)
2. Sign up/login
3. Go to API Keys section
4. Create a new key
5. Add to `.env` as `OPEN_AI_API_KEY`

### ElevenLabs

1. Sign up at [elevenlabs.io](https://elevenlabs.io)
2. Go to Profile → API Keys
3. Copy your API key
4. Add to `.env` as `ELEVEN_LABS_API_KEY`

### Vertex AI / Google Cloud

This one's a bit more involved:

1. Create a Google Cloud project
2. Enable Vertex AI API
3. Create a service account
4. Download credentials JSON
5. Run: `gcloud auth activate-service-account --key-file=credentials.json`
6. Get token: `gcloud auth print-access-token`
7. Or use the service account JSON directly (requires code changes)

For quick testing, you can use:
```bash
gcloud auth login
gcloud auth print-access-token
# Copy the token to GCLOUD_API_KEY
```

### Unreal Speech

1. Sign up at [unrealspeech.com](https://www.unrealspeech.com)
2. Get your API key from dashboard
3. Add to `.env` as `UNREAL_SPEECH_API_KEY`

## Testing the API

### Health Check

```bash
curl http://localhost:8080/health
```

### With API Key

```bash
curl -H "x-api-key: your-api-key-here" http://localhost:8080/api/users?email=test@example.com
```

## Common Issues

### "failed to connect database"

- Check MySQL is running
- Verify DSN format is correct
- Check username/password
- Make sure database exists

### "JWT_SECRET must be set"

- Make sure `.env` file exists
- Check you've set `JWT_SECRET` in `.env`
- Restart the server after changing `.env`

### Port already in use

Change the port in `.env`:
```env
PORT=8081
```

### CORS errors from frontend

Make sure `ALLOWED_ORIGINS` includes your frontend URL:
```env
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

## Next Steps

- Check out the [API documentation](http://localhost:8080/swagger/index.html)
- Set up your frontend to connect to this API
- Configure your AI provider API keys
- Start building!

