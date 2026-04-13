package types

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
