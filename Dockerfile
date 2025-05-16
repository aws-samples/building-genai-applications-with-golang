# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.21.5 AS build-stage

# Update packages and specifically update OpenSSL to address CVE-2023-5363
# Adding a timestamp to force rebuild: 2025-05-16 14:42:00
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y --no-install-recommends openssl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY bedrock ./bedrock

RUN CGO_ENABLED=0 GOOS=linux go build -o /genaiapp

# Run the tests in the container
FROM build-stage AS run-test-stage
# Add test commands here if needed
# RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /genaiapp /genaiapp
COPY static ./static

EXPOSE 3000

USER nonroot:nonroot

ENTRYPOINT ["/genaiapp"]
