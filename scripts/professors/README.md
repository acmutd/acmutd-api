# Professor Ratings Data Aggregator

This project aggregates professor ratings data from RateMyProfessors (RMP) and grade distribution data from UTD Grades, providing a more comprehensive view of professors' performance. It aims to improve upon existing solutions by handling name inconsistencies and duplicate professor entries.

## Key Features

- **Enhanced Professor Matching:**
  - Handles name variations (e.g., middle initials, misspellings, hyphenated names) through robust name normalization.
  - Resolves duplicate professor entries by matching based on courses taught.
  - Professors are validated uniquely by further matching data with Coursebook to ensure that duplicates can be properly identified
  - Implements direct matching for exact name matches.
  - Fuzzy matching for close but not exact matches.
- **Manual Mappings:** Allows for manual creation of mappings between professor names (e.g., nicknames, alternate names) to ensure accurate matches.
- **Scalable Data Structure:** Stores data in a dictionary where keys are normalized professor names and values are lists of professor entries, accommodating duplicate names.
- **Comprehensive Data Output:**
  - Combines RMP ratings (quality, difficulty, would-take-again, tags) with grade distribution data (overall grade rating, course-specific ratings).
  - Includes unmatched grade distribution data, enabling display of grade averages even for professors without RMP profiles.

## Data Structure

The data is stored in a JSON-like structure where keys are normalized professor names and values are lists of professor entries for professors with that name. Each professor entry contains:

```json
{
    "department": "...",
    "url": "...",
    "quality_rating": ...,
    "difficulty_rating": ...,
    "would_take_again": ...,
    "original_rmp_format": "...",
    "last_updated": "...",
    "ratings_count": ...,
    "tags": ["...", "..."],
    "rmp_id": "...",
    "instructor_id": "...",
    "overall_grade_rating": ...,
    "total_grade_count": ...,
    "course_ratings": {
        "COURSE_CODE": ...,
        "...": ...
    }
}
```

`department`: Professor's listed department from RMP.

`url`: RateMyProfessors profile URL.

`quality_rating`: Overall quality rating from RMP.

`difficulty_rating`: Difficulty rating from RMP.

`would_take_again`: Percentage of students who would take the professor again. (-1 if N/A)

`original_rmp_format`: The name that was originally stored on ratemyprofessor.

`last_updated`: Timestamp of the last data update.

`ratings_count`: Number of RMP ratings.

`tags`: RMP tags (e.g., "Amazing lectures").

`rmp_id`: RateMyProfessors ID.

`instructor_id`: Instructor ID listed on Coursebook.

`overall_grade_rating`: Overall grade rating from grade data (average from 4.0 scaled up to 5.0).

`total_grade_count`: Total number of grades in grade data.

`course_ratings`: Course-specific grade ratings (average from 4.0 scaled up to 5.0).

### Example Professor Entry

There are two professors with the name "Jason Bennett" at UTD, each is stored together in the list for the normalized key 'jason bennett', where one was matched with the singular RMP profile for a "Jason Bennett"

```
"jason bennett": [
        {
            "department": "Rhetoric",
            "url": "https://www.ratemyprofessors.com/professor/3044802",
            "quality_rating": 5,
            "difficulty_rating": 1,
            "would_take_again": 100,
            "original_rmp_format": "Jason Bennett",
            "last_updated": "2025-03-08T23:07:48.353461",
            "ratings_count": 1,
            "tags": [
                "Participation matters",
                "Amazing lectures ",
                "Gives good feedback"
            ],
            "rmp_id": "3044802",
            "instructor_id": "jhb042000",
            "overall_grade_rating": 4.13,
            "total_grade_count": 34,
            "course_ratings": {
                "RHET1302": 4.13
            }
        },
        {
            "instructor_id": "jxb230049",
            "overall_grade_rating": 4.5,
            "total_grade_count": 309,
            "course_ratings": {
                "MUSI1306": 4.5
            }
        }
    ],
```

## Matching Logic

1. **Manual Matches**: Applies predefined mappings from manual_matches.json for known name variations (Yu Chung Ng --> Vincent Ng).
2. **Direct Matches**: Matches professors with identical normalized names (John Cole --> John Cole).
3. **Duplicate Handling**: For duplicate professor names, matches based on course overlap (2x Jason Bennetts with grade distributions, 3x Hien Nguyen RMP profiles).
4. **Fuzzy Matches**: Applies fuzzy matching for professors with similar names, confirmed by course overlap (Joseph Nedbal --> Joe Nedbal, Andres Ricardo Sanchez De La Rosa --> Andres Sanchez).
5. **Unmatched Data**: Appends remaining unmatched grade distribution data to the corresponding professor entry.

## Name Normalization

The following regular expressions are used for name normalization (Python):

```
def normalize_name(name):
"""Normalizes names, removes periods, handles middle names, replaces hyphens, and potential swaps."""
name = name.strip()
name = re.sub(r"\s*,\s*", ", ", name) # standardize comma spacing
name = re.sub(r"\s+[A-Z](.[A-Z])*\s\_$", "", name) # remove middle initials
name = re.sub(r"([A-Z])\.([A-Z])", r"\1 \2", name) # add space between initials
name = re.sub(r"[.\s]+", " ", name) # removes periods and extra spaces
name = re.sub(r"[’'ʻ`]", "", name) # remove apostrophes
name = name.replace('-', ' ') # replace hyphens with spaces

    if ", " in name: # handle the Last, First formats by splitting up and swapping
        last, first = name.split(", ", 1)
        return f"{first.strip().lower()} {last.strip().lower()}"
    else:
        return name.strip().lower()
```

## Usage (UTD Grades)

1. Normalize the professor's name from UTD Grades.
2. Check if the normalized name exists as a key in the data structure.
3. If the key exists:
   - If there's only one entry, use it.
   - If there are multiple entries, match based on the course being taught.
4. Display the combined RMP and grade distribution data.

## Code

All original code and commit history is available at: https://github.com/emw8105/professor-ratings-script

* **`aggregator.py`:** This file aggregates professor grade data from the grade distributions sourced from UTD Grades. It includes functions for name normalization, extracting data from the given CSV grades data, matching the data with the Coursebook classes data, processing the data, and calculating both per course and overall grade ratings for each professor
* **`scraper.py`:** This file is responsible for scraping professor data from RateMyProfessors. It utilizes selenium to obtain header information on the RMP site to access the RMP internal GraphQL API, which it then sends requests to extract relevant information such as quality ratings, difficulty ratings, tags, and ratings counts.
* **`main.py`:** This file serves as the entry point for the program and contains the core logic for matching professor data from RateMyProfessors (RMP) and UTD Grades. It includes functionionality for direct matching, fuzzy matching, and handling duplicate professor entries. It also handles the creation of the final JSON output.

### Data Sources

The project utilizes the following data sources:

* **`data/classes`:** This file contains a list of classes and their respective professors from Coursebook. It is used to identify course overlap during the matching process. Data is sourced from [https://github.com/MichaelZhao21/scrape-coursebook](https://github.com/MichaelZhao21/scrape-coursebook)
* **`data/grades`:** This file contains grade distribution data from UTD Grades, including overall grade ratings and course-specific grade ratings. Data is sourced from [https://github.com/acmutd/utd-grades/tree/master/raw_data](https://github.com/acmutd/utd-grades/tree/master/raw_data)
* **RateMyProfessors:** The program queries the GraphQL API on the RateMyProfessors website to retrieve RMP information for UTD Professors.

The `classes` and `grades` files are assumed to be pre-existing and properly formatted. The program focuses on processing and merging this data with the scraped RMP data.

The python code can be found on GitHub: [https://github.com/emw8105/professor-ratings-script/tree/main](https://github.com/emw8105/professor-ratings-script/tree/main)

## Notes

- A "would_take_again" value of -1 indicates N/A.
- Due to variations in the GraphQL API, slight differences in results have been seen between executions
  - Ocassionally certain professors may have a -1 would_take_again value, no tags, or no courses, some manual confirmation of the data validity is necessary until fixed
