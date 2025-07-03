# Insider Message System

**Insider Message System** is an automatic message sending platform built with Go, Gin, and Swagger. It supports scheduled message delivery, webhook integration, and provides a RESTful API with interactive swagger documentation.

## Features

- **REST API** for sending and retrieving messages
- **Scheduler** for automatic message delivery
- **Webhook** support with circuit breaker
- **Swagger UI** for easy API exploration and testing
- **Docker & Docker Compose** ready

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/)

### Quick Start

1. **Clone the repository:**
   ```sh
   git clone https://github.com/sahmaragaev/insider-message-system
   cd insider-message-system
   ```

2. **Copy and edit environment variables:**
   ```sh
   cp .env.example .env
   # Edit .env as needed
   ```

3. **Build and run with Docker Compose:**
   ```sh
   docker-compose up --build
   ```

4. **Access the API documentation:**
   Swagger docs are auto-generated and available at `/swagger/index.html` when the server is running.

   Open [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) in your browser.

## API Endpoints

- `POST   /api/v1/messages` — Create a new message
- `GET    /api/v1/messages/sent` — List sent messages
- `POST   /api/v1/scheduler/start` — Start the scheduler
- `POST   /api/v1/scheduler/stop` — Stop the scheduler
- `GET    /api/v1/scheduler/status` — Scheduler status
- `GET    /health` — Health check

## Development

To run locally without Docker:

```sh
go run ./cmd/api
```

If you are running application separately, you should execute the command below for swagger docs. It has already been done; thus the swagger Docs are inside docs folder.

```sh
swag init -g cmd/api/main.go -o docs
```

## Running Tests

To run all tests in the project:

```
go test ./...
```

## Additional Info
The webhook.site seemed to have restrictions for returning a variable (or anything other than static response) so I decided to make a simple webhook on my own domain (sahmar.org). I hope it will not cause any problems.