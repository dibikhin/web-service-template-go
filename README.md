# ws-dummy-go
Web Service Template in Go

curl -v localhost:8080/createUser \
    -d '{"name":"juwis"}' \
    -H "Content-Type: application/json" 

Done:
+ docker-compose.yml
+ docker-infra.yml
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

Tests

Error Handling
Validation

Fake Redis
Fake S3
Fuzzy tests

Swagger
Transactions
Caching Decorator

HTTP Client + rate limiter, balancer, circuit braker
