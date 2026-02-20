package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type School string

// Need to do this because school field can be 999 for some reason
func (s *School) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = School(str)
		return nil
	}

	// If that fails, try as number
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*s = School(strconv.Itoa(num))
		return nil
	}

	return fmt.Errorf("school must be a string or number, got: %s", string(data))
}

// MarshalJSON ensures School is always serialized as a string
func (s School) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s School) String() string {
	return string(s)
}

// Course represents a course section stored in Firestore.
//
// Firestore Structure:
//   - courses/{course_prefix}/numbers/{course_number}/sections/{section_address}
//
// Courses are stored in a hierarchical structure for efficient queries:
//   - course_prefix and course_number are normalized to lowercase
//   - section_address is a unique identifier (e.g., "cs2305.001.23f")
//   - term field enables collection group queries across all sections
//
// Indexes Required:
//   - Collection group "sections" with term field (for term-based queries)
//   - Composite indexes for term+course_prefix, term+course_number queries
//
// Related Collections:
//   - terms/{term}/prefixes/{course_prefix} - metadata for available prefixes per term
type Course struct {
	// Core identifiers (normalized to lowercase during ingestion)
	SectionAddress string `json:"section_address" firestore:"section_address"` // Unique ID: {prefix}{number}.{section}.{term}
	CoursePrefix   string `json:"course_prefix" firestore:"course_prefix"`     // e.g., "cs" (normalized lowercase)
	CourseNumber   string `json:"course_number" firestore:"course_number"`     // e.g., "2305" (normalized lowercase)
	Section        string `json:"section" firestore:"section"`                 // e.g., "001" (normalized lowercase)
	Term           string `json:"term" firestore:"term"`                       // e.g., "23f" (normalized lowercase, indexed for collection group queries)

	// Course metadata
	ClassNumber string `json:"class_number" firestore:"class_number"` // UTD class number
	Title       string `json:"title" firestore:"title"`               // Course title
	Topic       string `json:"topic" firestore:"topic"`               // Special topics course name

	// Enrollment information
	EnrolledStatus  string `json:"enrolled_status" firestore:"enrolled_status"`   // "Open", "Closed", "Waitlist"
	EnrolledCurrent string `json:"enrolled_current" firestore:"enrolled_current"` // Current enrollment count
	EnrolledMax     string `json:"enrolled_max" firestore:"enrolled_max"`         // Maximum enrollment

	// Instructor information
	Instructors   string `json:"instructors" firestore:"instructors"`       // Comma-separated instructor names
	InstructorIDs string `json:"instructor_ids" firestore:"instructor_ids"` // Comma-separated instructor IDs (links to professors collection)
	Assistants    string `json:"assistants" firestore:"assistants"`         // Teaching assistants

	// Schedule information
	Session  string `json:"session" firestore:"session"`     // Session identifier
	Days     string `json:"days" firestore:"days"`           // Days of the week (e.g., "Monday, Wednesday")
	Times    string `json:"times" firestore:"times"`         // 24-hour time format
	Times12h string `json:"times_12h" firestore:"times_12h"` // 12-hour time format
	Location string `json:"location" firestore:"location"`   // Building and room

	// Academic categorization
	CoreArea     string `json:"core_area" firestore:"core_area"`         // Core curriculum area
	ActivityType string `json:"activity_type" firestore:"activity_type"` // Lecture, Lab, etc.
	School       School `json:"school" firestore:"school"`               // School code (can be string or number)
	Dept         string `json:"dept" firestore:"dept"`                   // Department

	// Additional resources
	Syllabus  string `json:"syllabus" firestore:"syllabus"`   // Syllabus URL or content
	Textbooks string `json:"textbooks" firestore:"textbooks"` // Required textbooks
}

// CourseQuery contains parameters for querying courses
type CourseQuery struct {
	Term         string // e.g., "23f"
	CoursePrefix string // e.g., "cs"
	CourseNumber string // e.g., "2305"
	Section      string // e.g., "001"
	School       string // School code
	Search       string // Search query for title, topic, instructors
	Limit        int    // Max results to return (0 for no limit)
	Offset       int    // Number of results to skip
}
