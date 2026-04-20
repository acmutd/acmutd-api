# ACM API Documentation

## Overview

The ACM API provides access to course and school data for the University of Texas at Dallas. All endpoints return JSON responses.

**Base URL**: `http://localhost:8080` (or your deployed server URL)

## Authentication

**All API endpoints (except `/health`) require authentication using an API key.**

Include your API key in the request header:

```bash
X-API-Key: your-api-key-here
```

Admin endpoints additionally require a key with `is_admin: true`.

## Response Format

Error responses follow this format:

```json
{
  "error": "Error message description"
}
```

## Pagination

Paginated endpoints accept `limit` (default `100`, max `100`) and `page` (default `1`) query parameters.

Every paginated response includes a `pagination` object:

```json
{
  "page": 1,
  "limit": 100,
  "has_next": true,
  "next_page": 2
}
```

When `has_next` is `false`, `total` is returned instead of `next_page`:

```json
{
  "page": 1,
  "limit": 100,
  "has_next": false,
  "total": 42
}
```

## Endpoints

### Health Check

**GET** `/health`

Check if the API is running. This endpoint does not require authentication.

**Response:**

```json
{
  "status": "healthy",
  "message": "ACM API is running"
}
```

**Example:**

```bash
curl http://localhost:8080/health
```

---

## Course Endpoints

### Get General Courses

**GET** `/api/v1/courses/general`

Returns course catalog entries with title, description, and grade distributions aggregated across all terms.

**Headers:**

- `X-API-Key`: Your API key (required)

**Query Parameters:**

| Parameter | Type   | Required | Description                                                                |
|-----------|--------|----------|----------------------------------------------------------------------------|
| `prefix`  | string | yes      | Course prefix, e.g. `cs`                                                   |
| `number`  | string | no       | Course number; if provided with `prefix`, returns a single course directly |
| `q`       | string | no       | Case-insensitive substring search on title or description                  |
| `limit`   | int    | no       | Page size (default `100`, max `100`)                                       |
| `page`    | int    | no       | Page number (default `1`)                                                  |

**Response:**

```json
{
  "count": 1,
  "courses": [
    {
      "Prefix": "cs",
      "Number": "1337",
      "Title": "Computer Science I",
      "Description": "Introduces a methodology for problem solving using the computer...",
      "CreditHours": 3,
      "School": "Erik Jonsson School of Engineering and Computer Science",
      "SchoolID": "ECS",
      "Core": "",
      "ScheduleFrequency": "Every semester",
      "SectionTypes": ["Lecture", "Laboratory"],
      "Grades": {
        "25s": {
          "A+": 12, "A": 48, "A-": 10,
          "B+": 8,  "B": 22, "B-": 5,
          "C+": 3,  "C": 10, "C-": 2,
          "D+": 1,  "D": 3,  "D-": 0,
          "F": 2, "W": 5, "I": 0, "CR": 0, "NC": 0, "NF": 0, "P": 0
        }
      },
      "LastUpdatedTerm": "25s"
    }
  ],
  "pagination": { "page": 1, "limit": 100, "has_next": false, "total": 1 }
}
```

**Examples:**

```bash
# Get all CS courses
curl "http://localhost:8080/api/v1/courses/general?prefix=cs" \
  -H "X-API-Key: your-api-key-here"

# Search CS courses by title or description
curl "http://localhost:8080/api/v1/courses/general?prefix=cs&q=data+structures" \
  -H "X-API-Key: your-api-key-here"
```

### Get General Course by Prefix and Number

**GET** `/api/v1/courses/general/{prefix}/{number}`

Returns a single course by prefix and number. Response is the course object above (not wrapped in a list).

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `prefix` (required): Course prefix, e.g. `cs`
- `number` (required): Course number, e.g. `1337`

**Example:**

```bash
curl http://localhost:8080/api/v1/courses/general/cs/1337 \
  -H "X-API-Key: your-api-key-here"
```

---

### Get Sections

**GET** `/api/v1/courses/sections`

Returns individual course sections for a given term.

**Headers:**

- `X-API-Key`: Your API key (required)

**Query Parameters:**

| Parameter    | Type   | Required | Description                                                                     |
|--------------|--------|----------|---------------------------------------------------------------------------------|
| `term`       | string | yes      | Term code, e.g. `25s` (Spring 2025), `24f` (Fall 2024), `24u` (Summer 2024)   |
| `prefix`     | string | no       | Course prefix, e.g. `cs`                                                        |
| `number`     | string | no       | Course number, e.g. `1337`                                                      |
| `instructor` | string | no       | Exact match on normalized instructor name, e.g. `john doe`                     |
| `days`       | string | no       | Exact match on a meeting day, e.g. `monday`                                    |
| `building`   | string | no       | Exact match on building code, e.g. `ecss` — **requires `prefix`**              |
| `title`      | string | no       | Case-insensitive substring match on section title — **requires `prefix`**      |
| `limit`      | int    | no       | Page size (default `100`, max `100`)                                            |
| `page`       | int    | no       | Page number (default `1`)                                                       |

**Response:**

```json
{
  "count": 1,
  "sections": [
    {
      "section_address": "cs1337.001.25s",
      "prefix": "cs",
      "number": "1337",
      "section": "001",
      "term": "25s",
      "title": "Computer Science I",
      "description": "Introduces a methodology for problem solving...",
      "credit_hours": 3,
      "activity_type": "Lecture",
      "instruction_mode": "Face-to-Face",
      "grading": "Standard Letter",
      "class_level": "Undergraduate",
      "enrolled_status": "Open",
      "enrolled_current": 34,
      "enrolled_max": 45,
      "waitlist": 0,
      "days": ["Tuesday", "Thursday"],
      "time": "10:00am - 11:15am",
      "building": "ecss",
      "room": "2.410",
      "location": "ECSS 2.410",
      "start_date": "2025-01-13",
      "end_date": "2025-05-02",
      "instructors": ["John Doe"],
      "instructor_ids": ["jxd123456"],
      "instructor_name_normalized": "john doe",
      "tas": [],
      "ta_ids": [],
      "cross_listed": [],
      "grades": {
        "A+": 5, "A": 20, "B": 12, "W": 2,
        "F": 1, "I": 0, "CR": 0, "NC": 0, "NF": 0, "P": 0
      }
    }
  ],
  "pagination": { "page": 1, "limit": 100, "has_next": false, "total": 1 }
}
```

**Examples:**

```bash
# Get all CS sections for Spring 2025
curl "http://localhost:8080/api/v1/courses/sections?term=25s&prefix=cs" \
  -H "X-API-Key: your-api-key-here"

# Get CS 1337 sections taught on Tuesdays
curl "http://localhost:8080/api/v1/courses/sections?term=25s&prefix=cs&number=1337&days=tuesday" \
  -H "X-API-Key: your-api-key-here"

# Get CS sections in ECSS building
curl "http://localhost:8080/api/v1/courses/sections?term=25s&prefix=cs&building=ecss" \
  -H "X-API-Key: your-api-key-here"
```

### Get Section by Address

**GET** `/api/v1/courses/sections/{prefix}/{number}/{section}/{term}`

Returns a single section by its full address. Response is the section object above (not wrapped in a list).

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `prefix` (required): Course prefix, e.g. `cs`
- `number` (required): Course number, e.g. `1337`
- `section` (required): Section number, e.g. `001`
- `term` (required): Term code, e.g. `25s`

**Example:**

```bash
curl http://localhost:8080/api/v1/courses/sections/cs/1337/001/25s \
  -H "X-API-Key: your-api-key-here"
```

---

## Term Endpoints

### Get All Terms

**GET** `/api/v1/terms/`

Retrieve all available academic terms in the database.

**Headers:**

- `X-API-Key`: Your API key (required)

**Query Parameters:**

| Parameter | Type | Required | Description                          |
|-----------|------|----------|--------------------------------------|
| `limit`   | int  | no       | Page size (default `100`, max `100`) |
| `page`    | int  | no       | Page number (default `1`)            |

**Response:**

```json
{
  "count": 3,
  "terms": ["25s", "24f", "24u"],
  "pagination": { "page": 1, "limit": 100, "has_next": false, "total": 3 }
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/terms/ \
  -H "X-API-Key: your-api-key-here"
```

---

## Professor Endpoints

### Get Professor by ID

**GET** `/api/v1/professors/id/{id}`

Retrieve a specific professor by their instructor ID.

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `id` (required): The professor's instructor ID, e.g. `jxd123456`

**Response:**

```json
{
  "professor": {
    "instructor_id": "jxd123456",
    "normalized_coursebook_name": "john doe",
    "original_rmp_format": "John Doe",
    "department": "Computer Science",
    "url": "https://www.ratemyprofessors.com/professor/12345",
    "quality_rating": 4.5,
    "difficulty_rating": 3.2,
    "would_take_again": 87,
    "ratings_count": 142,
    "tags": ["Gives good feedback", "Caring"],
    "rmp_id": "12345",
    "overall_grade_rating": 3.8,
    "total_grade_count": 312,
    "course_ratings": {
      "cs1337": 4.1,
      "cs2337": 3.9
    }
  }
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/professors/id/jxd123456 \
  -H "X-API-Key: your-api-key-here"
```

### Get Professors by Name

**GET** `/api/v1/professors/name/{name}`

Search for professors whose normalized name starts with the given value. Useful for autocomplete.

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `name` (required): Name prefix to search, e.g. `john doe`

**Query Parameters:**

| Parameter | Type | Required | Description                          |
|-----------|------|----------|--------------------------------------|
| `limit`   | int  | no       | Page size (default `100`, max `100`) |
| `page`    | int  | no       | Page number (default `1`)            |

**Response:**

```json
{
  "count": 1,
  "professors": [
    {
      "instructor_id": "jxd123456",
      "normalized_coursebook_name": "john doe",
      "original_rmp_format": "John Doe",
      "department": "Computer Science",
      "url": "https://www.ratemyprofessors.com/professor/12345",
      "quality_rating": 4.5,
      "difficulty_rating": 3.2,
      "would_take_again": 87,
      "ratings_count": 142,
      "tags": ["Gives good feedback", "Caring"],
      "rmp_id": "12345",
      "overall_grade_rating": 3.8,
      "total_grade_count": 312,
      "course_ratings": {
        "cs1337": 4.1,
        "cs2337": 3.9
      }
    }
  ],
  "pagination": { "page": 1, "limit": 100, "has_next": false, "total": 1 }
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/professors/name/john \
  -H "X-API-Key: your-api-key-here"
```

---

## Object Schemas

### Course (General)

| Field               | Type                        | Description                                              |
|---------------------|-----------------------------|----------------------------------------------------------|
| `Prefix`            | string                      | Course prefix, e.g. `cs`                                 |
| `Number`            | string                      | Course number, e.g. `1337`                               |
| `Title`             | string                      | Course title                                             |
| `Description`       | string                      | Course description                                       |
| `CreditHours`       | int                         | Credit hours                                             |
| `School`            | string                      | Full school name                                         |
| `SchoolID`          | string                      | School code, e.g. `ECS`                                  |
| `Core`              | string                      | Core curriculum area code, if applicable                 |
| `ScheduleFrequency` | string                      | How often the course is offered                          |
| `SectionTypes`      | []string                    | Types of sections offered, e.g. `["Lecture", "Lab"]`    |
| `Grades`            | map[string]GradeDistribution| Grade distributions keyed by term, e.g. `"25s"`         |
| `LastUpdatedTerm`   | string                      | Most recent term this course was seen                    |

### Section

| Field                        | Type              | Description                                           |
|------------------------------|-------------------|-------------------------------------------------------|
| `section_address`            | string            | Unique identifier, e.g. `cs1337.001.25s`              |
| `prefix`                     | string            | Course prefix                                         |
| `number`                     | string            | Course number                                         |
| `section`                    | string            | Section number                                        |
| `term`                       | string            | Term code                                             |
| `title`                      | string            | Section title                                         |
| `description`                | string            | Course description                                    |
| `credit_hours`               | int               | Credit hours                                          |
| `activity_type`              | string            | `Lecture`, `Laboratory`, etc.                         |
| `instruction_mode`           | string            | `Face-to-Face`, `Online`, etc.                        |
| `grading`                    | string            | Grading basis                                         |
| `class_level`                | string            | `Undergraduate`, `Graduate`                           |
| `enrolled_status`            | string            | `Open`, `Closed`, `Waitlist`                          |
| `enrolled_current`           | int               | Current enrollment count                              |
| `enrolled_max`               | int               | Maximum enrollment capacity                           |
| `waitlist`                   | int               | Current waitlist count                                |
| `days`                       | []string          | Meeting days, e.g. `["Tuesday", "Thursday"]`          |
| `time`                       | string            | Meeting time, e.g. `10:00am - 11:15am`               |
| `building`                   | string            | Building code, e.g. `ecss`                            |
| `room`                       | string            | Room number                                           |
| `location`                   | string            | Full location string                                  |
| `start_date`                 | string            | Section start date                                    |
| `end_date`                   | string            | Section end date                                      |
| `instructors`                | []string          | Instructor display names                              |
| `instructor_ids`             | []string          | Instructor IDs                                        |
| `instructor_name_normalized` | string            | Normalized instructor name used for filtering         |
| `tas`                        | []string          | Teaching assistant names                              |
| `ta_ids`                     | []string          | Teaching assistant IDs                                |
| `cross_listed`               | []string          | Cross-listed section addresses                        |
| `grades`                     | GradeDistribution | Grade distribution for this section (if available)    |

### Professor

| Field                        | Type               | Description                                          |
|------------------------------|--------------------|------------------------------------------------------|
| `instructor_id`              | string             | Unique instructor identifier                         |
| `normalized_coursebook_name` | string             | Lowercase name used for search and matching          |
| `original_rmp_format`        | string             | Name as it appears on RateMyProfessors               |
| `department`                 | string             | Department affiliation                               |
| `url`                        | string             | RateMyProfessors profile URL                         |
| `quality_rating`             | float64            | Overall quality rating (0–5)                         |
| `difficulty_rating`          | float64            | Difficulty rating (0–5)                              |
| `would_take_again`           | int                | Percentage of students who would take again (0–100)  |
| `ratings_count`              | int                | Total number of RMP ratings                          |
| `tags`                       | []string           | Common descriptors from student reviews              |
| `rmp_id`                     | string             | RateMyProfessors unique identifier                   |
| `overall_grade_rating`       | float64            | Average grade given (GPA scale)                      |
| `total_grade_count`          | int                | Total number of grades recorded                      |
| `course_ratings`             | map[string]float64 | Per-course average ratings, keyed by course code     |

---

## Error Codes

| Status Code | Description                                      |
|-------------|--------------------------------------------------|
| 200         | Success                                          |
| 400         | Bad Request — missing or invalid parameters      |
| 401         | Unauthorized — missing or invalid API key        |
| 403         | Forbidden — admin access required                |
| 429         | Too Many Requests — rate limit exceeded          |
| 500         | Internal Server Error — database or server error |

---

## Rate Limiting

Rate limits are configured per API key. Each key has a `rate_limit` (max requests) and `window_seconds` (rolling window). Exceeding the limit returns a `429` response.
