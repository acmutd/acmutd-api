FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    wget \
    gnupg \
    unzip \
    curl \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://dl-ssl.google.com/linux/linux_signing_key.pub -O /tmp/google.pub \
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
    && ln -s /usr/bin/chromedriver /usr/local/bin/chromedriver;

ENV CHROME_BIN=/usr/bin/google-chrome
ENV CHROME_PATH=/usr/bin/google-chrome

# Create app directory
WORKDIR /app

# Copy requirements first for better caching
COPY scripts/requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the scripts
COPY scripts/ .

# Create a non-root user and set up chromedriver permissions
RUN useradd -m -u 1000 scraper && chown -R scraper:scraper /app \
    && chown scraper:scraper /usr/local/bin/chromedriver \
    && chmod +x /usr/local/bin/chromedriver
USER scraper

# Set environment variables
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# Default command
CMD ["python", "main.py"]