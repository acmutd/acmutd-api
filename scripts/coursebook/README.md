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

> Optional: RESUME, resume from a specific filter or set to "combine" to skip the scraping and combine existing JSON data. Note that if the filter doesn't appear in prefix or school filters it will start from the beginning. (eg. cp_cs, col_mgt, combine)

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
The scraper handles two different response scenarios:

1. **Scrape Each Class**:
   - Coursebook request gives a list of classes
   - Get class overview html for each class in the list `https://coursebook.utdallas.edu/clips/clip-cb11-hat.zog`
   - Parse the html for all class data (`parse.py`).

2. **Retry Logic**:
   - If a request fails (network error, expired session), automatically refreshes the session token
   - Retries up to 3 times before giving up

### 3. Deduplication & Output
1. After each filter is completed, saves output to `{term}/{filter}.json`. If further filtered by day or level, results are saved to `{term}/{filter}/{sub_filter}.json`.

2. After all filters are completed, it combines all JSON in `{term}/`
    - Uses `section_address` (e.g., `acct2301.001.24f`) as a unique key
    - Automatically deduplicates classes found across multiple filter combinations
    - Writes final JSON array to `out/classes_{term}.json`

### Key Technical Details
- **Session Management**: The PTGSESSID cookie has a limited lifetime (~100 requests). The scraper detects failures and automatically re-authenticates.
- **Filter Strategy**: Two-pass approach (prefix + school) ensures all classes are captured, even those that might be missed by a single filter type.

## Output

The output will be placed in `out/classes_[term].json`, ex: `classes_25f.json`.

### Output Format

```json
    {
        "section_address": "acct2301.001.25s",
        "course_prefix": "acct",
        "course_number": "2301",
        "section": "001",
        "class_course_number": "26595 / 000061",
        "class_level": "Undergraduate",
        "instruction_mode": "Face-to-Face",
        "title": "Introductory Financial Accounting",
        "description": "ACCT 2301- Introductory Financial Accounting(3 semester credit hours) An introduction to financial reporting designed to create an awareness of the accounting concepts and principles for preparing the three basic financial statements: the income statement, balance sheet, and statement of cash flows. A minimum grade of C is required to take upper-division ACCT courses. (3-0) S",
        "enrolled_status": "OPEN",
        "enrolled_current": 64,
        "enrolled_max": 67,
        "waitlist": 0,
        "term": "25s",
        "days": [
            "Tuesday",
            "Thursday"
        ],
        "times_12h": "8:30am-9:45am",
        "location": "JSOM 2.717",
        "activity_type": "Lecture",
        "semester_credit_hours": "3",
        "core": null,
        "grading": "Graded - Undergraduate",
        "session_type": "Regular Academic Session",
        "add_consent": "No Consent",
        "enrollment_reqs": [
            "ACCT 2301 Repeat Restriction"
        ],
        "class_attributes": [],
        "class_notes": null,
        "instructors": [
            "Jieying Zhang",
            "Naim Bugra Ozel"
        ],
        "instructor_ids": [
            "jxz146230",
            "nbo150030"
        ],
        "tas": [
            "Galymzhan Tazhibayev",
            "Dipta Banik"
        ],
        "ta_ids": [
            "gxt230023",
            "dxb220047"
        ],
        "school": "Naveen Jindal School of Management",
        "school_id": "jsom",
        "syllabus": "syl152552"
    },
```
