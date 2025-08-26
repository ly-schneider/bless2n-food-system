# Bless2n Food System Backend

A Go-based HTTP driven backend service for the Bless2n Food System.

## Features

- RESTful API with Chi router
- MongoDB database
- Docker support for development
- Live-reload for development

## Development Setup

### Local Development

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd bless2n-food-system/backend
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Start services locally**:
   ```bash
   # Run docker
   make docker-up
   
   # Then run API with live-reload
   make dev
   ```

### Available Commands

Show all available commands

```bash
make help
```

## Entity Relationship Diagram

The database schema is visualized in `mermaidchart-erd.txt`, showing relationships between:
- Users and Roles
- Events and Participants 
- Products and Categories
- Orders and Items
- Devices and Pairings