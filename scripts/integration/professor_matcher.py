"""
Matcher module for fuzzy matching RMP data with professor ratings.
Handles direct matches, fuzzy matching, and course overlap validation.
"""

from fuzzywuzzy import fuzz
from utils import normalize_name, generate_name_variations, check_course_overlap


def process_direct_match(ratings_list, rmp_list):
    """Processes a direct match and returns the matched data."""
    if len(ratings_list) == 1 and len(rmp_list) == 1:
        rmp_info_cleaned = {k: v for k, v in rmp_list[0].items() if k != "courses"}
        return {**rmp_info_cleaned, **ratings_list[0]}

    # if there are multiple entries, find the most likely match based on courses and ratings count
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
        rmp_info_cleaned = {k: v for k, v in best_rmp_match.items() if k != "courses"}
        return {**rmp_info_cleaned, **best_ratings_match}

    return None


def remove_matched_entries(matched_ratings_entry, matched_rmp_entry, ratings, rmp_data):
    """Removes the specific matched entries from ratings and rmp_data."""
    for ratings_key, ratings_list in list(ratings.items()):
        ratings[ratings_key] = [entry for entry in ratings_list 
                               if entry.get("instructor_id") != matched_ratings_entry.get("instructor_id")]
        if not ratings[ratings_key]:
            del ratings[ratings_key]
    
    for rmp_key, rmp_list in list(rmp_data.items()):
        rmp_data[rmp_key] = [entry for entry in rmp_list 
                            if entry.get("rmp_id") != matched_rmp_entry.get("rmp_id")]
        if not rmp_data[rmp_key]:
            del rmp_data[rmp_key]


def match_professor_names(ratings, rmp_data, fuzzy_threshold=80):
    """Matches professor data using direct and fuzzy matching with name variations."""
    matched_data = {}
    ratings_to_append = list(ratings.keys())

    normalized_ratings = {normalize_name(name): (name, data) for name, data in ratings.items()}
    normalized_rmp_data = {normalize_name(name): data for name, data in rmp_data.items()}

    total_ratings_entries = sum(len(data_list) for _, data_list in normalized_ratings.values())
    total_rmp_entries = sum(len(rmp_list) for _, rmp_list in normalized_rmp_data.items())
    print(f"Matching {total_ratings_entries} grade ratings entries to {total_rmp_entries} RateMyProfessors entries...")

    direct_match_count = 0

    # Direct matches first
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

                if original_ratings_name in ratings and original_rmp_name in rmp_data:
                    remove_matched_entries(matched_entry, matched_entry, ratings, rmp_data)
                    direct_match_count += 1

    print(f"Direct Matches: {direct_match_count}")
    print(f"Remaining Ratings to Fuzzy Match: {len(ratings)}")

    # Fuzzy matching for remaining entries
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

                if original_ratings_name in ratings and original_rmp_name in rmp_data:
                    remove_matched_entries(ratings_info, best_rmp_match, ratings, rmp_data)

    # append the unmatched ratings data to the final matched data
    for original_ratings_name in ratings_to_append:
        if original_ratings_name in ratings:
            if original_ratings_name not in matched_data:
                matched_data[original_ratings_name] = ratings[original_ratings_name]
            else:
                matched_data[original_ratings_name].extend(ratings[original_ratings_name])

    print(f"Matched Professors: {len(matched_data)}")
    print(f"Unmatched Ratings: {len(ratings)}")
    print(f"Unmatched RMP: {len(rmp_data)}")

    return matched_data