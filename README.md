# ACM UTD API

A comprehensive REST API for accessing University of Texas at Dallas course data, including course information, grade distributions, and professor ratings. Built with Go, Firebase, and Python scrapers.

## Architecture Overview

This project consists of several key components:

- **Go API Server** (`cmd/api/`) - Main REST API with authentication, rate limiting, and CORS support. This is the main entry point for running the API.
- **Go Scraper Service** (`cmd/scraper/`) - Orchestrates data collection from various sources. This is used to scrape the data from the various sources and store it in different potential locations.
- **Python Scrapers** (`scripts/`) - Individual scrapers for different data sources
- **Firebase Integration** - Cloud Firestore for data storage and Cloud Storage for file management

### Data Sources

- **Coursebook** - UTD's official course catalog and scheduling system
- **Grade Distributions** - Historical grade data from UTD
- **Rate My Professor** - Professor ratings and reviews
- **Integration Service** - Combines data from multiple sources

## Quick Start

### Prerequisites

- **Go+** - [Download here](https://golang.org/dl/)
- **Python+** - [Download here](https://www.python.org/downloads/)
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
   SAVE_ENVIRONMENT=development

   INTEGRATION_MODE=local # local, dev, prod, rescrape

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
    When running the integration scraper, the `INTEGRATION_MODE` environment variable will determine whether to pull the data from Firebase, grab the data from local files, or rerun all the scrapers.

   ```bash
   go run cmd/scraper/main.go
   ```

## 📖 API Documentation

Comprehensive API documentation is available in [`API_DOCUMENTATION.md`](./API_DOCUMENTATION.md).

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
│   └── scraper/           # Scraper orchestrator
├── internal/              # Private application code
│   ├── firebase/          # Firebase integration
│   ├── scraper/           # Scraper implementations
│   ├── server/            # HTTP server and middleware
│   └── types/             # Data models
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
5. **API Server** → Serves integrated data via REST endpoints

## 🔧 Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `PORT` | API server port | No | `8080` |
| `FIREBASE_CONFIG` | Path to Firebase service account JSON | Yes | - |
| `SCRAPER` | Which scraper to run | Yes (for scraper) | - |
| `SAVE_ENVIRONMENT` | Environment for data saving | No | `development` |
| `NETID` | UTD NetID for coursebook access | Yes (for coursebook) | - |
| `PASSWORD` | UTD password for coursebook access | Yes (for coursebook) | - |
| `CLASS_TERMS` | Comma-separated terms to scrape | Yes (for scrapers) | - |
| `INTEGRATION_MODE` | Mode for integration scraper | Yes (for integration) | - |

### Term Format

Terms use a specific format: `{YY}{season}` where:

- `YY` is the 2-digit year (e.g., `24` for 2024)
- `season` is `f` (Fall), `s` (Spring), or `u` (Summer)

Examples: `24f` (Fall 2024), `25s` (Spring 2025), `24u` (Summer 2024)
