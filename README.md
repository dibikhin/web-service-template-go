# ws-dummy-go
Web Service Template in Go

curl -v localhost:8080/createUser \
    -d '{"name":"juwis"}' \
    -H "Content-Type: application/json" \
    -H "X-Request-ID: a1b2c3d4e3f2g1" 

Features:
- Go kit chassis
- Dockerized app and infrastructure
- Request validation
- Config files
- SQL builder
- Postgres connection pool
- DB migrator
- Data layer tests
- Mocks
- Rich Taskfile
- Graceful shutdown

golang-ci
Kafka
Fuzzy tests
Swagger
Caching Decorator
HTTP Client + rate limiter, balancer, circuit braker
