package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/types"
	"google.golang.org/api/iterator"
)

type Client struct {
	*firestore.Client
}

func NewClient(ctx context.Context, app *firebase.App) (*Client, error) {

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore client: %w", err)
	}

	return &Client{
		Client: client,
	}, nil
}

/*
Structure:

  - classes/{term}/courses/{coursePrefix-courseNumber-courseSection}

  - classes/{term}/indexes/{school}

    This will allow us to query for all courses by school, course prefix, course number, and course section.

    THIS WILL OVERWRITE ALL PREVIOUS DATA IN THE COLLECTION.
*/
func (c *Client) InsertClassesWithIndexes(ctx context.Context, courses []types.Course, term string) {
	batch := c.BulkWriter(ctx)

	// Group courses by school for indexing
	schoolGroups := make(map[string][]types.Course)

	for _, course := range courses {
		courseID := fmt.Sprintf("%s-%s-%s", course.CoursePrefix, course.CourseNumber, course.Section)

		// Store individual course directly as struct
		doc := c.Collection("classes").Doc(term).Collection("courses").Doc(courseID)
		batch.Set(doc, course)

		// Group by school for indexing
		school := course.CoursePrefix
		schoolGroups[school] = append(schoolGroups[school], course)
	}

	// Create index documents
	for school, schoolCourses := range schoolGroups {
		indexDoc := c.Collection("classes").Doc(term).Collection("indexes").Doc(school)
		batch.Set(indexDoc, map[string]any{
			"courses": schoolCourses,
			"count":   len(schoolCourses),
		})
	}

	batch.End()
}

type CourseIndex struct {
	Count   int            `json:"count" firestore:"count"`
	Courses []types.Course `json:"courses" firestore:"courses"`
}

func (c *Client) QueryByCourseNumber(ctx context.Context, term, coursePrefix, courseNumber string) ([]types.Course, error) {
	// Query individual courses directly - much more efficient
	// Note: You'll need to create a composite index on (course_prefix, course_number) for this query
	query := c.Collection("classes").Doc(term).Collection("courses").Where("course_prefix", "==", coursePrefix).Where("course_number", "==", courseNumber)
	iter := query.Documents(ctx)

	var courses []types.Course
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get next document: %w", err)
		}

		var course types.Course
		if err := doc.DataTo(&course); err != nil {
			return nil, fmt.Errorf("failed to convert document to course: %w", err)
		}
		courses = append(courses, course)
	}

	return courses, nil
}

// Alternative: Query by course prefix only (if you want all courses in a prefix)
func (c *Client) QueryByCoursePrefix(ctx context.Context, term, coursePrefix string) ([]types.Course, error) {
	query := c.Collection("classes").Doc(term).Collection("courses").Where("course_prefix", "==", coursePrefix)
	iter := query.Documents(ctx)

	var courses []types.Course
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get next document: %w", err)
		}

		var course types.Course
		if err := doc.DataTo(&course); err != nil {
			return nil, fmt.Errorf("failed to convert document to course: %w", err)
		}
		courses = append(courses, course)
	}

	return courses, nil
}
