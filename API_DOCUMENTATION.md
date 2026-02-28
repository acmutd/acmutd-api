# ACM API Documentation

## Overview

The ACM API provides access to course and school data for the University of Texas at Dallas. All endpoints return JSON responses and support CORS for cross-origin requests.

**Base URL**: `http://localhost:8080` (or your deployed server URL)

## Authentication

**All API endpoints (except `/health`) require authentication using an API key.**

Include your API key in the request header:

```bash
X-API-Key: your-api-key-here
```

### API Key Management

API keys are managed through admin endpoints and include:

- Rate limiting configuration
- Expiration dates
- Admin privileges
- Usage tracking

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

## Admin Endpoints

### Create API Key

**POST** `/admin/apikeys`

Create a new API key with specified rate limiting and permissions.

**Headers:**

- `X-API-Key`: Admin API key (required)

**Request Body:**

```json
{
  "rate_limit": 100,
  "window_seconds": 60,
  "is_admin": false,
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**Parameters:**

- `rate_limit` (required): Maximum requests allowed per window
- `window_seconds` (required): Time window in seconds for rate limiting
- `is_admin` (optional): Whether the key has admin privileges (default: false)
- `expires_at` (optional): Expiration date in ISO 8601 format

**Response:**

```json
{
  "key": "generated-api-key-string"
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/admin/apikeys \
  -H "X-API-Key: admin-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "rate_limit": 100,
    "window_seconds": 60,
    "is_admin": false,
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

### Get API Key Information

**GET** `/admin/apikeys/{key}`

Retrieve information about a specific API key.

**Headers:**

- `X-API-Key`: Admin API key (required)

**Path Parameters:**

- `key` (required): The API key to retrieve information for

**Response:**

```json
{
  "key": "api-key-string",
  "rate_limit": 100,
  "window_seconds": 60,
  "is_admin": false,
  "created_at": "2024-01-01T00:00:00Z",
  "expires_at": "2024-12-31T23:59:59Z",
  "last_used_at": "2024-01-15T10:30:00Z",
  "usage_count": 150
}
```

**Example:**

```bash
curl http://localhost:8080/admin/apikeys/api-key-here \
  -H "X-API-Key: admin-key-here"
```

---

## Course Endpoints

### Get Courses

**GET** `/api/v1/courses/`

Retrieve courses for a specific term with optional filtering.

**Headers:**

- `X-API-Key`: Your API key (required)

**Query Parameters:**

- `term` (required): The academic term (e.g., "24f", "25s")
- `prefix` (optional): Filter by course prefix (e.g., "cs", "math")
- `number` (optional): Filter by course number (e.g., "1337", "2305")
- `section` (optional): Filter by section (e.g., "001")
- `school` (optional): Filter by school code (e.g., "ecs", "nsm")
- `instructor` (optional): Filter by instructor name (substring match)
- `instructor_id` (optional): Filter by instructor ID (substring match)
- `days` (optional): Filter by days of the week (e.g., "monday", "monday, wednesday")
- `times` (optional): Filter by time in 24h format (e.g., "14:00 - 14:50")
- `times_12h` (optional): Filter by time in 12h format (e.g., "2:00 PM - 2:50 PM")
- `location` (optional): Filter by location (e.g., "SCI_1.210", supports spaces or underscores)
- `q` (optional): Search query for title, topic, or instructor name

**Response:**

```json
{
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
  ],
  "pagination": { ... },
  "query": { ... }
}
```

**Examples:**

```bash
# Get all courses for Fall 2024
curl "http://localhost:8080/api/v1/courses/?term=24f" \
  -H "X-API-Key: your-api-key-here"

# Get all CS courses for Fall 2024
curl "http://localhost:8080/api/v1/courses/?term=24f&prefix=cs" \
  -H "X-API-Key: your-api-key-here"

# Get CS 1337 for Fall 2024
curl "http://localhost:8080/api/v1/courses/?term=24f&prefix=cs&number=1337" \
  -H "X-API-Key: your-api-key-here"

# Search courses by title/instructor
curl "http://localhost:8080/api/v1/courses/?term=24f&q=Computer Science" \
  -H "X-API-Key: your-api-key-here"
```

---

## Term Endpoints

### Get All Terms

**GET** `/api/v1/terms/`

Retrieve all available academic terms in the database.

**Headers:**

- `X-API-Key`: Your API key (required)

**Response:**

```json
{
  "count": 3,
  "terms": ["24f", "25s", "23f"]
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

- `id` (required): The professor's instructor ID

**Response:**

```json
{
  "professor": {
    "instructor_id": "12345",
    "normalized_coursebook_name": "John Doe",
    "original_rmp_format": "Doe, John",
    "department": "Computer Science",
    "url": "https://www.ratemyprofessors.com/...",
    "quality_rating": 4.5,
    "difficulty_rating": 3.2,
    "would_take_again": 85,
    "ratings_count": 120,
    "tags": ["Tough grader", "Gives good feedback"],
    "rmp_id": "67890",
    "overall_grade_rating": 3.8,
    "total_grade_count": 500,
    "course_ratings": {
      "cs1337": 4.2,
      "cs2336": 4.5
    }
  }
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/professors/id/12345 \
  -H "X-API-Key: your-api-key-here"
```

### Get Professors by Name

**GET** `/api/v1/professors/name/{name}`

Search for professors by name (partial match supported).

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `name` (required): The professor's name or partial name to search for

**Response:**

```json
{
  "count": 2,
  "professors": [
    {
      "instructor_id": "12345",
      "normalized_coursebook_name": "John Doe",
      "original_rmp_format": "Doe, John",
      "department": "Computer Science",
      "url": "https://www.ratemyprofessors.com/...",
      "quality_rating": 4.5,
      "difficulty_rating": 3.2,
      "would_take_again": 85,
      "ratings_count": 120,
      "tags": ["Tough grader", "Gives good feedback"],
      "rmp_id": "67890",
      "overall_grade_rating": 3.8,
      "total_grade_count": 500,
      "course_ratings": {
        "cs1337": 4.2,
        "cs2336": 4.5
      }
    },
    {
      "instructor_id": "12346",
      "normalized_coursebook_name": "Jane Doe",
      "original_rmp_format": "Doe, Jane",
      "department": "Mathematics",
      "url": "https://www.ratemyprofessors.com/...",
      "quality_rating": 4.8,
      "difficulty_rating": 2.9,
      "would_take_again": 92,
      "ratings_count": 150,
      "tags": ["Amazing lectures", "Caring"],
      "rmp_id": "67891",
      "overall_grade_rating": 4.0,
      "total_grade_count": 600,
      "course_ratings": {
        "math2413": 4.7,
        "math2414": 4.9
      }
    }
  ]
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/professors/name/john \
  -H "X-API-Key: your-api-key-here"
```

---

## Course Object Schema

Each course object contains the following fields:

| Field              | Type   | Description                                |
| ------------------ | ------ | ------------------------------------------ |
| `section_address`  | string | Unique identifier for the course section   |
| `course_prefix`    | string | Course prefix (e.g., "CS", "MATH")         |
| `course_number`    | string | Course number (e.g., "1337", "2305")       |
| `section`          | string | Section number (e.g., "001", "002")        |
| `class_number`     | string | Unique class number                        |
| `title`            | string | Course title                               |
| `topic`            | string | Special topic (if applicable)              |
| `enrolled_status`  | string | Enrollment status ("Open", "Closed", etc.) |
| `enrolled_current` | string | Current enrollment count                   |
| `enrolled_max`     | string | Maximum enrollment capacity                |
| `instructors`      | string | Instructor names                           |
| `assistants`       | string | Teaching assistant names                   |
| `term`             | string | Academic term                              |
| `session`          | string | Session type ("Regular", "Summer", etc.)   |
| `days`             | string | Class days ("MW", "TR", "F", etc.)         |
| `times`            | string | Class times in 24-hour format              |
| `times_12h`        | string | Class times in 12-hour format              |
| `location`         | string | Classroom location                         |
| `core_area`        | string | Core curriculum area code                  |
| `activity_type`    | string | Activity type ("Lecture", "Lab", etc.)     |
| `school`           | string | School code                                |
| `dept`             | string | Department name                            |
| `syllabus`         | string | Syllabus URL                               |
| `textbooks`        | string | Textbook information                       |
| `instructor_ids`   | string | Instructor ID numbers                      |

---

## Professor Object Schema

Each professor object contains the following fields:

| Field                        | Type               | Description                                         |
| ---------------------------- | ------------------ | --------------------------------------------------- |
| `instructor_id`              | string             | Unique instructor identifier                        |
| `normalized_coursebook_name` | string             | Professor's name in standard format                 |
| `original_rmp_format`        | string             | Professor's name as it appears on RateMyProfessors  |
| `department`                 | string             | Department affiliation                              |
| `url`                        | string             | RateMyProfessors profile URL                        |
| `quality_rating`             | float64            | Overall quality rating (0-5 scale)                  |
| `difficulty_rating`          | float64            | Difficulty rating (0-5 scale)                       |
| `would_take_again`           | int                | Percentage of students who would take again (0-100) |
| `ratings_count`              | int                | Total number of ratings on RateMyProfessors         |
| `tags`                       | []string           | Common tags/descriptors from student reviews        |
| `rmp_id`                     | string             | RateMyProfessors unique identifier                  |
| `overall_grade_rating`       | float64            | Average grade given (GPA scale)                     |
| `total_grade_count`          | int                | Total number of grades recorded                     |
| `course_ratings`             | map[string]float64 | Per-course ratings (course code → rating)           |

---

## Error Codes

| Status Code | Description                                      |
| ----------- | ------------------------------------------------ |
| 200         | Success                                          |
| 400         | Bad Request - Missing or invalid parameters      |
| 401         | Unauthorized - Missing or invalid API key        |
| 403         | Forbidden - Admin access required                |
| 429         | Too Many Requests - Rate limit exceeded          |
| 500         | Internal Server Error - Database or server error |

---

## Rate Limiting

The API implements rate limiting based on your API key configuration:

- Each API key has a configurable rate limit and time window
- Rate limits are enforced per API key
- Rate limit information is included in your API key configuration

---

## CORS Support

The API supports Cross-Origin Resource Sharing (CORS) and allows requests from any origin with the following headers:

- Access-Control-Allow-Origin: \*
- Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
- Access-Control-Allow-Headers: Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key

---

## Examples in Different Languages

### JavaScript (Fetch API)

```javascript
// Get all CS courses for Fall 2024
fetch("http://localhost:8080/api/v1/courses/?term=24f&prefix=cs", {
  headers: {
    "X-API-Key": "your-api-key-here",
  },
})
  .then((response) => response.json())
  .then((data) => console.log(data))
  .catch((error) => console.error("Error:", error));

// Get professor by ID
fetch("http://localhost:8080/api/v1/professors/id/12345", {
  headers: {
    "X-API-Key": "your-api-key-here",
  },
})
  .then((response) => response.json())
  .then((data) => console.log(data))
  .catch((error) => console.error("Error:", error));

// Get professors by name
fetch("http://localhost:8080/api/v1/professors/name/john", {
  headers: {
    "X-API-Key": "your-api-key-here",
  },
})
  .then((response) => response.json())
  .then((data) => console.log(data))
  .catch((error) => console.error("Error:", error));
```

### Python (requests)

```python
import requests

# Get all CS courses for Fall 2024
headers = {'X-API-Key': 'your-api-key-here'}
response = requests.get('http://localhost:8080/api/v1/courses/?term=24f',
                       params={'prefix': 'cs'},
                       headers=headers)
data = response.json()
print(data)

# Get professor by ID
response = requests.get('http://localhost:8080/api/v1/professors/id/12345',
                       headers=headers)
professor = response.json()
print(professor)

# Get professors by name
response = requests.get('http://localhost:8080/api/v1/professors/name/john',
                       headers=headers)
professors = response.json()
print(professors)
```

### cURL

```bash
# Get all available terms
curl http://localhost:8080/api/v1/terms/ \
  -H "X-API-Key: your-api-key-here"

# Get all courses for a term
curl http://localhost:8080/api/v1/courses/?term=24f \
  -H "X-API-Key: your-api-key-here"

# Search for courses
curl "http://localhost:8080/api/v1/courses/?term=24f&q=Computer Science" \
  -H "X-API-Key: your-api-key-here"

# Get professor by ID
curl http://localhost:8080/api/v1/professors/id/12345 \
  -H "X-API-Key: your-api-key-here"

# Get professors by name
curl http://localhost:8080/api/v1/professors/name/john \
  -H "X-API-Key: your-api-key-here"

# Create a new API key (admin only)
curl -X POST http://localhost:8080/admin/apikeys \
  -H "X-API-Key: admin-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "rate_limit": 100,
    "window_seconds": 60,
    "is_admin": false,
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

---

## Versioning

This documentation covers API version 1 (`/api/v1/`). Future versions may introduce breaking changes and will be documented separately.

---

## Support

For issues or questions about the API, please contact the development team or create an issue in the project repository.
