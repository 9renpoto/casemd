# syntax=docker/dockerfile:1

ARG VERSION=dev

FROM golang:1.27-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS build

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
