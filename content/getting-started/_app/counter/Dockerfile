# Tutorial counter for the LiveTemplate "Your First App" walkthrough.
# Builds the same source the docs page slices via include="..." blocks.
# Deployed to lt-firstapp.fly.dev; embedded into the docs site via
# embed-lvt blocks pointing at upstream="https://lt-firstapp.fly.dev".

FROM golang:1.26-alpine AS build
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /firstapp-counter .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /firstapp-counter /firstapp-counter
COPY --from=build /app/counter.tmpl /counter.tmpl
EXPOSE 9090
CMD ["/firstapp-counter"]
