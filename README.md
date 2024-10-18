## Hacktoberfest 2024 ![Hacktoberfest](https://img.shields.io/badge/Hacktoberfest-2024-blueviolet?style=for-the-badge)

Ccontributions from all levels of experience are welcome! To contribute:

- Fork the repository
- Make your changes
- Open a pull request (PR)

Be sure to read the contribution guidelines and code of conduct before submitting your PR.

# Simple Go Load Balancer

This project demonstrates a basic load balancer implemented in Go. It only uses the standard http library.

It uses Docker and Docker Compose to spin up multiple backend HTTP servers and a load balancer that distributes incoming requests among these backends in round-robin fashion.

## Features

- **Round-Robin Load Balancing**: Distributes incoming requests evenly across backend servers.
- **Health Check**: Detects and excludes unhealthy backend servers from the pool.
- **Retry Mechanism**: Retries failed requests with different backend servers.
- **Dockerized**: Easy to deploy and manage using Docker and Docker Compose.

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://golang.org/) (if you want to run the code locally)

### Running the Load Balancer

1. **Build and start the containers:**

   ```bash
   docker compose up --build
   ```

2. **Test the Load Balancer:**

   ```bash
   curl http://localhost:8080
   ```

   You should see responses from different backend servers (e.g. Hello from backend running on port 80)

3. **Configuring health check**
   Set enviornment variable: `export HEALTH_CHECK_INTERVAL=60` (macOS/Linux), `set HEALTH_CHECK_INTERVAL=60` (Windows Command Prompt)

Inspired by [build-your-own-x](https://github.com/codecrafters-io/build-your-own-x) repo
