# syntax=docker/dockerfile:1

ARG VERSION=dev

FROM golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS build

ARG VERSION

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/casemd \
    ./cmd/casemd

FROM scratch

ARG VERSION

LABEL org.opencontainers.image.title="casemd" \
      org.opencontainers.image.description="Convert structured Markdown inspection checklists" \
      org.opencontainers.image.source="https://github.com/9renpoto/casemd" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="$VERSION"

WORKDIR /app
COPY --from=build /out/casemd /app/casemd
COPY --from=build /src/notes.md /app/notes.md
COPY --from=build /src/internal/interfaces/web/templates /app/internal/interfaces/web/templates
COPY --from=build /src/internal/interfaces/web/static /app/internal/interfaces/web/static

USER 65532:65532
EXPOSE 3000
ENTRYPOINT ["/app/casemd", "serve"]
