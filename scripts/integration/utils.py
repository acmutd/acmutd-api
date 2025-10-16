"""
Utility functions for the integration scraper.
Shared functions for name normalization and data processing.
"""

import re


def normalize_name(name):
    """Normalizes names, removes periods, handles middle names, replaces hyphens, and potential swaps."""
    name = name.strip()
    name = re.sub(r"\s*,\s*", ", ", name)  # standardize comma spacing
    name = re.sub(r"\s+[A-Z](\.[A-Z])*\s*$", "", name)  # remove middle initials
    name = re.sub(r"([A-Z])\.([A-Z])", r"\1 \2", name)  # add space between initials
    name = re.sub(r"[.\s]+", " ", name)  # removes periods and extra spaces
    name = re.sub(r"['’ʻ`]", "", name)  # remove apostrophes
    name = name.replace('-', ' ')  # replace hyphens with spaces

    if ", " in name:  # handle the Last, First formats by splitting up and swapping
        last, first = name.split(", ", 1)
        return f"{first.strip().lower()} {last.strip().lower()}"
    else:
        return name.strip().lower()


def extract_first_instructor(instructor_string, instructor_id_string):
    """Extracts the first instructor's name and ID from strings."""
    names = [normalize_name(name.strip()) for name in instructor_string.split(",")]
    ids = [id.strip() for id in instructor_id_string.split(",")]
    if names and ids:
        return names[0], ids[0]
    return None, None


def extract_course_department(course_code):
    """Extracts the department from a course code."""
    match = re.match(r"([A-Z]+)\d+", course_code)
    if match:
        return match.group(1)
    return None


def generate_name_variations(name):
    """Generates variations of a name by trying different combinations of parts."""
    parts = name.split()
    variations = {name}  # include the og name
    
    # based on number of parts in the name, include different variations for various cases
    if len(parts) >= 2:
        variations.add(f"{parts[1]} {parts[0]}")  # swap first and last (Bhadrachalam Chitturi --> Chitturi Bhadrachalam and Mohammed Ali --> Ali Mohammed)
        
        if len(parts) >= 3:
            variations.add(f"{parts[0]} {parts[-1]}")  # first and last (skip middle)
            variations.add(f"{parts[0]} {parts[1]}")  # first and second (e.g. Carlos Busso Recabarren --> Carlos Busso)
            variations.add(f"{parts[-1]} {parts[0]}")  # last and first
            variations.add(" ".join(parts[1:]))  # remove first name
            variations.add(" ".join(parts[:-1]))  # remove last name
            
            if len(parts) >= 4:
                variations.add(f"{parts[0]} {parts[2]}")  # first and third
                variations.add(f"{parts[0]} {parts[-2]} {parts[-1]}")  # first and last two
                variations.add(f"{parts[0]} {parts[-3]} {parts[-2]}")  # first and middle-last two
                variations.add(f"{parts[0]} {parts[-3]}")  # first and third-from-last

    return variations


def check_course_overlap(rmp_info, ratings_info):
    """Checks for course overlap between RMP and ratings data."""
    rmp_courses = set(rmp_info.get("courses", []))
    ratings_courses = set(ratings_info.get("course_ratings", {}).keys())

    rmp_headers = {extract_course_department(course) for course in rmp_courses if extract_course_department(course)}
    ratings_headers = {extract_course_department(course) for course in ratings_courses if extract_course_department(course)}

    rmp_numbers = {re.sub(r'[^\d]', '', course) for course in rmp_courses}
    ratings_numbers = {re.sub(r'[^\d]', '', course) for course in ratings_courses}

    return rmp_courses.intersection(ratings_courses) or rmp_headers.intersection(ratings_headers) or rmp_numbers.intersection(ratings_numbers)