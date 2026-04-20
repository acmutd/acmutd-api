# ACM UTD API

A comprehensive REST API for accessing University of Texas at Dallas course data, including course information, grade distributions, and professor ratings. Built with Go, Firebase, and Python scrapers.

## Architecture Overview

This project consists of several key components:

- **Go API Server** (`cmd/api/`) - Main REST API with authentication, rate limiting, and CORS support. This is the main entry point for running the API.
- **Go Scraper Service** (`cmd/scraper/`) - Orchestrates data collection from various sources. This is used to scrape the data from the various sources and store it in different potential locations.
- **Go Uploader Service** (`cmd/scraper/`) - Combines data collected by scraper and uploads the data into Firestore
- **Python Scrapers** (`scripts/`) - Individual scrapers for different data sources
- **Firebase Integration** - Cloud Firestore for data storage and Cloud Storage for file management

### Data Sources

- **Coursebook** - UTD's official course catalog and scheduling system
- **Grade Distributions** - Historical grade data from UTD
- **RMP** - Professor ratings and reviews
- **Integration Service** - Combines data from multiple sources

## Quick Start

### Prerequisites

- **Go** - [Download here](https://golang.org/dl/)
- **Python** - [Download here](https://www.python.org/downloads/)
- **Firebase Project** - Set up a Firebase project with Firestore and Storage. You will need to set up a service account and download the service account key.
- **Chrome/ChromeDriver** - Required for web scraping. This should be handled by the scripts

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/acmutd/acmutd-api.git
   cd acmutd-api
   ```

2. **Run the setup script**

   ```bash
   ./setup.sh
   ```

   This script will:
   - Install Go dependencies
   - Create a Python virtual environment
   - Install all Python requirements
   - Copy the `.env.example` file to `.env` for configuration

3. **Configure environment variables**
   Your `.env` file should look like this:

   ```env
   # Server Configuration
   PORT=8080

   # Firebase Configuration
   FIREBASE_CONFIG=path/to/your/firebase-service-account.json

   # Scraper Configuration
   SCRAPER=coursebook  # Options: coursebook, grades, rmp-profiles, integration
   SAVE_ENVIRONMENT=local  # Options: local, dev, prod

   # Integration Scraper Configuration
   INTEGRATION_SOURCE=local  # Options: local, dev, prod
   INTEGRATION_RESCRAPE=false  # Options: true, false

   # UTD Credentials (for coursebook scraper)
   NETID=your_netid
   PASSWORD=your_password

   # Terms to scrape (comma-separated)
   CLASS_TERMS=24f,25s,25f
   ```

4. **Start the API server**
   The API will be available at `http://localhost:8080`

   ```bash
   go run cmd/api/main.go
   ```

5. **Run the scraper**

   The scraper will run depending on the `SCRAPER` environment variable.
   Depending on the `SAVE_ENVIRONMENT` environment variable, the data will be saved locally or uploaded to Firebase.
   When running the integration scraper, the `INTEGRATION_SOURCE` environment variable determines the data source (local files, dev Firebase, or prod Firebase), and `INTEGRATION_RESCRAPE` determines whether to run scrapers first before processing the data.

   For detailed scraper documentation, see [`SCRAPER.md`](./SCRAPER.md).

   ```bash
   go run cmd/scraper/main.go
   ```

## 📖 API Documentation

Comprehensive API documentation is available in [`/cmd/api/README.md`](./cmd/api/README.md.md).

### Quick Examples

```bash
# Health check (no auth required)
curl http://localhost:8080/health

# Get all courses for Fall 2024 (requires API key)
curl -H "X-API-Key: your-api-key" \
     http://localhost:8080/api/v1/courses/24f

# Search for Computer Science courses
curl -H "X-API-Key: your-api-key" \
     "http://localhost:8080/api/v1/courses/24f?prefix=cs"
```

### Authentication

All API endpoints (except `/health`) require an API key. Admin users can create API keys via the admin endpoints. See the [API Documentation](./API_DOCUMENTATION.md) for details.

## 🛠️ Development

### Project Structure

```bash
acm-api/
├── cmd/                    # Main applications
│   ├── api/               # REST API server
│   ├── scraper/           # Scraper orchestrator
│   └── uploader/          # Data uploader
├── internal/              # Private application code
│   ├── firebase/          # Firebase integration
│   ├── scraper/           # Scraper implementations
│   ├── server/            # HTTP server and middleware
│   ├── types/             # Data models
│   └── uploader/          # Uploader implementation
├── scripts/               # Python scrapers
│   ├── coursebook/        # UTD Coursebook scraper
│   ├── grades/            # Grade distribution processor
│   ├── integration/       # Data integration service
│   ├── professors/        # Professor data aggregator
│   └── rmp-profiles/      # Rate My Professor scraper
└── setup.sh              # Development environment setup
```

### Data Flow

1. **Coursebook Scraper** → Extracts course data from UTD's coursebook system
2. **Grade Processor** → Processes Excel files containing grade distributions
3. **RMP Scraper** → Collects professor ratings from Rate My Professor
4. **Integration Service** → Combines all data sources and uploads to Firebase
5. **Uploader Service** → Collects scraped data and uploads to Firestore
6. **API Server** → Serves integrated data via REST endpoints

## 🔧 Configuration

### Environment Variables

| Variable               | Description                                                       | Required             | Default           |
| ---------------------- | ----------------------------------------------------------------- | -------------------- | ----------------- |
| `PORT`                 | API server port                                                   | No                   | `8080`            |
| `FB_CONFIG`            | Firebase service account JSON filename                            | Yes                  | `acmutd-api.json` |
| `SAVE_ENVIRONMENT`     | Where to save data (local/dev/prod)                               | No                   | `local`           |
| `CLASS_TERMS`          | Comma-separated terms to scrape (e.g., 24f,25s,25f)               | Yes (for scrapers)   | -                 |
| `SCRAPER`              | Which scraper to run (coursebook/grades/rmp-profiles/integration) | Yes (for scraper)    | -                 |
| `NETID`                | UTD NetID for coursebook access                                   | Yes (for coursebook) | -                 |
| `PASSWORD`             | UTD password for coursebook access                                | Yes (for coursebook) | -                 |
| `INTEGRATION_SOURCE`   | Data source for integration scraper (local/dev/prod)              | No                   | `local`           |
| `INTEGRATION_RESCRAPE` | Whether to run scrapers before integration (true/false)           | No                   | `false`           |
| `UPLOADER`             | Which data to upload to Firestore                                 | Yes                  | -                 |

## Deploying to EC2

### 1. Build the Linux Binary (from local Windows)

Cross-compilation is required — running `go build` without these flags produces a Windows executable that cannot run on EC2.

```powershell
# PowerShell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o acmutd-api ./cmd/api/main.go
```

Verify the output is a Linux binary before uploading:
```bash
file ./acmutd-api   # must say: ELF 64-bit LSB executable, x86-64
```

### 2. Build the Frontend

```bash
cd frontend && npm run build && cd ..
```

### 3. Copy Files to EC2

```bash
# Copy the binary
scp -i "path/to/acmapi-key.pem" ./acmutd-api ec2-user@<EC2_IPV4>:/opt/acmutd-api/acmutd-api

# Copy frontend dist — target the parent directory to avoid creating a nested dist/dist/
scp -O -r -i "path/to/acmapi-key.pem" ./frontend/dist ec2-user@<EC2_IPV4>:/opt/acmutd-api/frontend/

# If the service account changed, copy it too
scp -i "path/to/acmapi-key.pem" prod.service_account.json ec2-user@<EC2_IPV4>:/opt/acmutd-api/prod.service_account.json
```

### 4. Restart the Service

```bash
ssh -i "path/to/acmapi-key.pem" ec2-user@<EC2_IPV4>
chmod +x /opt/acmutd-api/acmutd-api
sudo systemctl restart acmutd-api
sudo systemctl status acmutd-api --no-pager
```

### EC2 Environment Configuration

The systemd service loads environment variables from **`/etc/acmutd-api.env`**, not from `/opt/acmutd-api/.env`. The `.env` file in the project directory is for local development only.

> **Important:** `godotenv` does not override environment variables already set by systemd. `/etc/acmutd-api.env` always takes precedence over `/opt/acmutd-api/.env` on EC2.

To view or edit the EC2 environment:
```bash
sudo nano /etc/acmutd-api.env
```

Example `/etc/acmutd-api.env`:
```env
PORT=8080
SAVE_ENVIRONMENT=prod
FB_CONFIG=service_account.json
FIREBASE_WEB_API_KEY=your-web-api-key
FIREBASE_AUTH_DOMAIN=your-project.firebaseapp.com
FIREBASE_PROJECT_ID=your-project-id
```

`SAVE_ENVIRONMENT` controls which Firebase project the server connects to:
- `dev` → uses `dev.service_account.json` → dev Firestore
- `prod` → uses `prod.service_account.json` → prod Firestore

API keys created on a dev server will not work on a prod server (and vice versa) since they are stored in separate Firestore databases.

### Term Format

Terms use a specific format: `{YY}{season}` where:

- `YY` is the 2-digit year (e.g., `24` for 2024)
- `season` is `f` (Fall), `s` (Spring), or `u` (Summer)

Examples: `24f` (Fall 2024), `25s` (Spring 2025), `24u` (Summer 2024)
