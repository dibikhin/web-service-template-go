# ws-dummy-go
Web Service Template in Go

curl -v localhost:8080/createUser \
    -d '{"name":"juwis"}' \
    -H "Content-Type: application/json" \
    -H "X-Request-ID: d3s32a1a1a" 

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
- Graciful shutdown

golang-ci
Kafka
Fuzzy tests
Swagger
Caching Decorator
HTTP Client + rate limiter, balancer, circuit braker
