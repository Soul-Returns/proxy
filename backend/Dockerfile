# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

COPY backend/frontend/package.json backend/frontend/package-lock.json* ./
RUN npm install

COPY backend/frontend/ ./
RUN npm run build

# Stage 2: Build Host Agent (cross-compiled for Windows and Linux)
FROM golang:1.22-alpine AS agent-builder

WORKDIR /agent

# Copy VERSION file
COPY VERSION /tmp/VERSION

# Extract agent version and set as ARG
RUN AGENT_VERSION=$(grep AGENT_VERSION /tmp/VERSION | cut -d '=' -f 2) && \
    echo "AGENT_VERSION=${AGENT_VERSION}" > /tmp/agent_version.env

# Copy agent source
COPY agent/go.mod ./
COPY agent/*.go ./
COPY agent/config/ ./config/
COPY agent/hosts/ ./hosts/
COPY agent/sync/ ./sync/
COPY agent/autostart/ ./autostart/
COPY agent/gui/ ./gui/
COPY agent/tray/ ./tray/
COPY agent/version/ ./version/

# Download dependencies
RUN go mod tidy && go mod download

# Build for Windows (amd64) with version in filename
RUN AGENT_VERSION=$(grep AGENT_VERSION /tmp/VERSION | cut -d '=' -f 2) && \
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui -X devproxy-agent/version.Current=${AGENT_VERSION}" -o devproxy-agent-v${AGENT_VERSION}.exe . && \
    cp devproxy-agent-v${AGENT_VERSION}.exe devproxy-agent.exe

# Build for Linux (amd64) with version in filename
RUN AGENT_VERSION=$(grep AGENT_VERSION /tmp/VERSION | cut -d '=' -f 2) && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X devproxy-agent/version.Current=${AGENT_VERSION}" -o devproxy-agent-v${AGENT_VERSION} . && \
    cp devproxy-agent-v${AGENT_VERSION} devproxy-agent

# Stage 3: Build Go backend
FROM golang:1.22-alpine AS backend-builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy VERSION file
COPY VERSION /tmp/VERSION

# Copy go mod files
COPY backend/go.mod backend/go.sum* ./

# Copy source code (needed for go mod tidy)
COPY backend/*.go ./
COPY backend/internal/ ./internal/

# Create frontend dist placeholder for embedding
RUN mkdir -p frontend/dist && touch frontend/dist/.gitkeep

# Download dependencies and generate go.sum
RUN go mod tidy && go mod download

# Copy built frontend for embedding
COPY --from=frontend-builder /frontend/dist ./frontend/dist

# Build with CGO for SQLite, embedding version
RUN BACKEND_VERSION=$(grep BACKEND_VERSION /tmp/VERSION | cut -d '=' -f 2) && \
    CGO_ENABLED=1 go build -ldflags="-X devproxy/internal/version.Current=${BACKEND_VERSION}" -o devproxy-api .

# Stage 4: Runtime
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache sqlite-libs ca-certificates

# Copy VERSION file
COPY VERSION /app/VERSION

COPY --from=backend-builder /build/devproxy-api /app/

# Copy agent binaries (both versioned and non-versioned) for download
COPY --from=agent-builder /agent/devproxy-agent*.exe /app/agent/
COPY --from=agent-builder /agent/devproxy-agent* /app/agent/

# Create data directory
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["/app/devproxy-api"]
