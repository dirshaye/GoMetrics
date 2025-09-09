# GoMetrics – Real-Time Cloud Resource Monitoring API

GoMetrics is a lightweight cloud-native monitoring service built with **Go**.  
It collects **CPU, RAM, Disk, and Network** metrics from servers and exposes them via **REST** and **WebSocket APIs**.  

The project is designed to demonstrate:
- Go concurrency (goroutines + channels)
- API-driven backend development
- Real-time streaming with WebSockets
- Containerization with Docker
- Deployment to Kubernetes with autoscaling

---

## Features
- Collects server metrics using [gopsutil](https://github.com/shirou/gopsutil)
- REST API endpoints for current system stats
- WebSocket endpoint for real-time streaming
- Concurrent metric collection with goroutines
- Containerized with Docker
---

## Tech Stack
- **Go** (backend, concurrency, API)
- **Gorilla WebSocket** (real-time communication)
- **gopsutil** (system metrics)
- **Docker** (cloud-native deployment)

---

##  Getting Started

Clone the repo and run locally:

```bash
git clone https://github.com/<your-username>/GoMetrics.git
cd GoMetrics

go mod tidy
go run main.go