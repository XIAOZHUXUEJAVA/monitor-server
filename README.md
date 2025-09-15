# Monitor Server

A Go-based system monitoring server that provides REST API endpoints for system resource monitoring.

## Features

- **CPU Monitoring**: Usage, frequency, temperature, and core information
- **Memory Monitoring**: RAM and swap usage statistics
- **Disk Monitoring**: Storage usage across multiple partitions
- **Network Monitoring**: Interface statistics and traffic data
- **System Information**: OS details, uptime, and load averages
- **Process Monitoring**: Running processes with resource usage

## Architecture

This project follows Go best practices with a clean architecture:

```
monitor-server/
├── cmd/server/          # Application entry points
├── internal/            # Private application code
│   ├── api/            # HTTP routing and API setup
│   ├── config/         # Configuration management
│   ├── handler/        # HTTP request handlers
│   ├── middleware/     # HTTP middleware
│   ├── model/          # Data models
│   ├── service/        # Business logic
│   └── repository/     # Data access layer (future)
├── pkg/                # Public packages
│   ├── logger/         # Logging utilities
│   └── response/       # HTTP response utilities
├── configs/            # Configuration files
└── scripts/            # Build and deployment scripts
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile commands)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd monitor-server
```

2. Download dependencies:
```bash
make deps
# or
go mod tidy
```

3. Run the server:
```bash
make dev
# or
go run cmd/server/main.go
```

The server will start on `http://localhost:9000` by default.

### Configuration

Configuration can be set via:
- YAML file: `configs/config.yaml`
- Environment variables (using `_` separator, e.g., `SERVER_PORT=9000`)

### API Endpoints

- `GET /health` - Health check
- `GET /api/cpu` - CPU monitoring data
- `GET /api/memory` - Memory usage data
- `GET /api/disk` - Disk usage data
- `GET /api/network` - Network statistics
- `GET /api/system` - System information
- `GET /api/processes` - Process list (supports `?limit=10&sort=cpu`)

### Development Commands

```bash
make dev          # Run in development mode
make build        # Build binary
make test         # Run tests
make fmt          # Format code
make lint         # Run linter
make clean        # Clean build artifacts
```

### Docker

```bash
make docker-build # Build Docker image
make docker-run   # Run in container
```

## TODO

This is a skeleton project. The following needs to be implemented:

1. **Service Layer**: Implement actual system monitoring using `gopsutil`
2. **Historical Data**: Add in-memory circular buffer for historical metrics
3. **Background Collection**: Implement goroutines for periodic data collection
4. **Error Handling**: Add proper error handling and recovery
5. **Testing**: Add unit and integration tests
6. **Metrics**: Add Prometheus metrics support
7. **Rate Limiting**: Add API rate limiting
8. **Authentication**: Add API authentication if needed

## Dependencies

- **Gin**: HTTP web framework
- **gopsutil**: System and process utilities (to be implemented)
- **Viper**: Configuration management
- **Zap**: Structured logging