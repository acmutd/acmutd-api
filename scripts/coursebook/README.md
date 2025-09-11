# Coursebook Scraper

This script will scrape coursebook and grab all the course data. You will be asked to log in with your netID and password (this only works per 100 requests, so you may need to refresh the token halfway through scraping).

## Setup

Go to https://googlechromelabs.github.io/chrome-for-testing/#stable to download the latest version of ChromeDriver. Copy the executable to the root folder of this project. You may also need the latest version of Chrome; make sure your chrome is updated.

The following environmental variables need to be defined, either passed in the environment or in a `.env` file in the root directory:

```
CLASS_TERMS=[Terms the class are in, comma separated]
NETID=[Your netid]
PASSWORD=[Your password]
```

> For CLASS_TERM, we need to use the format specified by Coursebook. It should be a 2-digit year number followed by either 'f' or 's', for "fall" or "spring" (eg. 23f, 24s, 24f, 25s).

Then, run the code with:

```bash
python main.py <semester>
```

## Output

The output will be placed in the root of the project, in a file called `classes.json`.

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
