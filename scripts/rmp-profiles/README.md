# RateMyProfessors (RMP) Scraper

This script scrapes professor data from RateMyProfessors.com for UTD using their internal GraphQL API. It collects comprehensive professor information including ratings, courses taught, student tags, and profile details.

## Setup

Install the required Python dependencies:

```bash
pip install -r requirements.txt
```

The following packages are required:
- `selenium-wire` - for intercepting network requests
- `beautifulsoup4` - for HTML parsing (legacy fallback)
- `aiohttp` - for async requests (legacy fallback)

**Chrome/ChromeDriver**: The scraper uses Selenium to automate Chrome. ChromeDriver will be automatically downloaded if not present, but ensure you have Chrome installed and updated.

## Usage

Run the scraper:

```bash
python main.py
```

The scraper will automatically:
1. Launch a headless Chrome browser
2. Navigate to RateMyProfessors and intercept API requests
3. Extract authentication headers and school ID
4. Query the GraphQL API to retrieve all professor data
5. Save results to `out/rmp_data.json`

## How It Works

### 1. Browser Automation Setup (`setup_driver`)
- Launches Chrome in headless mode using Selenium Wire (to intercept network traffic)
- Configures options to suppress logs and run without a GUI
- Automatically downloads ChromeDriver if not found in the system

### 2. Header and School ID Extraction (`get_headers`)

**Phase A: Page Navigation**
- Navigates to the UTD professor search page on RateMyProfessors
- URL: `https://www.ratemyprofessors.com/search/professors/1273?q=*`
- Closes the cookie consent popup if it appears

**Phase B: Request Interception**
- Clicks the "Show More" pagination button to trigger a GraphQL API request
- Uses Selenium Wire to intercept network traffic
- Waits for and captures the GraphQL request to `ratemyprofessors.com/graphql`

**Phase C: Data Extraction**
- Parses the intercepted request body to extract the `schoolID` (e.g., `U2Nob29sLTEyNzM=`)
- Extracts authentication headers including:
  - `Authorization` token
  - Custom headers like `apollographql-client-name` and `apollographql-client-version`
- These headers are required for subsequent API requests

### 3. GraphQL Query Construction (`build_graphql_query`)
Builds a comprehensive GraphQL query that retrieves:
- Basic info: `id`, `legacyId`, `firstName`, `lastName`
- Ratings: `avgRating`, `numRatings`, `wouldTakeAgainPercent`, `avgDifficulty`
- Academic info: `department`, `school`, `courseCodes`
- Student feedback: `teacherRatingTags` (e.g., "Tough grader", "Gives good feedback")
- Profile metadata: for constructing profile URLs

The query uses pagination to handle large result sets (up to 1000 professors per request).

### 4. API Request Execution (`query_rmp`)

**Phase A: Initial Request**
- Sends the GraphQL query with authentication headers
- Requests up to 1000 professors per page
- Parameters include school ID and pagination cursor

**Phase B: Pagination Loop**
- Checks `pageInfo.hasNextPage` in the response
- If more data exists, updates the `cursor` to the `endCursor` from the previous response
- Continues requesting until all professors are retrieved
- Implements retry logic (up to 3 attempts) for failed requests

**Phase C: Data Accumulation**
- Processes each professor node in the response
- Transforms raw GraphQL data into our standardized format
- Deduplicates professors by normalized name (lowercased, spaces collapsed)

### 5. Data Transformation (`transform_professor_data`)

**Professor Name Normalization**
- Converts names to lowercase
- Collapses multiple spaces into single spaces
- Example: "John  Smith" → "john smith"

**Course Name Normalization**
- Removes spaces, hyphens, and underscores
- Converts to uppercase
- Deduplicates courses
- Example: "CS-1337" → "CS1337"

**Tag Processing**
- Extracts student feedback tags (e.g., "Amazing lectures", "Tough grader")
- Sorts tags by frequency (`tagCount`)
- Includes tag counts for analysis

**Data Structure**
Each professor entry contains:
```json
{
    "department": "Computer Science",
    "url": "https://www.ratemyprofessors.com/professor/123456",
    "quality_rating": 4.5,
    "difficulty_rating": 3.2,
    "would_take_again": 85,
    "original_rmp_format": "John Smith",
    "last_updated": "2025-12-01T10:30:00",
    "ratings_count": 47,
    "courses": ["CS1337", "CS2336", "CS3345"],
    "tags": [
        {"name": "Amazing lectures", "count": 15},
        {"name": "Caring", "count": 12}
    ],
    "rmp_id": "123456"
}
```

### 6. Retry Logic and Error Handling

**Request-Level Retries** (`execute_graphql_request`)
- Retries individual GraphQL requests up to 3 times
- Implements exponential backoff (1s, 2s, 4s delays)
- Handles network errors and invalid responses

**Scraper-Level Retries** (`scrape_rmp_data`)
- If the entire scrape fails, retries the full process up to 3 times
- Re-initializes browser and re-extracts headers on each attempt
- Ensures robustness against transient failures

**Browser Cleanup**
- Always closes the browser in a `finally` block to prevent resource leaks
- Handles cleanup even if exceptions occur

### 7. Output Generation
- Professors are grouped by normalized name as the key
- Each name maps to an array (handles edge case of duplicate names)
- Writes formatted JSON to `out/rmp_data.json` with 4-space indentation
- Logs total professor count and execution time

## Output

The output file `out/rmp_data.json` contains professor data keyed by normalized names:

```json
{
    "john smith": [
        {
            "department": "Computer Science",
            "url": "https://www.ratemyprofessors.com/professor/123456",
            "quality_rating": 4.5,
            "difficulty_rating": 3.2,
            "would_take_again": 85,
            "original_rmp_format": "John Smith",
            "last_updated": "2025-12-01T10:30:00.123456",
            "ratings_count": 47,
            "courses": ["CS1337", "CS2336"],
            "tags": [
                {"name": "Amazing lectures", "count": 15}
            ],
            "rmp_id": "123456"
        }
    ]
}
```

### Key Fields
- **Normalized Name** (dict key): Lowercase name with collapsed spaces for matching
- **original_rmp_format**: Original name formatting from RMP for display
- **quality_rating**: Overall rating (0.0 - 5.0 scale)
- **difficulty_rating**: Course difficulty (0.0 - 5.0 scale)
- **would_take_again**: Percentage of students who would retake (0-100)
- **ratings_count**: Total number of ratings received
- **courses**: Normalized list of course codes taught
- **tags**: Student feedback tags sorted by frequency
- **rmp_id**: RMP legacy ID for constructing profile URLs

## Integration with Firebase

After scraping, the JSON file is uploaded to Firebase Cloud Storage when invoked via the Go driver with `SAVE_ENVIRONMENT=dev` or `SAVE_ENVIRONMENT=prod`. The upload is handled by `internal/scraper/rmp_profiles.go` which:
- Reads `out/rmp_data.json`
- Uploads to `rmp-profiles/rmp_data.json` in Firebase Storage
- Sets the content type to `application/json`
- Cleans up local output files after successful upload

## Technical Notes

### GraphQL API vs Web Scraping
The scraper uses RateMyProfessors' internal GraphQL API instead of traditional web scraping because:
- **Faster**: Single API call retrieves 1000 professors vs pagination through HTML pages
- **More reliable**: Structured JSON responses vs brittle HTML parsing
- **Complete data**: Includes courses, tags, and metadata in one request
- **Easier to maintain**: Less affected by UI changes

A legacy web scraping implementation is preserved in commented code at the bottom of `scraper.py` as a fallback if the API becomes unavailable.

### Rate Limiting and Retry Strategy
- The scraper implements retry logic at both request and scraper levels
- Exponential backoff prevents overwhelming the server
- Browser re-initialization on full scraper retry ensures fresh authentication

### Name Normalization Strategy
- Normalized names are used as keys to enable fuzzy matching with other data sources (coursebook, grades)
- Original formatting is preserved in `original_rmp_format` for display purposes
- Array values handle rare cases where multiple professors share the same normalized name

## Troubleshooting

**"Failed to start the Chrome driver"**
- Ensure Chrome browser is installed and up-to-date
- ChromeDriver should auto-download, but you can manually download from [Chrome for Testing](https://googlechromelabs.github.io/chrome-for-testing/)

**"Could not find GraphQL request"**
- The page structure may have changed
- Check if RateMyProfessors updated their API endpoint
- Increase timeout in `wait_for_graphql_request()` if network is slow

**"Scraped 0 professors"**
- Authentication headers may have expired or changed
- RateMyProfessors may have updated their API authentication
- Check terminal output for error messages from API requests

**Incomplete data (missing courses/tags)**
- GraphQL API occasionally returns partial data
- The scraper includes retry logic to mitigate this
- If persistent, increase `max_retries` in `scrape_rmp_data()`

**Browser doesn't close properly**
- Check for zombie Chrome processes in Task Manager
- The scraper should clean up automatically via `finally` blocks
- Manually kill processes if needed: `Get-Process chrome | Stop-Process`