# 3L1 — Go Microservices Platform

[![Go Version](https://img.shields.io/badge/Go-1.24.5-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](https://opensource.org/licenses/MIT)
[![HTTP API](https://img.shields.io/badge/HTTP-API-%23007EC6?style=flat-square)](https://developer.mozilla.org/en-US/docs/Web/HTTP)
[![gRPC](https://img.shields.io/badge/gRPC-Ready-%23007EC6?style=flat-square&logo=google)](https://grpc.io/)
[![JWT Auth](https://img.shields.io/badge/Auth-JWT%20%2B%20Sessions-green?style=flat-square)](https://jwt.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791?style=flat-square&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-Sessions%20%26%20RateLimit-DC382D?style=flat-square&logo=redis)](https://redis.io/)
[![Prometheus](https://img.shields.io/badge/Monitoring-Prometheus-orange?style=flat-square&logo=prometheus)](https://prometheus.io/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)

Production-oriented microservices platform written in Go, designed to practice real-world backend architecture, networking and reliability patterns.

The system is built around an API Gateway and multiple gRPC services, implementing user management and order processing with hybrid authentication, observability and fault tolerance in mind.

---

## Architecture Overview

The platform follows a classic microservice architecture:

- **API Gateway**
  - Single entry point for HTTP clients
  - Handles authentication, rate limiting and request routing
- **User Service (gRPC)**
  - User registration, authentication and session management
- **Order Service (gRPC)**
  - Order creation, retrieval and lifecycle management

---

## Key Features

### API Gateway
- Hybrid authentication: **JWT + cookie-based sessions**
- JWT validation with automatic refresh via session key
- Redis-based rate limiting
- Circuit breaker for downstream gRPC services
- Prometheus metrics
- Graceful shutdown
- CORS configuration
- Gzip response compression
- Request throttling

### Authentication Flow (High Level)
1. Incoming request is authenticated using JWT
2. If JWT is expired, session key (cookie) is validated via gRPC
3. Session and JWT are reissued transparently

### User Service
- User registration and login
- Secure password hashing (bcrypt)
- JWT issuance and validation
- Session storage in Redis
- User deletion with access checks

### Order Service
- Order creation
- Order lookup
- Order deletion
- Order status management (processing / done / cancelled)

---

## Tech Stack

- **Language**: Go
- **Protocols**: HTTP, gRPC
- **Databases**: PostgreSQL, Redis
- **Auth**: JWT + server-side sessions (cookies)
- **Observability**: Prometheus, Grafana
- **Logging**: Uber Zap
- **Routing**: Chi
- **Circuit Breaker**: Sony gobreaker
- **Containerization**: Docker, Docker Compose

---

### Services
- API Gateway: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000
- pprof: http://localhost:6060/debug/pprof

---

## API Endpoints

### Users
POST   /api/users/reg   — register  
POST   /api/users/log   — login  
POST   /api/users/ext   — extract data from token  
DELETE /api/users/del   — delete user  

### Orders
POST   /api/orders/add  — create order  
GET    /api/orders/info — get order info  
DELETE /api/orders/del  — delete order  

---

## Development Notes

- gRPC is used for internal service communication
- API Gateway acts as a boundary for auth, rate limiting and observability
- Redis is used for both rate limiting and session storage
- Emphasis is placed on clean shutdowns and failure isolation

---

## Purpose

This project was built to practice:
- Microservice architecture
- Authentication strategies (stateless + stateful)
- gRPC communication
- Reliability patterns
- Observability and monitoring
- Backend system design in Go

---

## Licenses

- **License:** This project is licensed under  [MIT](LICENSE)
- **Third-party Licenses:** The full license texts are available in the  [licenses/](licenses/)
