FROM golang:1.22.4-bookworm as builder

COPY ${PWD} /app
WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags '-s -w' -trimpath -o main ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot as runner

COPY --from=builder /app/main /home/nonroot/main
COPY --from=builder /app/configs /home/nonroot/configs
WORKDIR /home/nonroot

EXPOSE 8080

ENTRYPOINT [ "./main", "-config" ]
CMD [ "./configs/dev.docker.env" ]
