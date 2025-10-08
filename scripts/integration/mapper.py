"""
Mapper module for mapping grades data to coursebook sections and instructor IDs.
Handles the mapping of grade CSV rows to coursebook sections and instructor extraction.
"""

import csv
import os
from utils import normalize_name


def create_section_lookup(coursebook_data):
    """Creates a lookup dictionary for section data using section_address."""
    section_lookup = {}
    
    for section in coursebook_data:
        section_address = section.get("section_address", "").lower()
        if section_address:
            section_lookup[section_address] = section
    
    return section_lookup


def find_instructor_id_by_section_address(section_lookup, subject, catalog_nbr, section):
    """Finds instructor ID using section address matching (more reliable approach)."""
    # Build section key: e.g., ACCT2301.002 (case-insensitive)
    key = f"{subject}{catalog_nbr}.{section}".lower()
    
    # Look for section address that starts with our key
    for section_address, section_data in section_lookup.items():
        if section_address.startswith(key):
            instructor_ids = section_data.get('instructor_ids', '').split(',')
            return instructor_ids[0].strip() if instructor_ids and instructor_ids[0].strip() else ''
    
    return ''


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
    
    # Stats tracking
    total_grades = 0
    section_matches = 0
    fallback_matches = 0
    no_matches = 0
    
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
                total_grades += 1
                
                # Extract grade data fields
                subject = row.get("Subject", "").strip().upper()
                catalog_nbr = row.get('"Catalog Nbr"') or row.get("Catalog Nbr", "")
                catalog_nbr = catalog_nbr.strip()
                section = row.get("Section", "").strip()
                
                # Enhanced row with instructor information
                enhanced_row = dict(row)
                enhanced_row["instructor_id"] = ""
                enhanced_row["instructor_name_normalized"] = ""
                enhanced_row["has_rmp_data"] = False
                
                # Try section address lookup (more reliable method)
                instructor_id = find_instructor_id_by_section_address(section_lookup, subject, catalog_nbr, section)
                enhanced_row["instructor_id"] = instructor_id
                
                # Check if we have professor data for this instructor
                if instructor_id and instructor_id in instructor_by_id:
                    section_matches += 1
                    instructor_data = instructor_by_id[instructor_id]
                    enhanced_row["instructor_name_normalized"] = normalize_name(
                        instructor_data.get("original_rmp_format", "")
                    )
                    enhanced_row["has_rmp_data"] = True
                elif not instructor_id:
                    # Fallback: Try to match by instructor name if no section address match found
                    instructor_1 = row.get("Instructor 1", "").strip()
                    if instructor_1:
                        normalized_instructor = normalize_name(instructor_1)
                        enhanced_row["instructor_name_normalized"] = normalized_instructor
                        
                        # Try to find in matched professor data
                        matched_fallback = False
                        for professor_name, professor_list in matched_professor_data.items():
                            if normalize_name(professor_name) == normalized_instructor:
                                if professor_list and len(professor_list) > 0:
                                    fallback_instructor_id = professor_list[0].get("instructor_id", "")
                                    enhanced_row["instructor_id"] = fallback_instructor_id
                                    enhanced_row["has_rmp_data"] = True
                                    matched_fallback = True
                                    fallback_matches += 1
                                    break
                        
                        if not matched_fallback:
                            no_matches += 1
                    else:
                        no_matches += 1
                else:
                    # Have instructor_id but no RMP data
                    enhanced_row["instructor_name_normalized"] = normalize_name(
                        row.get("Instructor 1", "")
                    )
                    enhanced_row["has_rmp_data"] = False
                    no_matches += 1
                
                enhanced_grades.append(enhanced_row)
    
    # Print mapping statistics
    print(f"\n--- Instructor Mapping Statistics ---")
    print(f"Total grades processed: {total_grades}")
    print(f"Section address matches: {section_matches} ({section_matches/total_grades*100:.1f}%)")
    print(f"Fallback name matches: {fallback_matches} ({fallback_matches/total_grades*100:.1f}%)")
    print(f"No matches found: {no_matches} ({no_matches/total_grades*100:.1f}%)")
    print(f"Total matched: {section_matches + fallback_matches} ({(section_matches + fallback_matches)/total_grades*100:.1f}%)")
    
    return enhanced_grades