# ACM API Documentation

## Overview

The ACM API provides access to course and school data for the University of Texas at Dallas. All endpoints return JSON responses and support CORS for cross-origin requests.

**Base URL**: `http://localhost:8080` (or your deployed server URL)

## Authentication

Currently, no authentication is required for any endpoints.

## Response Format

All successful responses follow this general format:

```json
{
  "term": "string",
  "count": number,
  "data": [...]
}
```

Error responses follow this format:

```json
{
  "error": "Error message description"
}
```

## Endpoints

### Health Check

**GET** `/health`

Check if the API is running.

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

### Get All Courses by Term

**GET** `/api/v1/courses/{term}`

Retrieve all courses for a specific term.

**Path Parameters:**

- `term` (required): The academic term (e.g., "2024FALL", "2025SPRING")

**Query Parameters:**

- `prefix` (optional): Filter by course prefix (e.g., "CS", "MATH")
- `number` (optional): Filter by course number (e.g., "1337", "2305")
- `school` (optional): Filter by school (e.g., "ECS", "NSM")

**Response:**

```json
{
  "term": "2024FALL",
  "count": 150,
  "courses": [
    {
      "section_address": "string",
      "course_prefix": "CS",
      "course_number": "1337",
      "section": "001",
      "class_number": "12345",
      "title": "Computer Science I",
      "topic": "",
      "enrolled_status": "Open",
      "enrolled_current": "25",
      "enrolled_max": "30",
      "instructors": "Dr. John Doe",
      "assistants": "",
      "term": "2024FALL",
      "session": "Regular",
      "days": "MW",
      "times": "10:00-11:15",
      "times_12h": "10:00 AM-11:15 AM",
      "location": "ECSS 2.415",
      "core_area": "020",
      "activity_type": "Lecture",
      "school": "ECS",
      "dept": "Computer Science",
      "syllabus": "https://example.com/syllabus",
      "textbooks": "Required textbook information",
      "instructor_ids": "12345"
    }
  ]
}
```

**Examples:**

```bash
# Get all courses for Fall 2024
curl http://localhost:8080/api/v1/courses/2024FALL

# Get all CS courses for Fall 2024
curl "http://localhost:8080/api/v1/courses/2024FALL?prefix=CS"

# Get CS 1337 for Fall 2024
curl "http://localhost:8080/api/v1/courses/2024FALL?prefix=CS&number=1337"

# Get all ECS school courses for Fall 2024
curl "http://localhost:8080/api/v1/courses/2024FALL?school=ECS"
```

### Get Courses by Prefix

**GET** `/api/v1/courses/{term}/prefix/{prefix}`

Retrieve all courses with a specific prefix for a term.

**Path Parameters:**

- `term` (required): The academic term
- `prefix` (required): The course prefix (e.g., "CS", "MATH", "PHYS")

**Response:** Same format as above, but filtered by prefix.

**Example:**

```bash
curl http://localhost:8080/api/v1/courses/2024FALL/prefix/CS
```

### Get Courses by Number

**GET** `/api/v1/courses/{term}/prefix/{prefix}/number/{number}`

Retrieve specific courses by prefix and number for a term.

**Path Parameters:**

- `term` (required): The academic term
- `prefix` (required): The course prefix
- `number` (required): The course number

**Response:** Same format as above, but filtered by prefix and number.

**Example:**

```bash
curl http://localhost:8080/api/v1/courses/2024FALL/prefix/CS/number/1337
```

### Get Courses by School

**GET** `/api/v1/courses/{term}/school/{school}`

Retrieve all courses from a specific school for a term.

**Path Parameters:**

- `term` (required): The academic term
- `school` (required): The school code (e.g., "ECS", "NSM", "JSOM")

**Response:** Same format as above, but filtered by school.

**Example:**

```bash
curl http://localhost:8080/api/v1/courses/2024FALL/school/ECS
```

### Search Courses

**GET** `/api/v1/courses/{term}/search`

Search courses by title, instructor, or other text fields.

**Path Parameters:**

- `term` (required): The academic term

**Query Parameters:**

- `q` (required): Search query string

**Response:** Same format as above, but filtered by search query.

**Example:**

```bash
curl "http://localhost:8080/api/v1/courses/2024FALL/search?q=Computer Science"
```

---

## School Endpoints

### Get Schools by Term

**GET** `/api/v1/schools/{term}`

Retrieve all schools that have courses in a specific term.

**Path Parameters:**

- `term` (required): The academic term

**Response:**

```json
{
  "term": "2024FALL",
  "count": 8,
  "schools": [
    "ECS",
    "NSM",
    "JSOM",
    "AH",
    "BBS",
    "EPPS",
    "IS",
    "ATEC"
  ]
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/schools/2024FALL
```

---

## Course Object Schema

Each course object contains the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `section_address` | string | Unique identifier for the course section |
| `course_prefix` | string | Course prefix (e.g., "CS", "MATH") |
| `course_number` | string | Course number (e.g., "1337", "2305") |
| `section` | string | Section number (e.g., "001", "002") |
| `class_number` | string | Unique class number |
| `title` | string | Course title |
| `topic` | string | Special topic (if applicable) |
| `enrolled_status` | string | Enrollment status ("Open", "Closed", etc.) |
| `enrolled_current` | string | Current enrollment count |
| `enrolled_max` | string | Maximum enrollment capacity |
| `instructors` | string | Instructor names |
| `assistants` | string | Teaching assistant names |
| `term` | string | Academic term |
| `session` | string | Session type ("Regular", "Summer", etc.) |
| `days` | string | Class days ("MW", "TR", "F", etc.) |
| `times` | string | Class times in 24-hour format |
| `times_12h` | string | Class times in 12-hour format |
| `location` | string | Classroom location |
| `core_area` | string | Core curriculum area code |
| `activity_type` | string | Activity type ("Lecture", "Lab", etc.) |
| `school` | string | School code |
| `dept` | string | Department name |
| `syllabus` | string | Syllabus URL |
| `textbooks` | string | Textbook information |
| `instructor_ids` | string | Instructor ID numbers |

---

## Common School Codes

| Code | Full Name |
|------|-----------|
| ECS | Erik Jonsson School of Engineering and Computer Science |
| NSM | School of Natural Sciences and Mathematics |
| JSOM | Naveen Jindal School of Management |
| AH | School of Arts and Humanities |
| BBS | School of Behavioral and Brain Sciences |
| EPPS | School of Economic, Political and Policy Sciences |
| IS | School of Interdisciplinary Studies |
| ATEC | School of Arts, Technology, and Emerging Communication |

---

## Error Codes

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 400 | Bad Request - Missing or invalid parameters |
| 500 | Internal Server Error - Database or server error |

---

## Rate Limiting

Currently, no rate limiting is implemented. Please be respectful of the API and avoid making excessive requests.

---

## CORS Support

The API supports Cross-Origin Resource Sharing (CORS) and allows requests from any origin with the following headers:

- Access-Control-Allow-Origin: *
- Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
- Access-Control-Allow-Headers: Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization

---

## Examples in Different Languages

### JavaScript (Fetch API)

```javascript
// Get all CS courses for Fall 2024
fetch('http://localhost:8080/api/v1/courses/2024FALL?prefix=CS')
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));
```

### Python (requests)

```python
import requests

# Get all CS courses for Fall 2024
response = requests.get('http://localhost:8080/api/v1/courses/2024FALL',
                       params={'prefix': 'CS'})
data = response.json()
print(data)
```

### cURL

```bash
# Get all courses for a term
curl http://localhost:8080/api/v1/courses/2024FALL

# Search for courses
curl "http://localhost:8080/api/v1/courses/2024FALL/search?q=Computer Science"

# Get schools for a term
curl http://localhost:8080/api/v1/schools/2024FALL
```

---

## Versioning

This documentation covers API version 1 (`/api/v1/`). Future versions may introduce breaking changes and will be documented separately.

---

## Support

For issues or questions about the API, please contact the development team or create an issue in the project repository.
