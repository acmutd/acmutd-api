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

// GradeDistribution holds the letter-grade counts for a section.
type GradeDistribution struct {
	APlus  int `json:"A+" firestore:"A+"`
	A      int `json:"A" firestore:"A"`
	AMinus int `json:"A-" firestore:"A-"`
	BPlus  int `json:"B+" firestore:"B+"`
	B      int `json:"B" firestore:"B"`
	BMinus int `json:"B-" firestore:"B-"`
	CPlus  int `json:"C+" firestore:"C+"`
	C      int `json:"C" firestore:"C"`
	CMinus int `json:"C-" firestore:"C-"`
	DPlus  int `json:"D+" firestore:"D+"`
	D      int `json:"D" firestore:"D"`
	DMinus int `json:"D-" firestore:"D-"`
	F      int `json:"F" firestore:"F"`
	NF     int `json:"NF" firestore:"NF"`
	CR     int `json:"CR" firestore:"CR"`
	I      int `json:"I" firestore:"I"`
	NC     int `json:"NC" firestore:"NC"`
	P      int `json:"P" firestore:"P"`
	W      int `json:"W" firestore:"W"`
}

// RawMeeting represents a meeting block from the scraper JSON output.
type RawMeeting struct {
	DateRange string   `json:"date_range"`
	Days      []string `json:"days"`
	Time      string   `json:"time"`
	Location  string   `json:"location"`
}

// RawCourse represents a course section as output by the coursebook scraper.
type RawCourse struct {
	SectionAddress      string       `json:"section_address"`
	CoursePrefix        string       `json:"course_prefix"`
	CourseNumber        string       `json:"course_number"`
	Section             string       `json:"section"`
	ClassCourseNumber   string       `json:"class_course_number"`
	ClassLevel          string       `json:"class_level"`
	InstructionMode     string       `json:"instruction_mode"`
	Title               string       `json:"title"`
	Description         string       `json:"description"`
	EnrolledStatus      string       `json:"enrolled_status"`
	EnrolledCurrent     int          `json:"enrolled_current"`
	EnrolledMax         int          `json:"enrolled_max"`
	Waitlist            int          `json:"waitlist"`
	Term                string       `json:"term"`
	StartDate           string       `json:"start_date"`
	EndDate             string       `json:"end_date"`
	Meetings            []RawMeeting `json:"meetings"`
	ActivityType        string       `json:"activity_type"`
	SemesterCreditHours string       `json:"semester_credit_hours"`
	Core                *string      `json:"core"`
	Grading             string       `json:"grading"`
	SessionType         *string      `json:"session_type"`
	AddConsent          string       `json:"add_consent"`
	OrionDatetime       string       `json:"orion_datetime"`
	ScheduleFrequency   *string      `json:"schedule_frequency"`
	EnrollmentReqs      []string     `json:"enrollment_reqs"`
	ClassAttributes     []string     `json:"class_attributes"`
	ClassNotes          *string      `json:"class_notes"`
	Instructors         []string     `json:"instructors"`
	InstructorIDs       []string     `json:"instructor_ids"`
	TAs                 []string     `json:"tas"`
	TAIDs               []string     `json:"ta_ids"`
	School              string       `json:"school"`
	SchoolID            string       `json:"school_id"`
	CrossListed         []string     `json:"cross_listed"`
	Syllabus            *string      `json:"syllabus"`
}

// SectionDoc represents a course section as stored in Firestore.
// Path: courses/{prefix}/numbers/{number}/sections/{section_address}
// Includes denormalized course fields so collectionGroup queries
// return complete data without reading the parent doc.
type SectionDoc struct {
	SectionAddress string `json:"section_address" firestore:"section_address"`
	Prefix         string `json:"prefix" firestore:"prefix"`
	Number         string `json:"number" firestore:"number"`
	Section        string `json:"section" firestore:"section"`
	Term           string `json:"term" firestore:"term"`

	Title       string  `json:"title" firestore:"title"`
	Subtitle    *string `json:"subtitle,omitempty" firestore:"subtitle"`
	Description string  `json:"description" firestore:"description"`
	School      string  `json:"school" firestore:"school"`
	SchoolID    string  `json:"school_id" firestore:"school_id"`
	Core        *string `json:"core" firestore:"core"`

	CreditHours     int      `json:"credit_hours" firestore:"credit_hours"`
	ActivityType    string   `json:"activity_type" firestore:"activity_type"`
	CrossListed     []string `json:"cross_listed" firestore:"cross_listed"`
	EnrollmentReqs  []string `json:"enrollment_reqs" firestore:"enrollment_reqs"`
	ClassAttributes []string `json:"class_attributes" firestore:"class_attributes"`

	ClassID         *string `json:"class_id" firestore:"class_id"`
	ClassLevel      string  `json:"class_level" firestore:"class_level"`
	InstructionMode string  `json:"instruction_mode" firestore:"instruction_mode"`
	Grading         string  `json:"grading" firestore:"grading"`
	SessionType     *string `json:"session_type" firestore:"session_type"`
	AddConsent      string  `json:"add_consent" firestore:"add_consent"`
	ClassNotes      *string `json:"class_notes" firestore:"class_notes"`
	SyllabusID      *string `json:"syllabus_id" firestore:"syllabus_id"`

	EnrolledStatus  string `json:"enrolled_status" firestore:"enrolled_status"`
	EnrolledCurrent int    `json:"enrolled_current" firestore:"enrolled_current"`
	EnrolledMax     int    `json:"enrolled_max" firestore:"enrolled_max"`
	Waitlist        int    `json:"waitlist" firestore:"waitlist"`

	StartDate string   `json:"start_date" firestore:"start_date"`
	EndDate   string   `json:"end_date" firestore:"end_date"`
	Days      []string `json:"days" firestore:"days"`
	Time      string   `json:"time" firestore:"time"`
	Location  string   `json:"location" firestore:"location"`
	Building  string   `json:"building" firestore:"building"`
	Room      string   `json:"room" firestore:"room"`

	Instructors   []string `json:"instructors" firestore:"instructors"`
	InstructorIDs []string `json:"instructor_ids" firestore:"instructor_ids"`
	TAs           []string `json:"tas" firestore:"tas"`
	TAIDs         []string `json:"ta_ids" firestore:"ta_ids"`

	OrionDatetime     string  `json:"orion_datetime" firestore:"orion_datetime"`
	ScheduleFrequency *string `json:"schedule_frequency" firestore:"schedule_frequency"`

	Grades *GradeDistribution `json:"grades,omitempty" firestore:"grades,omitempty"`
}

// Course represents a course section stored in Firestore, combining coursebook
// data with grade distribution.
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
	Syllabus string `json:"syllabus" firestore:"syllabus"` // Syllabus Code or content
}

type CourseGeneralInfo struct {
	Prefix            string                       `firestore:"course_prefix"`
	Number            string                       `firestore:"course_number"`
	CreditHours       int                          `firestore:"credit_hours"`
	School            string                       `firestore:"school"`
	SchoolID          string                       `firestore:"school_id"`
	Core              *string                      `firestore:"core,omitempty"`
	ScheduleFrequency *string                      `firestore:"schedule_frequency"`
	SectionTypes      []string                     `firestore:"section_types"`
	Grades            map[string]GradeDistribution `firestore:"grades"`
}

// CourseQuery contains parameters for querying courses
type CourseQuery struct {
	Term         string // e.g., "23f"
	CoursePrefix string // e.g., "cs"
	CourseNumber string // e.g., "2305"
	Section      string // e.g., "001"
	School       string // School code
	Instructor   string // Filter by instructor name
	InstructorID string // Filter by instructor ID
	Days         string // Filter by days (e.g., "Monday, Wednesday")
	Times        string // Filter by times (24h format, e.g., "14:00 - 14:50")
	Times12h     string // Filter by times (12h format, e.g., "2:00 PM - 2:50 PM")
	Location     string // Filter by location (e.g., "SCI_1.210")
	Search       string // Search query for title, topic, instructors
	Limit        int    // Max results to return (0 for no limit)
	Offset       int    // Number of results to skip
}
