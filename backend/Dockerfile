# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.22-alpine AS backend-builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod file
COPY go.mod ./

# Copy source code first (needed for go mod tidy)
COPY *.go ./

# Create frontend dist placeholder for embedding
RUN mkdir -p frontend/dist && touch frontend/dist/.gitkeep

# Download dependencies and generate go.sum
RUN go mod tidy && go mod download

# Copy built frontend for embedding
COPY --from=frontend-builder /frontend/dist ./frontend/dist

# Build with CGO for SQLite
RUN CGO_ENABLED=1 go build -o devproxy-api .

# Stage 3: Runtime
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache sqlite-libs ca-certificates

COPY --from=backend-builder /build/devproxy-api /app/

# Create data directory
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["/app/devproxy-api"]
