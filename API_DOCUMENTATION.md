# ACM API Documentation

## Overview

The ACM API provides access to course and school data for the University of Texas at Dallas. All endpoints return JSON responses and support CORS for cross-origin requests.

**Base URL**: `http://localhost:8080` (or your deployed server URL)

## Authentication

The ACM API uses API key authentication for all endpoints except the health check. You must include your API key in the `Authorization` header for all requests.

### Getting an API Key

API keys can be created by administrators using the admin endpoints. Contact your system administrator to obtain an API key.

### Using API Keys

Include your API key in the `Authorization` header:

```bash
Authorization: Bearer YOUR_API_KEY_HERE
```

**Example:**

```bash
curl -H "Authorization: Bearer abc123def456..." http://localhost:8080/api/v1/courses/2024FALL
```

### Rate Limiting

Each API key has configurable rate limits:

- **Rate Limit**: Maximum number of requests per time window
- **Rate Interval**: Time window duration (e.g., 30 minutes)

Rate limit information is included in response headers:

- `X-RateLimit-Limit`: Maximum requests per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: When the current window resets (RFC3339 format)

When rate limit is exceeded, you'll receive a `429 Too Many Requests` response with rate limit information.

### Error Responses

Authentication errors return `401 Unauthorized`:

```json
{
  "error": "API key required. Include 'Authorization: Bearer YOUR_API_KEY' header"
}
```

Rate limit errors return `429 Too Many Requests`:

```json
{
  "error": "Rate limit exceeded",
  "rate_limit_info": {
    "remaining_requests": 0,
    "reset_time": "2024-01-15T10:30:00Z",
    "rate_limit": 30,
    "rate_interval": "30m0s"
  }
}
```

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
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/courses/2024FALL

# Get all CS courses for Fall 2024
curl -H "Authorization: Bearer YOUR_API_KEY" "http://localhost:8080/api/v1/courses/2024FALL?prefix=CS"

# Get CS 1337 for Fall 2024
curl -H "Authorization: Bearer YOUR_API_KEY" "http://localhost:8080/api/v1/courses/2024FALL?prefix=CS&number=1337"

# Get all ECS school courses for Fall 2024
curl -H "Authorization: Bearer YOUR_API_KEY" "http://localhost:8080/api/v1/courses/2024FALL?school=ECS"
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
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/courses/2024FALL/prefix/CS
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
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/courses/2024FALL/prefix/CS/number/1337
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
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/courses/2024FALL/school/ECS
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

## Term Endpoints

### Get All Terms

**GET** `/api/v1/terms/`

Retrieve all available academic terms in the database.

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

curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/terms/
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
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/schools/2024FALL
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

## Admin Endpoints

The admin endpoints allow administrators to manage API keys. These endpoints require admin authentication using the `X-Admin-Token` header.

### Admin Authentication

Include the admin token in the `X-Admin-Token` header:

```bash
X-Admin-Token: ADMIN_TOKEN
```

### Create API Key

**POST** `/admin/keys/`

Create a new API key with specified rate limits.

**Headers:**

- `X-Admin-Token`: Admin authentication token
- `Content-Type`: application/json

**Request Body:**

```json
{
  "rate_limit": 30,
  "rate_interval": "30m",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**Parameters:**

- `rate_limit` (required): Maximum requests per time window (1-1000)
- `rate_interval` (required): Time window duration as a string (e.g., "30m", "1h", "24h"). Must be between 1 minute and 24 hours.
- `expires_at` (optional): Expiration date (ISO 8601 format)

**Valid rate_interval formats:**

- `"1m"` - 1 minute
- `"30m"` - 30 minutes
- `"1h"` - 1 hour
- `"2h30m"` - 2 hours 30 minutes
- `"24h"` - 24 hours

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "key": "generated-api-key-here",
  "expires_at": "2024-12-31T23:59:59Z",
  "rate_limit": 30,
  "rate_interval": "30m0s",
  "is_active": true,
  "created_at": "2024-01-15T10:00:00Z",
  "usage_count": 0,
  "last_used_at": "0001-01-01T00:00:00Z"
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/admin/keys/ \
  -H "X-Admin-Token: your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"rate_limit": 30, "rate_interval": "30m"}'
```

### Get All API Keys

**GET** `/admin/keys/`

Retrieve all API keys (keys are masked for security).

**Headers:**

- `X-Admin-Token`: Admin authentication token

**Response:**

```json
{
  "count": 2,
  "keys": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "key": "***",
      "expires_at": "2024-12-31T23:59:59Z",
      "rate_limit": 30,
      "rate_interval": "30m0s",
      "is_active": true,
      "created_at": "2024-01-15T10:00:00Z",
      "usage_count": 15,
      "last_used_at": "2024-01-15T09:30:00Z"
    }
  ]
}
```

### Get API Key by ID

**GET** `/admin/keys/{id}`

Retrieve a specific API key by ID.

**Path Parameters:**

- `id` (required): API key ID

**Response:** Same format as create response, but key is masked.

### Update API Key

**PUT** `/admin/keys/{id}`

Update an existing API key's rate limits or expiration.

**Path Parameters:**

- `id` (required): API key ID

**Request Body:** Same format as create request.

**Response:** Same format as create response, but key is masked.

### Delete API Key

**DELETE** `/admin/keys/{id}`

Delete an API key.

**Path Parameters:**

- `id` (required): API key ID

**Response:**

```json
{
  "message": "API key deleted successfully"
}
```

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
fetch('http://localhost:8080/api/v1/courses/2024FALL?prefix=CS', {
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY'
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
headers = {'Authorization': 'Bearer YOUR_API_KEY'}
response = requests.get('http://localhost:8080/api/v1/courses/2024FALL',
                       params={'prefix': 'CS'},
                       headers=headers)
data = response.json()
print(data)
```

### cURL

```bash
# Get all available terms
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/terms/

# Get all courses for a term
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/courses/2024FALL

# Search for courses
curl -H "Authorization: Bearer YOUR_API_KEY" "http://localhost:8080/api/v1/courses/2024FALL/search?q=Computer Science"

# Get schools for a term
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/schools/2024FALL
```

---

## Versioning

This documentation covers API version 1 (`/api/v1/`). Future versions may introduce breaking changes and will be documented separately.

---

## Support

For issues or questions about the API, please contact the development team or create an issue in the project repository.
