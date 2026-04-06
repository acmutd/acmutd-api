package uploader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/acmutd/acmutd-api/internal/types"
)

// loadAndCombine reads coursebook JSON and enhanced grades CSV for the given terms,
// then combines entries by section_address into SectionDoc objects ready for Firestore.
func loadAndCombine(inputBase string, terms []string) (map[string]*types.SectionDoc, map[string]map[string]*types.CourseGeneralInfo, error) {
	allSections := make(map[string]*types.SectionDoc)
	allCourses := make(map[string]map[string]*types.CourseGeneralInfo)

	for _, term := range terms {
		log.Printf("Combining data for term %s", term)

		sections, courses, err := loadCoursebookTerm(inputBase, term)
		if err != nil {
			log.Printf("Warning: could not load coursebook for term %s: %v", term, err)
		}

		grades, err := loadGradesTerm(inputBase, term)
		if err != nil {
			log.Printf("Warning: could not load grades for term %s: %v", term, err)
		}

		combined, err := mergeGrades(sections, grades, term)
		if err != nil {
			log.Printf("Warning: could not merge grades for term %s: %v", term, err)
			continue
		}

		for addr, sec := range combined {
			allSections[addr] = sec
		}

		// Keep latest course info but accumulate grades across terms
		for prefix, numbers := range courses {
			if _, exists := allCourses[prefix]; !exists {
				allCourses[prefix] = make(map[string]*types.CourseGeneralInfo)
			}
			for number, course := range numbers {
				if existing, exists := allCourses[prefix][number]; exists {
					prevGrades := existing.Grades
					allCourses[prefix][number] = course
					if course.Grades == nil {
						course.Grades = prevGrades
					} else if prevGrades != nil {
						for k, v := range prevGrades {
							course.Grades[k] = v
						}
					}
				} else {
					allCourses[prefix][number] = course
				}
			}
		}
	}

	return allSections, allCourses, nil
}

func mergeGrades(sections map[string]*types.SectionDoc, grades map[string]types.GradeDistribution, term string) (map[string]*types.SectionDoc, error) {
	combined := make(map[string]*types.SectionDoc, len(sections))
	for addr, sec := range sections {
		combined[addr] = sec
	}

	for addr, gd := range grades {
		gd := gd
		if entry, ok := combined[addr]; ok {
			entry.Grades = &gd
		} else {
			log.Printf("Warning: grade record for %s has no matching course section", addr)
			combined[addr] = &types.SectionDoc{
				SectionAddress: addr,
				Term:           term,
				Grades:         &gd,
			}
		}
	}
	return combined, nil
}

// loadCoursebookTerm reads a classes_{term}.json file, parses RawCourse entries,
// and converts them into SectionDoc objects keyed by section_address.
func loadCoursebookTerm(inputBase, term string) (map[string]*types.SectionDoc, map[string]map[string]*types.CourseGeneralInfo, error) {
	filePath := filepath.Join(inputBase, "coursebook", fmt.Sprintf("classes_%s.json", term))

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var rawCourses []types.RawCourse
	if err := json.Unmarshal(data, &rawCourses); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	sections := make(map[string]*types.SectionDoc, len(rawCourses))
	courses := make(map[string]map[string]*types.CourseGeneralInfo)
	for i := range rawCourses {
		rc := &rawCourses[i]
		addr := strings.ToLower(strings.TrimSpace(rc.SectionAddress))
		if addr == "" {
			continue
		}

		sec := rawToSectionDoc(rc)
		sections[addr] = sec

		if _, exists := courses[sec.Prefix]; !exists {
			courses[sec.Prefix] = make(map[string]*types.CourseGeneralInfo)
		}
		if _, exists := courses[sec.Prefix][sec.Number]; !exists {
			courses[sec.Prefix][sec.Number] = &types.CourseGeneralInfo{
				Prefix:            sec.Prefix,
				Number:            sec.Number,
				CreditHours:       sec.CreditHours,
				School:            sec.School,
				SchoolID:          sec.SchoolID,
				Core:              sec.Core,
				ScheduleFrequency: sec.ScheduleFrequency,
				SectionTypes:      []string{},
			}
		}

		if sec.Subtitle != nil {
			course := courses[sec.Prefix][sec.Number]
			subtitleVal := *sec.Subtitle

			found := false
			for _, st := range course.SectionTypes {
				if st == subtitleVal {
					found = true
					break
				}
			}
			if !found {
				course.SectionTypes = append(course.SectionTypes, subtitleVal)
			}
		}
	}

	log.Printf("Loaded %d courses for term %s", len(sections), term)
	return sections, courses, nil
}

// rawToSectionDoc transforms a RawCourse (scraper output) into a SectionDoc (Firestore document).
func rawToSectionDoc(rc *types.RawCourse) *types.SectionDoc {
	classID, _ := splitClassCourseNumber(rc.ClassCourseNumber)
	_, subtitle := splitTitle(rc.Title)
	creditHours := parseCreditHours(rc.SemesterCreditHours, rc.ActivityType)

	// just grab first meeting for simplicity
	var days []string
	var time, location, building, room string
	if len(rc.Meetings) > 0 {
		m := rc.Meetings[0]
		days = m.Days
		time = m.Time
		location = m.Location
		building, room = splitLocation(m.Location)
	}

	var classIDPtr *string
	if classID != "" {
		classIDPtr = &classID
	}

	return &types.SectionDoc{
		SectionAddress: strings.ToLower(strings.TrimSpace(rc.SectionAddress)),
		Prefix:         strings.ToLower(strings.TrimSpace(rc.CoursePrefix)),
		Number:         strings.ToLower(strings.TrimSpace(rc.CourseNumber)),
		Section:        strings.ToLower(strings.TrimSpace(rc.Section)),
		Term:           strings.ToLower(strings.TrimSpace(rc.Term)),

		Title:       rc.Title,
		Subtitle:    subtitle,
		Description: rc.Description,
		School:      rc.School,
		SchoolID:    strings.ToLower(strings.TrimSpace(rc.SchoolID)),
		Core:        rc.Core,

		CreditHours:     creditHours,
		ActivityType:    rc.ActivityType,
		CrossListed:     rc.CrossListed,
		EnrollmentReqs:  rc.EnrollmentReqs,
		ClassAttributes: rc.ClassAttributes,

		ClassID:         classIDPtr,
		ClassLevel:      rc.ClassLevel,
		InstructionMode: rc.InstructionMode,
		Grading:         rc.Grading,
		SessionType:     rc.SessionType,
		AddConsent:      rc.AddConsent,
		ClassNotes:      rc.ClassNotes,
		SyllabusID:      rc.Syllabus,

		EnrolledStatus:  rc.EnrolledStatus,
		EnrolledCurrent: rc.EnrolledCurrent,
		EnrolledMax:     rc.EnrolledMax,
		Waitlist:        rc.Waitlist,

		StartDate: rc.StartDate,
		EndDate:   rc.EndDate,
		Days:      days,
		Time:      time,
		Location:  location,
		Building:  building,
		Room:      room,

		Instructors:   rc.Instructors,
		InstructorIDs: rc.InstructorIDs,
		TAs:           rc.TAs,
		TAIDs:         rc.TAIDs,

		OrionDatetime:     rc.OrionDatetime,
		ScheduleFrequency: rc.ScheduleFrequency,
	}
}

// splitClassCourseNumber splits "21466 / 016508" into classID and courseID.
func splitClassCourseNumber(s string) (classID, courseID string) {
	parts := strings.SplitN(s, " / ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(s), ""
}

// splitTitle splits a title like "Linear Algebra - Problem Section" into main title and a subtitle. Only splits on the last " - " if the right
func splitTitle(title string) (string, *string) {
	idx := strings.LastIndex(title, " - ")
	if idx < 0 {
		return title, nil
	}
	right := strings.TrimSpace(title[idx+3:])
	if len(right) > 50 || right == "" {
		return title, nil
	}
	left := strings.TrimSpace(title[:idx])
	return left, &right
}

// splitLocation splits "JO 4.614" into building="JO" and room="4.614".
func splitLocation(loc string) (building, room string) {
	loc = strings.TrimSpace(loc)
	if loc == "" || strings.EqualFold(loc, "No Meeting Room") {
		return "", ""
	}
	parts := strings.SplitN(loc, " ", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return loc, ""
}

// parseCreditHours parses semester credit hours string to int.
// Returns 0 for non-enroll activity types.
func parseCreditHours(raw string, activityType string) int {
	if strings.Contains(strings.ToLower(activityType), "non-enroll") {
		return 0
	}
	v, _ := strconv.Atoi(strings.TrimSpace(raw))
	return v
}

// loadGradesTerm reads an enhanced_grades_{term}.csv file and returns GradeDistribution entries keyed by section_address.
func loadGradesTerm(inputBase, term string) (map[string]types.GradeDistribution, error) {
	filePath := filepath.Join(inputBase, "enhanced_grades", fmt.Sprintf("enhanced_grades_%s.csv", term))

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	headerIdx := make(map[string]int)
	for i, h := range headers {
		headerIdx[strings.TrimSpace(strings.Trim(h, "\""))] = i
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV records: %w", err)
	}

	grades := make(map[string]types.GradeDistribution, len(records))
	for _, row := range records {
		subject := strings.ToLower(strings.TrimSpace(getField(row, headerIdx, "Subject")))
		catalogNbr := strings.ToLower(strings.TrimSpace(getField(row, headerIdx, "Catalog Nbr")))
		section := strings.ToLower(strings.TrimSpace(getField(row, headerIdx, "Section")))

		if subject == "" || catalogNbr == "" {
			continue
		}

		addr := fmt.Sprintf("%s%s.%s.%s", subject, catalogNbr, section, term)

		gd := types.GradeDistribution{
			APlus:  atoi(getField(row, headerIdx, "A+")),
			A:      atoi(getField(row, headerIdx, "A")),
			AMinus: atoi(getField(row, headerIdx, "A-")),
			BPlus:  atoi(getField(row, headerIdx, "B+")),
			B:      atoi(getField(row, headerIdx, "B")),
			BMinus: atoi(getField(row, headerIdx, "B-")),
			CPlus:  atoi(getField(row, headerIdx, "C+")),
			C:      atoi(getField(row, headerIdx, "C")),
			CMinus: atoi(getField(row, headerIdx, "C-")),
			DPlus:  atoi(getField(row, headerIdx, "D+")),
			D:      atoi(getField(row, headerIdx, "D")),
			DMinus: atoi(getField(row, headerIdx, "D-")),
			F:      atoi(getField(row, headerIdx, "F")),
			NF:     atoi(getField(row, headerIdx, "NF")),
			CR:     atoi(getField(row, headerIdx, "CR")),
			I:      atoi(getField(row, headerIdx, "I")),
			NC:     atoi(getField(row, headerIdx, "NC")),
			P:      atoi(getField(row, headerIdx, "P")),
			W:      atoi(getField(row, headerIdx, "W")),
		}

		grades[addr] = gd
	}

	log.Printf("Loaded %d grade records for term %s", len(grades), term)
	return grades, nil
}

func getField(row []string, headerIdx map[string]int, field string) string {
	if idx, ok := headerIdx[field]; ok && idx < len(row) {
		return row[idx]
	}
	return ""
}

func atoi(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// CSV values are floats like "14.0", parse as float then truncate
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int(f)
}
