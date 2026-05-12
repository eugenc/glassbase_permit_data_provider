FROM node:20-alpine AS frontend

WORKDIR /src
COPY frontend/package*.json ./frontend/
RUN cd frontend && npm ci
COPY frontend ./frontend
RUN cd frontend && npm run build

FROM golang:1.25-alpine AS builder

WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend /src/internal/static/web ./internal/static/web

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/glassbase ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/scrape-one ./cmd/scrape-one && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/diagnose ./cmd/diagnose && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/onboard ./cmd/onboard && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/cc-repair ./cmd/cc-repair

FROM chromedp/headless-shell:latest AS chrome

FROM alpine:3.19

WORKDIR /app

COPY --from=chrome /headless-shell /headless-shell

RUN apk add --no-cache ca-certificates tzdata nodejs npm

RUN npm install -g @anthropic-ai/claude-code

COPY --from=builder /out/glassbase /app/glassbase
COPY --from=builder /out/scrape-one /app/scrape-one
COPY --from=builder /out/diagnose /app/diagnose
COPY --from=builder /out/onboard /app/onboard
COPY --from=builder /out/cc-repair /app/cc-repair
COPY --from=builder /app/migrations /app/migrations

COPY CLAUDE.md .claude /app/

ENV CHROME_BIN=/headless-shell/headless-shell
ENV GLASSBASE_CLI_WRAPPER=binary

EXPOSE 8080
CMD ["/app/glassbase"]
