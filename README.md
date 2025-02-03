# Load Balancer

A simple load balancer implementation in Go designed to distribute incoming HTTP traffic across multiple backend servers using configurable algorithms. Supports Round Robin and Least Connections strategies.

![Go](https://img.shields.io/badge/Go-1.22+-blue.svg)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Features

- **Round Robin & Least Connections (TODO) ** algorithms
- Configurable server pools
- Health checks for backend servers
- Request logging
- Scalable architecture for future enhancements

## Installation

1. Ensure [Go 1.22+](https://golang.org/dl/) is installed
2. Clone repository:
   ```bash
   git clone https://github.com/vinit-chauhan/load-balancer.git
   cd load-balancer
   ```
3. Build & run:
   ```bash
   go build -o load-balancer
   ./load-balancer -config=config.json
   ```

## Configuration

Modify `config.yaml`:
```yaml
services:
  - name: web-app-1
    endpoint: entertainment
    urls:
      - localhost:8900
      - localhost:8901
```
- **urls**: List of backend servers
- **endpoint**: endpoint to host the services
- **name**: name for the service

## Usage

Start load balancer:
```bash
./load-balancer -config=config.yaml
```

Send requests:
```bash
curl http://localhost:8080
```

## Logging

Requests are logged to `logs/load_balancer_<timestamp>.log` with format:
```
[YYYY-MM-DD HH:MM:SS] <client_ip> -> <backend_server>
```

## Testing

1. Start backend servers on ports 8900 and 8901
2. Use load testing tool:
   ```bash
   hey -n 1000 -c 50 http://localhost:8000
   ```

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Submit a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file
