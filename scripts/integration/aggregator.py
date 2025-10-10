"""
Aggregator module for processing grades and coursebook data.
Handles the aggregation of grade distributions across semesters.
"""

import csv
import os
from utils import normalize_name, extract_first_instructor


def process_section_data(coursebook_data):
    """Processes section data to create a name-based professor mapping."""
    professor_name_map = {}
    
    for section in coursebook_data:
        instructor_names = section.get("instructors", "")
        instructor_ids = section.get("instructor_ids", "")
        instructor_name, instructor_id = extract_first_instructor(instructor_names, instructor_ids)
        course = f"{section['course_prefix'].upper()}{section['course_number']}"

        if instructor_name:
            if instructor_name not in professor_name_map:
                professor_name_map[instructor_name] = []

            # check if the instructor_id already exists for this name
            found = False
            for prof in professor_name_map[instructor_name]:
                if prof["instructor_id"] == instructor_id:
                    if "courses" not in prof:
                        prof["courses"] = set()
                    prof["courses"].add(course)
                    found = True
                    break

            if not found:
                professor_name_map[instructor_name].append({
                    "instructor_id": instructor_id,
                    "courses": {course}
                })

    # convert sets to lists before serialization
    for instructor_name, profiles in professor_name_map.items():
        for profile in profiles:
            if "courses" in profile:
                profile["courses"] = list(profile["courses"])

    return professor_name_map


def calculate_professor_ratings_from_grades(grades_files, coursebook_data):
    """Calculates professor ratings based on grade distributions from CSV files."""
    professor_data = {}
    professor_name_map = process_section_data(coursebook_data)
    print("Professor data retrieved from coursebook sections, processing grade data...")
    
    grade_values = {
        "A+": 4.0, "A": 4.0, "A-": 3.67, "B+": 3.33, "B": 3.0, "B-": 2.67,
        "C+": 2.33, "C": 2.0, "C-": 1.67, "D+": 1.33, "D": 1.00, "D-": 0.67,
        "F": 0.0, "W": 0.67, "P": 4.0, "NP": 0.0
    }

    try:
        print("Aggregating grade data from files...")
        for filepath in grades_files:
            with open(filepath, "r", encoding="utf-8-sig") as csvfile:
                # print(f"Aggregating {os.path.basename(filepath)}...")
                reader = csv.DictReader(csvfile)
                for row in reader:
                    instructor = normalize_name(row.get("Instructor 1", ""))
                    subject = row.get("Subject", "").strip()
                    catalog_nbr = row.get('"Catalog Nbr"') or row.get("Catalog Nbr", "")
                    catalog_nbr = catalog_nbr.strip()
                    course = f"{subject}{catalog_nbr}"
                    row_grades = {grade: int(float(row.get(grade, 0) or 0)) for grade in grade_values}
                    
                    if not instructor or not subject or not catalog_nbr or sum(row_grades.values()) == 0:
                        continue

                    if instructor in professor_name_map:
                        profiles = professor_name_map[instructor]
                        for profile in profiles:
                            if course in profile["courses"]:
                                instructor_id = profile["instructor_id"]
                                if instructor_id not in professor_data:
                                    professor_data[instructor_id] = {"course_grades": {}}
                                if course not in professor_data[instructor_id]["course_grades"]:
                                    professor_data[instructor_id]["course_grades"][course] = {g: 0 for g in grade_values}
                                for grade, count in row_grades.items():
                                    professor_data[instructor_id]["course_grades"][course][grade] += count

    except Exception as e:
        print("Error processing grade data:", e)
        return None

    # Convert to final format matching the original aggregator output
    filtered_data = {}
    for instructor_id, data in professor_data.items():
        all_grades = {}
        for course, grades in data["course_grades"].items():
            for grade, count in grades.items():
                all_grades[grade] = all_grades.get(grade, 0) + count
        
        total_points = sum(grade_values[grade] * count for grade, count in all_grades.items())
        total_count = sum(all_grades.values())
        overall_rating = round((total_points / total_count) / 4.0 * 5, 2) if total_count > 0 else "N/A"
        
        course_ratings = {}
        for course, grades in data["course_grades"].items():
            course_points = sum(grade_values[grade] * count for grade, count in grades.items())
            course_count = sum(grades.values())
            course_ratings[course] = round((course_points / course_count) / 4.0 * 5, 2) if course_count > 0 else "N/A"

        instructor_name = next((name for name, profs in professor_name_map.items() 
                              if any(prof['instructor_id'] == instructor_id for prof in profs)), None)

        # it would be nice to have the original instructor name attached as a property, but we have to normalize it because the grades data is aggregated across
        # multiple sections and the names are not always consistent (i.e. John Cole vs John P Cole)
        if instructor_name:
            if instructor_name not in filtered_data:
                filtered_data[instructor_name] = []
            filtered_data[instructor_name].append({
                "instructor_id": instructor_id,
                "overall_grade_rating": overall_rating,
                "total_grade_count": total_count,
                "course_ratings": course_ratings,
            })
    
    # identify names with multiple IDs
    for name, profiles in filtered_data.items():
        if len(profiles) > 1:
            print(f"Instructor name '{name}' has multiple associated IDs:")
            for profile in profiles:
                print(f"  - ID: {profile['instructor_id']}")

    return filtered_data