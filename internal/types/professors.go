package types

type Professor struct {
	InstructorID             string             `json:"instructor_id" firestore:"instructor_id"`
	NormalizedCoursebookName string             `json:"normalized_coursebook_name" firestore:"normalized_coursebook_name"`
	OriginalRMPFormat        string             `json:"original_rmp_format" firestore:"original_rmp_format"`
	Department               string             `json:"department" firestore:"department"`
	URL                      string             `json:"url" firestore:"url"`
	QualityRating            float64            `json:"quality_rating" firestore:"quality_rating"`
	DifficultyRating         float64            `json:"difficulty_rating" firestore:"difficulty_rating"`
	WouldTakeAgain           int                `json:"would_take_again" firestore:"would_take_again"`
	RatingsCount             int                `json:"ratings_count" firestore:"ratings_count"`
	Tags                     []string           `json:"tags" firestore:"tags"`
	RMPID                    string             `json:"rmp_id" firestore:"rmp_id"`
	OverallGradeRating       float64            `json:"overall_grade_rating" firestore:"overall_grade_rating"`
	TotalGradeCount          int                `json:"total_grade_count" firestore:"total_grade_count"`
	CourseRatings            map[string]float64 `json:"course_ratings" firestore:"course_ratings"`
}
