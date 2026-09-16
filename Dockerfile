# syntax=docker/dockerfile:1

FROM golang:1.27-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

ARG VERSION=dev

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/casemd \
    ./cmd/casemd

FROM scratch

ARG VERSION=dev

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
