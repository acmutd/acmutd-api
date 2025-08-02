FROM golang:1.24.5-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/scraper/main.go

# Install Chrome dependencies
FROM python:3.11-slim AS chrome

# Install system dependencies
RUN apt-get update && apt-get install -y \
    wget \
    gnupg \
    unzip \
    curl \
    && rm -rf /var/lib/apt/lists/*

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

COPY scripts/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

WORKDIR /app

# Copy only necessary items
COPY scripts/ /app/scripts/
COPY --from=build /app/main /app/main

COPY ${FIREBASE_CONFIG} /app/${FIREBASE_CONFIG}

VOLUME /app/output
CMD ["./main"]