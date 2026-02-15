FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/bin/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache curl
COPY --from=build /app/bin/server /app/server
COPY --from=build /app/docs /app/docs
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD curl -fsS http://localhost:8080/healthz || exit 1
ENTRYPOINT ["/app/server"]
