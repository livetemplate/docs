# Multi-stage build for the LiveTemplate docs site.
# Tinkerdown is built from source because `go install ...@latest` fails
# until upstream fixes the vendored asset embed (Phase 0 finding T0-1).
# `TINKERDOWN_REF` selects which branch/tag to clone and is overridable.

ARG TINKERDOWN_REF=v0.1.2

# ---- Stage 1: Build TypeScript client assets for tinkerdown ----
FROM node:20-alpine AS client-builder
ARG TINKERDOWN_REF
RUN apk add --no-cache git
WORKDIR /src
RUN git clone --depth=1 --branch=${TINKERDOWN_REF} https://github.com/livetemplate/tinkerdown.git .
WORKDIR /src/client
RUN npm ci --prefer-offline --no-audit && npm run build

# ---- Stage 2: Build the tinkerdown binary ----
FROM golang:1.26-alpine AS go-builder
ARG TINKERDOWN_REF
RUN apk add --no-cache git ca-certificates
ENV GOTOOLCHAIN=auto
WORKDIR /src
RUN git clone --depth=1 --branch=${TINKERDOWN_REF} https://github.com/livetemplate/tinkerdown.git .
COPY --from=client-builder /src/client/dist/ ./client/dist/
RUN mkdir -p internal/assets/client && \
    cp client/dist/tinkerdown-client.browser.js internal/assets/client/ && \
    cp client/dist/tinkerdown-client.browser.js.map internal/assets/client/ && \
    cp client/dist/tinkerdown-client.browser.css internal/assets/client/ && \
    cp client/dist/tinkerdown-client.browser.css.map internal/assets/client/
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/tinkerdown ./cmd/tinkerdown

# ---- Stage 3: Runtime image — binary + the docs content ----
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 1000 docs
WORKDIR /site
COPY --from=go-builder /out/tinkerdown /usr/local/bin/tinkerdown
COPY content/ /site/
RUN chown -R docs:docs /site
USER docs
EXPOSE 8080
CMD ["tinkerdown", "serve", "--host", "0.0.0.0", "--port", "8080", "/site"]
