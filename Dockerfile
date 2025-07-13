# Build stage for Go
FROM golang:1.24.2-alpine AS go-builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage with both Go and Python
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    wget \
    gnupg \
    unzip \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Chrome and ChromeDriver based on architecture
RUN if [ "$(uname -m)" = "x86_64" ]; then \
    wget https://dl-ssl.google.com/linux/linux_signing_key.pub -O /tmp/google.pub \
    && gpg --no-default-keyring --keyring /etc/apt/keyrings/google-chrome.gpg --import /tmp/google.pub \
    && echo 'deb [arch=amd64 signed-by=/etc/apt/keyrings/google-chrome.gpg] http://dl.google.com/linux/chrome/deb/ stable main' | tee /etc/apt/sources.list.d/google-chrome.list \
    && apt-get update \
    && apt-get install -y google-chrome-stable \
    && CHROME_VERSION=$(google-chrome --version | grep -oE "[0-9]+\.[0-9]+\.[0-9]+") \
    && wget -O /tmp/chromedriver.zip https://chromedriver.storage.googleapis.com/LATEST_RELEASE_${CHROME_VERSION%%.*} \
    && wget -O /tmp/chromedriver.zip https://chromedriver.storage.googleapis.com/$(cat /tmp/chromedriver.zip)/chromedriver_linux64.zip \
    && unzip /tmp/chromedriver.zip -d /usr/local/bin/ \
    && chmod +x /usr/local/bin/chromedriver \
    && rm /tmp/chromedriver.zip; \
    elif [ "$(uname -m)" = "aarch64" ]; then \
    apt-get update \
    && apt-get install -y chromium chromium-driver \
    && ln -s /usr/bin/chromium /usr/bin/google-chrome \
    && ln -s /usr/bin/chromedriver /usr/local/bin/chromedriver; \
    fi

ENV CHROME_BIN=/usr/bin/google-chrome
ENV CHROME_PATH=/usr/bin/google-chrome

# Create app directory
WORKDIR /app

# Copy Python requirements and install dependencies
COPY scraper/scripts/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy Python scripts
COPY scraper/scripts/ ./scripts/

# Copy Go binary from builder stage
COPY --from=go-builder /app/main ./main

# Create output directory and set up permissions
RUN mkdir -p /app/output

# Create a non-root user and set up permissions
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app \
    && chown appuser:appuser /usr/local/bin/chromedriver \
    && chmod +x /usr/local/bin/chromedriver
USER appuser

# Set environment variables
ENV PYTHONPATH=/app/scripts
ENV PYTHONUNBUFFERED=1
ENV DOCKER_CONTAINER=true

# Expose port
EXPOSE 8080

# Run the Go application
CMD ["./main"]