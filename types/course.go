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

type Course struct {
	SectionAddress  string `json:"section_address" firestore:"section_address"`
	CoursePrefix    string `json:"course_prefix" firestore:"course_prefix"`
	CourseNumber    string `json:"course_number" firestore:"course_number"`
	Section         string `json:"section" firestore:"section"`
	ClassNumber     string `json:"class_number" firestore:"class_number"`
	Title           string `json:"title" firestore:"title"`
	Topic           string `json:"topic" firestore:"topic"`
	EnrolledStatus  string `json:"enrolled_status" firestore:"enrolled_status"`
	EnrolledCurrent string `json:"enrolled_current" firestore:"enrolled_current"`
	EnrolledMax     string `json:"enrolled_max" firestore:"enrolled_max"`
	Instructors     string `json:"instructors" firestore:"instructors"`
	Assistants      string `json:"assistants" firestore:"assistants"`
	Term            string `json:"term" firestore:"term"`
	Session         string `json:"session" firestore:"session"`
	Days            string `json:"days" firestore:"days"`
	Times           string `json:"times" firestore:"times"`
	Times12h        string `json:"times_12h" firestore:"times_12h"`
	Location        string `json:"location" firestore:"location"`
	CoreArea        string `json:"core_area" firestore:"core_area"`
	ActivityType    string `json:"activity_type" firestore:"activity_type"`
	School          School `json:"school" firestore:"school"`
	Dept            string `json:"dept" firestore:"dept"`
	Syllabus        string `json:"syllabus" firestore:"syllabus"`
	Textbooks       string `json:"textbooks" firestore:"textbooks"`
	InstructorIDs   string `json:"instructor_ids" firestore:"instructor_ids"`
}
