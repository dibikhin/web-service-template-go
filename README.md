# ws-dummy-go
Web Service Template in Go

## Features
- Go kit chassis
- Dockerized app and infrastructure
- Postgres, MongoDB, Redis
- Config files
- SQL builder
- DB migrator
- Data layer tests
- Mocks
- Rich Taskfile
- Request validation
- Graceful shutdown

## Run

`task run`

## cURL

```
curl -v localhost:8080/createUser \
    -d '{"name":"juwis"}' \
    -H "Content-Type: application/json" \
    -H "X-Request-ID: a1b2c3d4e3f2g1"

curl -v localhost:8080/updateUser \
    -d '{"userId":"5718336260054177893","name":"jella"}' \
    -H "Content-Type: application/json" \
    -H "X-Request-ID: a1b2c3d4e3f2g1"
```