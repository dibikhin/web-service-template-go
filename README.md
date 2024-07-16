# ws-dummy-go
Web Service Template in Go

curl -v localhost:8080/createUser \
    -d '{"name":"juwis"}' \
    -H "Content-Type: application/json" \
    -H "X-Request-ID: d3s32a1a1a" 

Done:
+ docker-compose.yaml
+ docker-infra.yaml
+ Dockerfile
+ Taskfile
+ Graceful shutdown
+ Configs
+ Mongo
+ Redis
+ Postgres + query builder
+ Recovery

golang-ci
Kafka

Error Handling
Validation
Fuzzy tests

Swagger

Caching Decorator

HTTP Client + rate limiter, balancer, circuit braker
