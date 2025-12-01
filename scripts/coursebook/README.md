# Coursebook Scraper

This script will scrape coursebook and grab all the course data. You will be asked to log in with your netID and password (this only works per 100 requests, so you may need to refresh the token halfway through scraping).

## Setup

The scraper uses Selenium to automate Chrome. ChromeDriver will be automatically downloaded if not present, but ensure you have Chrome installed and updated.

If desired, go to https://googlechromelabs.github.io/chrome-for-testing/#stable to download the latest version of ChromeDriver to save time downloading it dynamically. Copy the executable to the root folder of this project.

The following environmental variables need to be defined, either passed in the environment or in a `.env` file in the root directory:

```
CLASS_TERMS=[Terms the class are in, comma separated]
NETID=[Your netid]
PASSWORD=[Your password]
```

> For CLASS_TERMS, we need to use the format specified by Coursebook. It should be a 2-digit year number followed by either 'f', 's', or 'u' for "fall", "spring", "summer" (eg. 23f, 24s, 24u, 24f). Note that the terms can listed be in any order.

Then, run the code with:

```bash
python main.py
```

## How It Works

The coursebook scraper operates in several stages to collect comprehensive course data:

### 1. Authentication (`login.py`)
- Uses Selenium WebDriver to automate browser login to UTD Coursebook
- Navigates to the coursebook website and clicks the "protected authentication" link
- Enters your NetID and password credentials
- Waits for successful login and extracts the `PTGSESSID` cookie
- Returns the session token for authenticated API requests
- Automatically refreshes the session if it expires during scraping (after ~100 requests)

### 2. Data Collection (`grab_data.py`)

**Phase A: Filter Discovery**
- Scrapes the coursebook homepage to identify all available filter options:
  - Course prefixes (e.g., `cp_acct`, `cp_cs`, `cp_math`)
  - Schools (e.g., `col_aht`, `col_ecs`, `col_nsm`)
  - Days of the week
  - Course levels (undergraduate, graduate)

**Phase B: Recursive Filter Processing**
- Processes filters in two passes to ensure complete coverage:
  1. **Prefix-based**: Iterates through all course prefixes, then days, then levels
  2. **School-based**: Iterates through all schools, then days, then levels
- For each filter combination, makes a POST request to the coursebook API
- Uses authenticated session cookies to access protected data

**Phase C: Data Extraction**
The scraper handles three different response scenarios:

1. **Multiple Classes (Standard Case)**:
   - Coursebook returns a "report monkey" export ID
   - Makes a follow-up request to `/reportmonkey/cb11-export/{report_id}/json`
   - Extracts comprehensive class data including enrollment, location, instructors, etc.

2. **Single Class (Edge Case)**:
   - No report monkey ID is generated
   - Manually parses the HTML response using BeautifulSoup
   - Extracts limited fields: section, title, days, times, location, instructors

3. **Retry Logic**:
   - If a request fails (network error, expired session), automatically refreshes the session token
   - Retries up to 3 times before giving up

**Phase D: Instructor Netid Extraction**
- Parses HTML to find instructor profile links
- Extracts netIDs from URLs like `http://coursebook.utdallas.edu/search/{netid}`
- Maps instructor names to their netIDs for the `instructor_ids` field

### 3. Deduplication & Output
- Uses `section_address` (e.g., `acct2301.001.24f`) as a unique key
- Automatically deduplicates classes found across multiple filter combinations
- Writes final JSON array to `out/classes_{term}.json`

### Key Technical Details
- **Session Management**: The PTGSESSID cookie has a limited lifetime (~100 requests). The scraper detects failures and automatically re-authenticates.
- **Filter Strategy**: Two-pass approach (prefix + school) ensures all classes are captured, even those that might be missed by a single filter type.
- **Edge Case Handling**: Special logic for single-class results and courses with prefix "utd" (which have hidden course number "STAB").

## Output

The output will be placed in the root of the project, in a file called `classes_[term].json`, ex: `classes_25f.json`.

### Output Format

For most classes, we can get this data:

```json
    {
        "section_address": "lit1301.001.24f",
        "course_prefix": "lit",
        "course_number": "1301",
        "section": "001 ",
        "class_number": "80970",
        "title": "Introduction to Literature ",
        "topic": "",
        "enrolled_status": "Open",
        "enrolled_current": "128",
        "enrolled_max": "130",
        "instructors": "Peter Ingrao",
        "assistants": "",
        "term": "24f",
        "session": "1",
        "days": "Monday, Wednesday",
        "times": "10:00 - 11:15",
        "times_12h": "10:00am - 11:15am",
        "location": "JO_3.516",
        "core_area": "Texas Core Areas 040+090 - Language, Philosophy and Culture + CAO",
        "activity_type": "Lecture",
        "school": "aht",
        "dept": "ahtc",
        "syllabus": "syl149039",
        "textbooks": "9780593450086, 9780804172448, 9780871403315, 9780871403629, 9781538732182 "
    }
```

However, there are TWO edge cases (bruh) that can only get this data:

```json
    {
        "section_address": "lats6300.001.24f",
        "course_prefix": "lats",
        "course_number": "6300",
        "section": "001",
        "title": "Introduction to Latin American Studies  (3 Semester Credit Hours)",
        "term": "24f",
        "instructors": "Humberto Gonzalez Nunez",
        "days": "Tuesday",
        "times_12h": "4:00pm - 6:45pm",
        "location": "JO 3.536"
    }
```
