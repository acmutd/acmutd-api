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

### Get All Courses by Term

**GET** `/api/v1/courses/{term}`

Retrieve all courses for a specific term.

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `term` (required): The academic term (e.g., "24f", "25s")

**Query Parameters:**

- `prefix` (optional): Filter by course prefix (e.g., "cs", "math")
- `number` (optional): Filter by course number (e.g., "1337", "2305")
- `school` (optional): Filter by school (e.g., "ecs", "nsm")

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
curl http://localhost:8080/api/v1/courses/24f \
  -H "X-API-Key: your-api-key-here"

# Get all CS courses for Fall 2024
curl "http://localhost:8080/api/v1/courses/24f?prefix=cs" \
  -H "X-API-Key: your-api-key-here"

# Get CS 1337 for Fall 2024
curl "http://localhost:8080/api/v1/courses/24f?prefix=cs&number=1337" \
  -H "X-API-Key: your-api-key-here"

# Get all ECS school courses for Fall 2024
curl "http://localhost:8080/api/v1/courses/24f?school=ecs" \
  -H "X-API-Key: your-api-key-here"
```

### Get Courses by Prefix

**GET** `/api/v1/courses/{term}/prefix/{prefix}`

Retrieve all courses with a specific prefix for a term.

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `term` (required): The academic term
- `prefix` (required): The course prefix (e.g., "CS", "MATH", "PHYS")

**Response:** Same format as above, but filtered by prefix.

**Example:**

```bash
curl http://localhost:8080/api/v1/courses/24f/prefix/cs \
  -H "X-API-Key: your-api-key-here"
```

### Get Courses by Number

**GET** `/api/v1/courses/{term}/prefix/{prefix}/number/{number}`

Retrieve specific courses by prefix and number for a term.

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `term` (required): The academic term
- `prefix` (required): The course prefix
- `number` (required): The course number

**Response:** Same format as above, but filtered by prefix and number.

**Example:**

```bash
curl http://localhost:8080/api/v1/courses/24f/prefix/cs/number/1337 \
  -H "X-API-Key: your-api-key-here"
```

### Get Courses by School

**GET** `/api/v1/courses/{term}/school/{school}`

Retrieve all courses from a specific school for a term.

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `term` (required): The academic term
- `school` (required): The school code (e.g., "ECS", "NSM", "JSOM")

**Response:** Same format as above, but filtered by school.

**Example:**

```bash
curl http://localhost:8080/api/v1/courses/24f/school/ecs \
  -H "X-API-Key: your-api-key-here"
```

### Search Courses

**GET** `/api/v1/courses/{term}/search`

Search courses by title, instructor, or other text fields.

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `term` (required): The academic term

**Query Parameters:**

- `q` (required): Search query string

**Response:** Same format as above, but filtered by search query.

**Example:**

```bash
curl "http://localhost:8080/api/v1/courses/24f/search?q=Computer Science" \
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
  "terms": [
    "24f",
    "25s",
    "23f"
  ]
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/terms/ \
  -H "X-API-Key: your-api-key-here"
```

---

## School Endpoints

### Get Schools by Term

**GET** `/api/v1/schools/{term}`

Retrieve all schools that have courses in a specific term.

**Headers:**

- `X-API-Key`: Your API key (required)

**Path Parameters:**

- `term` (required): The academic term

**Response:**

```json
{
  "term": "24f",
  "count": 8,
  "schools": [
    "ecs",
    "nsm",
    "jsom",
    "ah",
    "bbs",
    "epps",
    "is",
    "atec"
  ]
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/schools/24f \
  -H "X-API-Key: your-api-key-here"
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
| 401 | Unauthorized - Missing or invalid API key |
| 403 | Forbidden - Admin access required |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Database or server error |

---

## Rate Limiting

The API implements rate limiting based on your API key configuration:

- Each API key has a configurable rate limit and time window
- Rate limits are enforced per API key
- Rate limit information is included in your API key configuration

---

## CORS Support

The API supports Cross-Origin Resource Sharing (CORS) and allows requests from any origin with the following headers:

- Access-Control-Allow-Origin: *
- Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
- Access-Control-Allow-Headers: Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key

---

## Examples in Different Languages

### JavaScript (Fetch API)

```javascript
// Get all CS courses for Fall 2024
fetch('http://localhost:8080/api/v1/courses/24f?prefix=cs', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
})
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));
```

### Python (requests)

```python
import requests

# Get all CS courses for Fall 2024
headers = {'X-API-Key': 'your-api-key-here'}
response = requests.get('http://localhost:8080/api/v1/courses/24f',
                       params={'prefix': 'cs'},
                       headers=headers)
data = response.json()
print(data)
```

### cURL

```bash
# Get all available terms
curl http://localhost:8080/api/v1/terms/ \
  -H "X-API-Key: your-api-key-here"

# Get all courses for a term
curl http://localhost:8080/api/v1/courses/24f \
  -H "X-API-Key: your-api-key-here"

# Search for courses
curl "http://localhost:8080/api/v1/courses/24f/search?q=Computer Science" \
  -H "X-API-Key: your-api-key-here"

# Get schools for a term
curl http://localhost:8080/api/v1/schools/24f \
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
