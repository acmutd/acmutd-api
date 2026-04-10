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

// Course represents a course stored in Firestore
//
// Firestore Structure:
//   - courses/{course_prefix}/numbers/{course_number}
type CourseGeneralInfo struct {
	Prefix            string                       `firestore:"course_prefix"`
	Number            string                       `firestore:"course_number"`
	Title             string                       `firestore:"title"`
	Description       string                       `firestore:"description"`
	CreditHours       int                          `firestore:"credit_hours"`
	School            string                       `firestore:"school"`
	SchoolID          string                       `firestore:"school_id"`
	Core              string                       `firestore:"core"`
	ScheduleFrequency string                       `firestore:"schedule_frequency"`
	SectionTypes      []string                     `firestore:"section_types"`
	Grades            map[string]GradeDistribution `firestore:"grades"`
	LastUpdatedTerm   string                       `firestore:"last_updated_term"`

	// Sections holds the associated section documents for this course.
	// Populated during the combine step; not written to Firestore.
	Sections []*SectionDoc `firestore:"-"`
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
type SectionDoc struct {
	// Core identifiers (normalized to lowercase)
	SectionAddress string `json:"section_address" firestore:"section_address"`
	Prefix         string `json:"prefix" firestore:"prefix"`
	Number         string `json:"number" firestore:"number"`
	Section        string `json:"section" firestore:"section"`
	Term           string `json:"term" firestore:"term"`

	// Course metadata
	Title       string `json:"title" firestore:"title"`
	Subtitle    string `json:"subtitle" firestore:"subtitle"`
	Description string `json:"description" firestore:"description"`
	School      string `json:"school" firestore:"school"`
	SchoolID    string `json:"school_id" firestore:"school_id"`
	Core        string `json:"core" firestore:"core"`

	CreditHours     int      `json:"credit_hours" firestore:"credit_hours"`
	ActivityType    string   `json:"activity_type" firestore:"activity_type"`
	CrossListed     []string `json:"cross_listed" firestore:"cross_listed"`
	EnrollmentReqs  []string `json:"enrollment_reqs" firestore:"enrollment_reqs"`
	ClassAttributes []string `json:"class_attributes" firestore:"class_attributes"`

	ClassID         string `json:"class_id" firestore:"class_id"`
	ClassLevel      string `json:"class_level" firestore:"class_level"`
	InstructionMode string `json:"instruction_mode" firestore:"instruction_mode"`
	Grading         string `json:"grading" firestore:"grading"`
	SessionType     string `json:"session_type" firestore:"session_type"`
	AddConsent      string `json:"add_consent" firestore:"add_consent"`
	ClassNotes      string `json:"class_notes" firestore:"class_notes"`
	SyllabusID      string `json:"syllabus_id" firestore:"syllabus_id"`

	// Enrollment information
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

	OrionDatetime     string `json:"orion_datetime" firestore:"orion_datetime"`
	ScheduleFrequency string `json:"schedule_frequency" firestore:"schedule_frequency"`

	Grades *GradeDistribution `json:"grades,omitempty" firestore:"grades,omitempty"`
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

// SectionQuery contains parameters for querying SectionDoc documents
type SectionQuery struct {
	Term       string // required; e.g., "23f"
	Prefix     string // e.g., "cs"
	Number     string // e.g., "2305"
	Instructor string // exact match on an element of the instructors array
	Days       string // exact match on an element of the days array (e.g., "Monday")
	Limit      int
	Offset     int
}

// GeneralCourseQuery contains parameters for querying CourseGeneralInfo documents.
type GeneralCourseQuery struct {
	Prefix string // required; scopes query to courses/{prefix}/numbers
	Search string // substring match on title or description
	Limit  int
	Offset int
}
