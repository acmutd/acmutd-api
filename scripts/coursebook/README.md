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

1. **Scrape Each Class (Standard Case)**:
   - Coursebook request gives a list of classes
   - Get class overview html for each class `https://coursebook.utdallas.edu/clips/clip-cb11-hat.zog`
   - Parse the html for all class data.

2. **Retry Logic**:
   - If a request fails (network error, expired session), automatically refreshes the session token
   - Retries up to 3 times before giving up

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
        "section_address": "acct2302.010.26s",
        "course_prefix": "acct",
        "course_number": "2302",
        "section": "010",
        "class_number": "28741",
        "class_level": "Undergraduate",
        "instruction_mode": "Face-to-Face",
        "title": "Introductory Management Accounting",
        "description": "ACCT 2302- Introductory Management Accounting(3 semester credit hours) This course helps students to build the necessary skills in the managerial use of accounting information for planning, decision making, performance evaluation, and controlling operations. The course uses a general framework for product costing systems, budgeting and variance analysis in order to benefit all students with a wide variety of career paths. A minimum grade of C is required to take upper-division ACCT courses. Prerequisite:ACCT 2301. (3-0) S",
        "enrolled_status": "OPEN",
        "enrolled_current": 62,
        "enrolled_max": 63,
        "waitlist": 0,
        "term": "26s",
        "days": "Tuesday, Thursday",
        "times_12h": "1:00pm-2:15pm",
        "location": "JSOM 2.106",
        "activity_type": "Lecture",
        "instructors": [
            "Christopher Hes"
        ],
        "instructor_netids": [
            "cah041000"
        ],
        "tas": [
            "Meghana Sri Vedala"
        ],
        "ta_netids": [
            "mxv200018"
        ],
        "school": "Naveen Jindal School of Management"
    },
```

**Possible future data:**
- syllabus
- textbooks
- core_area
- topic