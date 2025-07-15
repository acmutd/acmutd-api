package firebase

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/types"
	"google.golang.org/api/iterator"
)

type Firestore struct {
	*firestore.Client
}

func NewFirestore(ctx context.Context, app *firebase.App) (*Firestore, error) {

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore client: %w", err)
	}

	return &Firestore{
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
func (c *Firestore) InsertClassesWithIndexes(ctx context.Context, courses []types.Course, term string) {
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

func (c *Firestore) InsertTerms(ctx context.Context, terms []string) {
	batch := c.BulkWriter(ctx)

	for _, term := range terms {
		doc := c.Collection("classes").Doc(term)
		batch.Set(doc, map[string]any{
			"term": term,
		})
	}

	batch.End()
}

func (c *Firestore) QueryAllTerms(ctx context.Context) ([]string, error) {
	query := c.Collection("classes")
	iter := query.Documents(ctx)

	var terms []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get next document: %w", err)
		}
		terms = append(terms, doc.Ref.ID)
	}

	return terms, nil
}
func (c *Firestore) QueryByCourseNumber(ctx context.Context, term, coursePrefix, courseNumber string) ([]types.Course, error) {
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

func (c *Firestore) QueryByCoursePrefix(ctx context.Context, term, coursePrefix string) ([]types.Course, error) {
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

// GetAllCoursesByTerm returns all courses for a given term
func (c *Firestore) GetAllCoursesByTerm(ctx context.Context, term string) ([]types.Course, error) {
	query := c.Collection("classes").Doc(term).Collection("courses")
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

// QueryBySchool returns courses by school for a given term
func (c *Firestore) QueryBySchool(ctx context.Context, term, school string) ([]types.Course, error) {
	query := c.Collection("classes").Doc(term).Collection("courses").Where("school", "==", school)
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

// SearchCourses searches courses by title, topic, or instructor name
func (c *Firestore) SearchCourses(ctx context.Context, term, searchQuery string) ([]types.Course, error) {
	// TODO: Figure out a nicer way to do this
	courses, err := c.GetAllCoursesByTerm(ctx, term)
	if err != nil {
		return nil, err
	}

	var filteredCourses []types.Course
	query := strings.ToLower(searchQuery)

	for _, course := range courses {
		// Search in title, topic, and instructors
		if strings.Contains(strings.ToLower(course.Title), query) ||
			strings.Contains(strings.ToLower(course.Topic), query) ||
			strings.Contains(strings.ToLower(course.Instructors), query) {
			filteredCourses = append(filteredCourses, course)
		}
	}

	return filteredCourses, nil
}

// GetSchoolsByTerm returns all schools for a given term
func (c *Firestore) GetSchoolsByTerm(ctx context.Context, term string) ([]string, error) {
	query := c.Collection("classes").Doc(term).Collection("indexes")
	iter := query.Documents(ctx)

	var schools []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get next document: %w", err)
		}
		schools = append(schools, doc.Ref.ID)
	}

	return schools, nil
}

func (c *Firestore) CreateAPIKey(ctx context.Context, apiKey *types.APIKey) error {
	doc := c.Collection("api_keys").Doc(apiKey.ID)
	_, err := doc.Set(ctx, apiKey)
	return err
}

func (c *Firestore) GetAPIKeyByKey(ctx context.Context, key string) (*types.APIKey, error) {
	query := c.Collection("api_keys").Where("key", "==", key)
	iter := query.Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("api key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get api key: %w", err)
	}

	var apiKey types.APIKey
	if err := doc.DataTo(&apiKey); err != nil {
		return nil, fmt.Errorf("failed to convert document to api key: %w", err)
	}

	return &apiKey, nil
}

func (c *Firestore) GetAPIKeyByID(ctx context.Context, id string) (*types.APIKey, error) {
	doc := c.Collection("api_keys").Doc(id)
	snapshot, err := doc.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get api key: %w", err)
	}

	var apiKey types.APIKey
	if err := snapshot.DataTo(&apiKey); err != nil {
		return nil, fmt.Errorf("failed to convert document to api key: %w", err)
	}

	return &apiKey, nil
}

func (c *Firestore) UpdateAPIKey(ctx context.Context, apiKey *types.APIKey) error {
	doc := c.Collection("api_keys").Doc(apiKey.ID)
	_, err := doc.Set(ctx, apiKey)
	return err
}

func (c *Firestore) GetAllAPIKeys(ctx context.Context) ([]types.APIKey, error) {
	query := c.Collection("api_keys")
	iter := query.Documents(ctx)

	var apiKeys []types.APIKey
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get next document: %w", err)
		}

		var apiKey types.APIKey
		if err := doc.DataTo(&apiKey); err != nil {
			return nil, fmt.Errorf("failed to convert document to api key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

func (c *Firestore) DeleteAPIKey(ctx context.Context, id string) error {
	doc := c.Collection("api_keys").Doc(id)
	_, err := doc.Delete(ctx)
	return err
}
