"""
Mapper module for mapping grades data to coursebook sections and instructor IDs.
Handles the mapping of grade CSV rows to coursebook sections and instructor extraction.
"""

import csv
import os
from utils import normalize_name


def create_section_lookup(coursebook_data):
    """Creates a lookup dictionary for section data."""
    section_lookup = {}
    
    for section in coursebook_data:
        subject = section.get("course_prefix", "").upper()
        course_number = section.get("course_number", "")
        section_number = section.get("section_number", "")
        
        # Create key that matches grades CSV format: SUBJECT,COURSE_NUMBER,SECTION
        key = f"{subject},{course_number},{section_number}"
        section_lookup[key] = section
    
    return section_lookup


def create_instructor_id_lookup(matched_professor_data):
    """Creates lookup dictionary with instructor ID as key."""
    instructor_lookup = {}
    
    for professor_name, professor_list in matched_professor_data.items():
        for professor_entry in professor_list:
            instructor_id = professor_entry.get("instructor_id")
            if instructor_id:
                instructor_lookup[instructor_id] = professor_entry
    
    return instructor_lookup


def map_grades_to_instructors(grades_files, coursebook_data, matched_professor_data):
    """Maps grade CSV rows to coursebook sections and extracts instructor IDs."""
    section_lookup = create_section_lookup(coursebook_data)
    enhanced_grades = []
    
    # Create instructor lookup by ID from matched data
    instructor_by_id = {}
    for professor_name, professor_list in matched_professor_data.items():
        for professor_entry in professor_list:
            instructor_id = professor_entry.get("instructor_id")
            if instructor_id:
                instructor_by_id[instructor_id] = professor_entry

    for filepath in grades_files:
        with open(filepath, "r", encoding="utf-8-sig") as csvfile:
            print(f"Mapping grades in {os.path.basename(filepath)}...")
            reader = csv.DictReader(csvfile)
            
            for row in reader:
                # Create lookup key
                subject = row.get("Subject", "").strip()
                catalog_nbr = row.get('"Catalog Nbr"') or row.get("Catalog Nbr", "")
                catalog_nbr = catalog_nbr.strip()
                section = row.get("Section", "").strip()
                
                lookup_key = f"{subject},{catalog_nbr},{section}"
                
                # Enhanced row with instructor information
                enhanced_row = dict(row)
                enhanced_row["instructor_id"] = ""
                enhanced_row["instructor_name_normalized"] = ""
                enhanced_row["has_rmp_data"] = False
                
                # Try to find matching section in coursebook
                if lookup_key in section_lookup:
                    section_data = section_lookup[lookup_key]
                    instructor_ids = section_data.get("instructor_ids", "")
                    
                    if instructor_ids:
                        # Get first instructor ID
                        first_instructor_id = instructor_ids.split(",")[0].strip()
                        enhanced_row["instructor_id"] = first_instructor_id
                        
                        # Check if we have professor data for this instructor
                        if first_instructor_id in instructor_by_id:
                            instructor_data = instructor_by_id[first_instructor_id]
                            enhanced_row["instructor_name_normalized"] = normalize_name(
                                instructor_data.get("original_rmp_format", "")
                            )
                            enhanced_row["has_rmp_data"] = True
                else:
                    # Try to match by instructor name fallback
                    instructor_1 = row.get("Instructor 1", "").strip()
                    if instructor_1:
                        normalized_instructor = normalize_name(instructor_1)
                        enhanced_row["instructor_name_normalized"] = normalized_instructor
                        
                        # Try to find in matched professor data
                        for professor_name, professor_list in matched_professor_data.items():
                            if normalize_name(professor_name) == normalized_instructor:
                                if professor_list and len(professor_list) > 0:
                                    instructor_id = professor_list[0].get("instructor_id", "")
                                    enhanced_row["instructor_id"] = instructor_id
                                    enhanced_row["has_rmp_data"] = True
                                    break
                
                enhanced_grades.append(enhanced_row)
    
    return enhanced_grades