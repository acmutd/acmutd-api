import json
from fuzzywuzzy import fuzz
from fuzzywuzzy import process
import argparse
import time
import re
import os
from scraper import scrape_rmp_data
from aggregator import calculate_professor_ratings, normalize_name

def extract_course_department(course_code):
    """Extracts the department from a course code."""
    match = re.match(r"([A-Z]+)\d+", course_code)
    if match:
        return match.group(1)
    return None


def generate_name_variations(name):
    """Generates variations of a name by trying different combinations of parts."""
    parts = name.split()
    variations = {name}  # include the og name as a variation as well
    if len(parts) > 1:
        variations.add(f"{parts[1]} {parts[0]}") # swap last and first name, this one is for you Bhadrachalam Chitturi --> Chitturi Bhadrachalam and Mohammed Ali --> Ali Mohammed

        if len(parts) > 2:
            # multi-part last name handling
            variations.add(f"{parts[0]} {parts[-1]}") # first and last
            variations.add(f"{parts[0]} {parts[1]}") # first and second part of name, this one is for you Carlos Busso Recabarren --> Carlos Busso
            variations.add(f"{parts[-1]} {parts[0]}") # last and first

            if len(parts) > 3:
                variations.add(f"{parts[0]} {parts[2]}") # first and first part of last name, this one is for you Andres Ricardo Sanchez De La Rosa --> Andres Sanchez
                variations.add(" ".join(parts[1:])) # remove first name.
                variations.add(" ".join(parts[:-1])) # remove last name.
                variations.add(f"{parts[0]} {parts[-2]} {parts[-1]}")  # first and last two parts
                variations.add(f"{parts[0]} {parts[-3]} {parts[-2]} {parts[-1]}") # first and last 3 parts
                variations.add(f"{parts[0]} {parts[-3]} {parts[-2]}") # first and last 2 parts minus the last
                variations.add(f"{parts[0]} {parts[-3]}") # first and first part of the last name

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


# direct match is when the names are exactly the same, or when the names are effectively the same after normalization
def process_direct_match(ratings_list, rmp_list):
    """Processes a direct match and returns the matched data."""
    if len(ratings_list) == 1 and len(rmp_list) == 1: # if there's only one entry in each list, we can assume they are the same person and match them directly
        rmp_info_cleaned = {k: v for k, v in rmp_list[0].items() if k != "courses"}
        return {**rmp_info_cleaned, **ratings_list[0]}

    # if there are multiple entries, we need to find the most likely match based on the courses taught by each and the number of ratings (sometimes the same prof has multiple RMP profiles so this selects the most used one effectively)
    best_rmp_match = None
    best_ratings_match = None
    best_rmp_score = 0

    for ratings_info in ratings_list:
        for rmp_info in rmp_list:
            if check_course_overlap(rmp_info, ratings_info):
                score = rmp_info.get("ratings_count", 0)
                if score > best_rmp_score:
                    best_rmp_score = score
                    best_rmp_match = rmp_info
                    best_ratings_match = ratings_info

    if best_rmp_match:
        rmp_info_cleaned = {k: v for k, v in best_rmp_match.items() if k != "courses"} # remove the RMP course list from the final data since the courses are already in the ratings data
        return {**rmp_info_cleaned, **best_ratings_match}

    return None


def remove_matched_entries(matched_ratings_entry, matched_rmp_entry, ratings, rmp_data):
    """Removes the specific matched entries from ratings and rmp_data."""
    for ratings_key, ratings_list in list(ratings.items()):
        ratings[ratings_key] = [entry for entry in ratings_list if entry.get("instructor_id") != matched_ratings_entry.get("instructor_id")] # use the instructor_id to remove the proper entry from the list of profs with that name
        if not ratings[ratings_key]:
            del ratings[ratings_key]
    for rmp_key, rmp_list in list(rmp_data.items()):
        rmp_data[rmp_key] = [entry for entry in rmp_list if entry.get("rmp_id") != matched_rmp_entry.get("rmp_id")] # use the rmp_id to remove the proper entry from the list of profs with that name
        if not rmp_data[rmp_key]:
            del rmp_data[rmp_key]


# applies manual matches from a JSON file, i.e. Yu Chung Ng is Vincent Ng in RMP so that matching is done from deliberate user input
def apply_manual_matches(ratings, rmp_data, matched_data, normalized_ratings, normalized_rmp_data):
    """Applies manual matches from a JSON file, normalizing names before matching."""
    try:
        with open("manual_matches.json", "r", encoding="utf-8") as f:
            manual_matches = json.load(f)
    except FileNotFoundError:
        print("manual_matches.json not found. Manual matches will be skipped.")
        return

    # a lot of this logic is the exact same as the direct match logic but refactoring it into a function is a bit tough because of the minor differences between them
    for match in manual_matches:
        ratings_name = normalize_name(match["ratings_name"])
        rmp_name = normalize_name(match["rmp_name"])

        if ratings_name in normalized_ratings and rmp_name in normalized_rmp_data:
            original_ratings_name, ratings_list = normalized_ratings[ratings_name]
            rmp_list = normalized_rmp_data[rmp_name]

            matched_entry = process_direct_match(ratings_list, rmp_list)

            if matched_entry:
                if original_ratings_name not in matched_data:
                    matched_data[original_ratings_name] = []
                matched_data[original_ratings_name].append(matched_entry)
                original_rmp_name = None
                for original_name, norm_data in rmp_data.items():
                    if normalize_name(original_name) == rmp_name:
                        original_rmp_name = original_name
                        break

                if original_ratings_name in ratings and original_rmp_name in rmp_data:
                    remove_matched_entries(matched_entry, matched_entry, ratings, rmp_data)
                    print(f"Manual match applied: {original_ratings_name} -> {original_rmp_name}")
                else:
                    print(f"Manual match failed: Could not find entries in source dictionaries.")
            else:
                print(f"Manual match failed: No matching courses found for {ratings_name} -> {rmp_name}")
        else:
            print(f"Manual match failed: {ratings_name} or {rmp_name} not found.")


# main match logic driver function
def match_professor_names(ratings, rmp_data, fuzzy_threshold=80):
    """Matches professor data, handles name variations, and saves unmatched names."""
    matched_data = {}
    ratings_to_append = list(ratings.keys())
    matched_names = set()

    normalized_ratings = {normalize_name(name): (name, data) for name, data in ratings.items()}
    normalized_rmp_data = {normalize_name(name): data for name, data in rmp_data.items()}

    apply_manual_matches(ratings, rmp_data, matched_data, normalized_ratings, normalized_rmp_data) # apply manual matches before processing

    total_ratings_entries = sum(len(data_list) for _, data_list in normalized_ratings.values())
    total_rmp_entries = sum(len(rmp_list) for _, rmp_list in normalized_rmp_data.items())
    print(f"Now matching {total_ratings_entries} grade ratings entries to {total_rmp_entries} RateMyProfessors entries...")

    direct_match_count = 0

    for rmp_norm, rmp_list in normalized_rmp_data.items():
        if rmp_norm in normalized_ratings:
            original_ratings_name, ratings_list = normalized_ratings[rmp_norm]
            matched_entry = process_direct_match(ratings_list, rmp_list)

            if matched_entry:
                if original_ratings_name not in matched_data:
                    matched_data[original_ratings_name] = []
                matched_data[original_ratings_name].append(matched_entry)
                original_rmp_name = None
                for original_name, norm_data in rmp_data.items():
                    if normalize_name(original_name) == rmp_norm:
                        original_rmp_name = original_name
                        break

                if original_rmp_name is None:
                    print(f"Warning: Original RMP name not found for normalized name {rmp_norm}.")
                    continue

                if original_ratings_name in ratings and original_rmp_name in rmp_data:
                    remove_matched_entries(matched_entry, matched_entry, ratings, rmp_data)
                    matched_names.add(original_ratings_name)
                    direct_match_count += 1

    print(f"Direct Matches: {direct_match_count}")
    print(f"Remaining Ratings to Fuzzy Match: {len(normalized_ratings)}, now matching...")

    for original_ratings_name, ratings_list in list(ratings.items()):
        if original_ratings_name not in ratings:
            continue
        ratings_norm = normalize_name(original_ratings_name)
        best_match = None
        best_score = 0

        ratings_info = ratings_list[0]

        for rmp_norm, rmp_list in normalized_rmp_data.items():
            for ratings_variation in generate_name_variations(ratings_norm):
                for rmp_variation in generate_name_variations(rmp_norm):
                    score = fuzz.ratio(ratings_variation, rmp_variation)

                    if score > best_score and score >= fuzzy_threshold:
                        best_score = score
                        best_match = rmp_norm

        if best_match:
            best_rmp_match = None
            best_rmp_score = 0

            for rmp_info in normalized_rmp_data[best_match]:
                if check_course_overlap(rmp_info, ratings_info):
                    score = rmp_info.get("ratings_count", 0)
                    if score > best_rmp_score:
                        best_rmp_score = score
                        best_rmp_match = rmp_info

            if best_rmp_match:
                rmp_info_cleaned = {k: v for k, v in best_rmp_match.items() if k not in ["courses"]}
                if original_ratings_name not in matched_data:
                    matched_data[original_ratings_name] = []
                matched_data[original_ratings_name].append({**rmp_info_cleaned, **ratings_info})
                original_rmp_name = None
                for original_name, norm_data in rmp_data.items():
                    if normalize_name(original_name) == best_match:
                        original_rmp_name = original_name
                        break

                if original_rmp_name is None:
                    print(f"Warning: Original RMP name not found for normalized name {best_match}.")
                    continue

                if original_ratings_name in ratings and original_rmp_name in rmp_data:
                    remove_matched_entries(ratings_info, best_rmp_match, ratings, rmp_data)
                    matched_names.add(original_ratings_name)
                else:
                    print(f"Fuzzy match rejected for {original_ratings_name} due to no matching courses.")
                    remove_matched_entries(ratings_info, best_rmp_match, ratings, rmp_data)
            else:
                print(f"Fuzzy match rejected for {original_ratings_name} due to no matching RMP professor with shared courses.")
        else:
            print(f"Fuzzy match rejected for {original_ratings_name} due to no name matches found.")

    matched_professors_count = len(matched_data) # this is an estimate because it doesnt count the elements in the lists, just the keys so profs with the same name are considered 1
    print(f"Matched Professors: {matched_professors_count}")

    # append the unmatched ratings data to the final matched data using original names
    for original_ratings_name in ratings_to_append:
        if original_ratings_name in ratings:
            if original_ratings_name not in matched_data:
                matched_data[original_ratings_name] = ratings[original_ratings_name]
            else:
                matched_data[original_ratings_name].extend(ratings[original_ratings_name])

    print(f"Unmatched Ratings: {len(ratings)}")
    print(f"Unmatched RMP: {len(rmp_data)}")

    total_professors = len(matched_data)
    print(f"Total professors in data: {total_professors}") # this is an estimate because it doesnt count the elements in the lists, just the keys so profs with the same name are considered 1

    with open("unmatched/unmatched_ratings.json", "w", encoding="utf-8") as f:
        json.dump(ratings, f, indent=4, ensure_ascii=False)

    with open("unmatched/unmatched_rmp.json", "w", encoding="utf-8") as f:
        json.dump(rmp_data, f, indent=4, ensure_ascii=False)

    return matched_data


def main():
    parser = argparse.ArgumentParser(description="Professor Data Matching Script")
    parser.add_argument("mode", nargs="?", default="normal", choices=["normal", "reload"], help="Execution mode: normal or reload")
    args = parser.parse_args()
    total_start_time = time.time()

    os.makedirs("ratings", exist_ok=True)
    os.makedirs("unmatched", exist_ok=True)
    os.makedirs("matched", exist_ok=True)


    if args.mode == "reload": # load existing data if it exists and matches it
        print("Loading professor ratings data...")
        with open("ratings/grade_ratings.json", "r", encoding="utf-8") as file:
            ratings = json.load(file)

        print("Loading RateMyProfessors data...")
        with open("ratings/rmp_ratings.json", "r", encoding="utf-8") as file:
            rmp_data = json.load(file)

        print("Matching professor data from both sources...")
        matched_data = match_professor_names(ratings, rmp_data)

        with open("matched/matched_professor_data.json", "w", encoding="utf-8") as outfile:
            json.dump(matched_data, outfile, indent=4, ensure_ascii=False)

        print(f"Matched professor data saved to matched/matched_professor_data.json")

    else: # scrape RMP data and recalculates professor ratings before running the resulting data
        print("Calculating professor ratings...")
        ratings = calculate_professor_ratings()

        print("Scraping professor data from RateMyProfessors...")
        rmp_data = scrape_rmp_data(university_id="1273")

        print("Matching professor data from both sources...")
        matched_data = match_professor_names(ratings, rmp_data)

        with open("matched/matched_professor_data.json", "w", encoding="utf-8") as outfile:
            json.dump(matched_data, outfile, indent=4, ensure_ascii=False)

        print(f"Matched professor data saved to matched/matched_professor_data.json")

    total_end_time = time.time()
    print(f"Total execution complete in {total_end_time - total_start_time:.2f} seconds.")


if __name__ == "__main__":
    main()