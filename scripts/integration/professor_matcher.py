"""
Matcher module for fuzzy matching RMP data with professor ratings.
Handles direct matches, fuzzy matching, and course overlap validation.
"""

import json
import os
from rapidfuzz import fuzz
from rapidfuzz.process import cdist
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


def remove_entry(data_dict, key, id_field, id_value):
    """Remove the entry matching id_value from data_dict[key], deleting the key if empty."""
    if key not in data_dict:
        return
    data_dict[key] = [e for e in data_dict[key] if e.get(id_field) != id_value]
    if not data_dict[key]:
        del data_dict[key]


def apply_manual_matches(ratings, rmp_data, matched_data, normalized_ratings, normalized_rmp_data, norm_to_original_rmp):
    """Applies manual matches from a JSON file, normalizing names before matching."""
    manual_matches_file = "manual_matches.json"

    if not os.path.exists(manual_matches_file):
        print(f"{manual_matches_file} not found. Manual matches will be skipped.")
        return

    try:
        with open(manual_matches_file, "r", encoding="utf-8") as f:
            manual_matches = json.load(f)
    except (json.JSONDecodeError, FileNotFoundError) as e:
        print(f"Error loading manual matches: {e}. Manual matches will be skipped.")
        return

    print(f"Applying {len(manual_matches)} manual matches...")

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

                original_rmp_name = norm_to_original_rmp.get(rmp_name)
                if original_ratings_name in ratings and original_rmp_name in rmp_data:
                    remove_entry(ratings, original_ratings_name, "instructor_id", matched_entry.get("instructor_id"))
                    remove_entry(rmp_data, original_rmp_name, "rmp_id", matched_entry.get("rmp_id"))
                    print(f"Manual match applied: {original_ratings_name} -> {original_rmp_name}")
                else:
                    print(f"Manual match failed: Could not find entries in source dictionaries.")
            else:
                print(f"Manual match failed: No matching courses found for {ratings_name} -> {rmp_name}")
        else:
            print(f"Manual match failed: {ratings_name} or {rmp_name} not found.")


def match_professor_names(ratings, rmp_data, fuzzy_threshold=80):
    """Matches professor data using direct and fuzzy matching with name variations."""
    matched_data = {}
    ratings_to_append = list(ratings.keys())

    normalized_ratings = {normalize_name(name): (name, data) for name, data in ratings.items()}
    normalized_rmp_data = {normalize_name(name): data for name, data in rmp_data.items()}

    # O(1) reverse lookup: normalized rmp name -> original rmp name
    norm_to_original_rmp = {normalize_name(name): name for name in rmp_data}

    total_ratings_entries = sum(len(data_list) for _, data_list in normalized_ratings.values())
    total_rmp_entries = sum(len(rmp_list) for rmp_list in normalized_rmp_data.values())
    print(f"Matching {total_ratings_entries} grade ratings entries to {total_rmp_entries} RateMyProfessors entries...")

    # Pre-compute name variations for every RMP entry once
    rmp_norm_variations = {rmp_norm: generate_name_variations(rmp_norm) for rmp_norm in normalized_rmp_data}

    # Build flat index: variation_string -> rmp_norm, plus a list for batch fuzzy scoring
    rmp_variation_to_norm: dict[str, str] = {}
    for rmp_norm, variations in rmp_norm_variations.items():
        for var in variations:
            rmp_variation_to_norm[var] = rmp_norm
    all_rmp_variation_strings = list(rmp_variation_to_norm.keys())

    apply_manual_matches(ratings, rmp_data, matched_data, normalized_ratings, normalized_rmp_data, norm_to_original_rmp)

    direct_match_count = 0
    matched_rmp_norms: set[str] = set()

    # Direct matches first
    for rmp_norm in list(normalized_rmp_data.keys()):
        if rmp_norm not in normalized_ratings:
            continue

        original_ratings_name, ratings_list = normalized_ratings[rmp_norm]
        rmp_list = normalized_rmp_data[rmp_norm]
        matched_entry = process_direct_match(ratings_list, rmp_list)

        if matched_entry:
            if original_ratings_name not in matched_data:
                matched_data[original_ratings_name] = []
            matched_data[original_ratings_name].append(matched_entry)

            original_rmp_name = norm_to_original_rmp[rmp_norm]
            if original_ratings_name in ratings and original_rmp_name in rmp_data:
                remove_entry(ratings, original_ratings_name, "instructor_id", matched_entry.get("instructor_id"))
                remove_entry(rmp_data, original_rmp_name, "rmp_id", matched_entry.get("rmp_id"))
                matched_rmp_norms.add(rmp_norm)
                direct_match_count += 1

    print(f"Direct Matches: {direct_match_count}")
    print(f"Remaining Ratings to Fuzzy Match: {len(ratings)}")

    # Fuzzy matching for remaining entries.
    # Build a flat list of (original_name, variation) for all unmatched ratings entries.
    remaining_names = list(ratings.keys())
    ratings_var_index: list[tuple[str, str]] = []  # (original_name, variation_string)
    for name in remaining_names:
        norm = normalize_name(name)
        for var in generate_name_variations(norm):
            ratings_var_index.append((name, var))

    if ratings_var_index and all_rmp_variation_strings:
        query_strings = [var for _, var in ratings_var_index]

        # rows = ratings variations, cols = rmp variations
        score_matrix = cdist(query_strings, all_rmp_variation_strings, scorer=fuzz.ratio, score_cutoff=fuzzy_threshold)

        # Aggregate: for each ratings professor, find the highest-scoring RMP variation
        best_per_name: dict[str, tuple[float, str]] = {}  # name -> (score, rmp_norm)
        for i, (original_name, _) in enumerate(ratings_var_index):
            row = score_matrix[i]
            if not row.any():
                continue
            best_j = int(row.argmax())
            score = float(row[best_j])
            rmp_norm = rmp_variation_to_norm[all_rmp_variation_strings[best_j]]
            if rmp_norm in matched_rmp_norms:
                continue
            prev_score, _ = best_per_name.get(original_name, (0, ""))
            if score > prev_score:
                best_per_name[original_name] = (score, rmp_norm)
    else:
        best_per_name = {}

    for original_ratings_name, (best_score, best_match_rmp_norm) in best_per_name.items():
        if original_ratings_name not in ratings or best_score < fuzzy_threshold:
            continue
        ratings_list = ratings[original_ratings_name]
        ratings_info = ratings_list[0]

        if best_match_rmp_norm and best_match_rmp_norm in normalized_rmp_data:
            best_rmp_match = None
            best_rmp_score = 0
            for rmp_info in normalized_rmp_data[best_match_rmp_norm]:
                if check_course_overlap(rmp_info, ratings_info):
                    score = rmp_info.get("ratings_count", 0)
                    if score > best_rmp_score:
                        best_rmp_score = score
                        best_rmp_match = rmp_info

            if best_rmp_match:
                rmp_info_cleaned = {k: v for k, v in best_rmp_match.items() if k != "courses"}
                if original_ratings_name not in matched_data:
                    matched_data[original_ratings_name] = []
                matched_data[original_ratings_name].append({**rmp_info_cleaned, **ratings_info})

                original_rmp_name = norm_to_original_rmp[best_match_rmp_norm]
                if original_ratings_name in ratings and original_rmp_name in rmp_data:
                    remove_entry(ratings, original_ratings_name, "instructor_id", ratings_info.get("instructor_id"))
                    remove_entry(rmp_data, original_rmp_name, "rmp_id", best_rmp_match.get("rmp_id"))
                    matched_rmp_norms.add(best_match_rmp_norm)

    # Append unmatched ratings data to final output
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
