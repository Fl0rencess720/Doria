# Doria

Doria is a microservices-based Go application that provides AI-powered chat and image processing capabilities. The system follows clean architecture principles and uses gRPC for service-to-service communication.

## Table of Contents

- [Architecture](#architecture)
- [Services](#services)
  - [Gateway Service](#gateway-service)
  - [Chat Service](#chat-service)
  - [Image Service](#image-service)
  - [User Service](#user-service)
- [Key Features](#key-features)
- [Technology Stack](#technology-stack)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Running Services](#running-services)
- [Development](#development)
  - [Generating Protobuf Files](#generating-protobuf-files)
- [Configuration](#configuration)
- [API Endpoints](#api-endpoints)

## Architecture

The system follows a clean architecture pattern with clear separation of concerns:

```
Client → Gateway → Chat/Image/User Services
```

Each service is independently deployable and communicates via gRPC. The gateway acts as the entry point for all HTTP requests and routes them to the appropriate backend services.


## Key Features

- **AI-Powered Chat**: Conversational AI with tool integration capabilities
- **Image Processing**: Image analysis and description generation
- **User Management**: Complete user authentication and management system
- **Microservices Architecture**: Independently deployable services
- **gRPC Communication**: Efficient service-to-service communication
- **Observability**: Tracing with OpenTelemetry and profiling with Pyroscope
- **Clean Architecture**: Well-defined separation of concerns
- **Dependency Injection**: Compile-time dependency injection with Google Wire

## Technology Stack

- **Language**: Go
- **Communication**: gRPC, Protocol Buffers
- **Database**: PostgreSQL, Redis
- **Observability**: OpenTelemetry, Pyroscope, Langfuse
- **AI Framework**: Eino framework for agent composition
- **Service Discovery**: Consul
- **Vector Database**: Milvus

## Getting Started

### Prerequisites

- Go 1.24.6+
- PostgreSQL
- Redis
- Consul
- Docker (recommended for dependencies)

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd Doria
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running Services

Use the provided Makefile commands to run services:

```bash
# Run gateway service
make run gateway

# Run chat service
make run chat

# Run image service
make run image

# Run user service
make run user
```

Services will be available at:
- Gateway: http://localhost:8000
- Chat service: localhost:9000
- Image service: localhost:9001
- User service: localhost:9002

## Development

### Generating Protobuf Files

To regenerate protobuf files after modifying the IDL files:

```bash
make proto
```

## Configuration

Each service has its own configuration file located in `configs/config.yaml` within the service directory. Key configuration options include:

- Server settings (port, timeout)
- Database connections
- Tracing and profiling settings
- AI model configurations
- Service discovery settings
