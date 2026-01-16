# Goggles - Website Health Checker

Website health checker that monitors multiple websites and sends Discord notifications when sites are down.

## Quick Start

1. **Configure:**

   - Edit `config.yaml` with your Discord webhook URL
   - Edit `urls.csv` with websites to monitor

2. **Run:**

   ```bash
   # Using Docker (recommended)
   docker-compose up -d

   # Or using Go directly
   go run cmd/main.go
   ```

## Configuration

### config.yaml

```yaml
discord_webhook_url: https://discord.com/api/webhooks/YOUR_WEBHOOK_URL
timezone: Asia/Tokyo
start: 2025/6/3 21:07:25
end: 2025/6/9 21:07:25
threads: 4
mentions:
  - 12345 # Discord user ID
```

### urls.csv

```csv
URL, CHECK_WORDS
https://www.google.com, Google
https://example.com, example,test
```

## Usage

```bash
# Build and run
make build
./goggles

# Verbose mode (saves HTML logs when checks fail)
./goggles -v

# Single URL check
./goggles -s https://example.com

# Scheduler mode
docker-compose up -d scheduler 
```

## Available Commands

- `make build` - Build the application
- `make run` - Run the application
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make deps` - Install dependencies

