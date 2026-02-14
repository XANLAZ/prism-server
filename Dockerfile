FROM golang:1.21.13 AS builder
WORKDIR /app

# Copy dependency files first for caching
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download

# Copy build script
COPY build.sh ./

# Copy rest of the code
COPY . .

# Build
RUN ./build.sh

FROM ubuntu:latest
RUN apt update -y && apt install -y ffmpeg psmisc && apt-get clean
WORKDIR /app
COPY --from=builder /app/teamgramd/ /app/
RUN chmod +x /app/docker/entrypoint.sh
ENTRYPOINT /app/docker/entrypoint.sh
