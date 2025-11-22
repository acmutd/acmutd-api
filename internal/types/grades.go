package types

// Grades represents a grade distribution for a course section.
type Grades struct {
	CoursePrefix             string `json:"course_prefix" firestore:"course_prefix"`
	CourseNumber             string `json:"course_number" firestore:"course_number"`
	Term                     string `json:"term" firestore:"term"`
	InstructorID             string `json:"instructor_id" firestore:"instructor_id"`
	InstructorNameNormalized string `json:"instructor_name_normalized" firestore:"instructor_name_normalized"`
	Section                  string `json:"section" firestore:"section"`
	Instructor1              string `json:"instructor_1" firestore:"instructor_1"`
	Instructor2              string `json:"instructor_2" firestore:"instructor_2"`
	Instructor3              string `json:"instructor_3" firestore:"instructor_3"`
	Instructor4              string `json:"instructor_4" firestore:"instructor_4"`
	Instructor5              string `json:"instructor_5" firestore:"instructor_5"`
	Instructor6              string `json:"instructor_6" firestore:"instructor_6"`
	APlus                    string `json:"A+" firestore:"A+"`
	A                        string `json:"A" firestore:"A"`
	AMinus                   string `json:"A-" firestore:"A-"`
	BPlus                    string `json:"B+" firestore:"B+"`
	B                        string `json:"B" firestore:"B"`
	BMinus                   string `json:"B-" firestore:"B-"`
	CPlus                    string `json:"C+" firestore:"C+"`
	C                        string `json:"C" firestore:"C"`
	CMinus                   string `json:"C-" firestore:"C-"`
	DPlus                    string `json:"D+" firestore:"D+"`
	D                        string `json:"D" firestore:"D"`
	DMinus                   string `json:"D-" firestore:"D-"`
	F                        string `json:"F" firestore:"F"`
	NF                       string `json:"NF" firestore:"NF"`
	CR                       string `json:"CR" firestore:"CR"`
	I                        string `json:"I" firestore:"I"`
	NC                       string `json:"NC" firestore:"NC"`
	P                        string `json:"P" firestore:"P"`
	W                        string `json:"W" firestore:"W"`
}
