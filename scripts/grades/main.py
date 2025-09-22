from excel_to_csv import excel_to_csv
import os
import re
import sys

def get_term_name(term):
    # Converts TERM like '24s' to 'Spring 2024', etc.
    if len(term) < 2:
        return term
    year = '20' + term[:2]
    season_code = term[2].lower()
    season = {'s': 'Spring', 'u': 'Summer', 'f': 'Fall'}.get(season_code, season_code)
    return f"{season} {year}"

def extract_term_from_filename(filename):
    # extract explicit term from filename like "Spring 2024 Grades.xlsb"
    match = re.search(r'(Spring|Summer|Fall)[\s\-]+(20\d{2})', filename, re.IGNORECASE)
    if match:
        season_word = match.group(1).lower()
        season_map = {'spring': 's', 'summer': 'u', 'fall': 'f'}
        season = season_map.get(season_word, season_word[0])
        year = match.group(2)[2:]  # last two digits
        return f"{year}{season}"
    # fallback logic: look for term code like 25s, 24u, etc.
    fallback = re.search(r'(\d{2})([suf])', filename, re.IGNORECASE)
    if fallback:
        return fallback.group(1) + fallback.group(2).lower()
    return None


def main():
    excel_dir = os.path.join(os.path.dirname(__file__), 'put-excel-here')
    if not os.path.isdir(excel_dir):
        print(f"Excel input folder not found: {excel_dir}")
        sys.exit(1)

    excel_files = [f for f in os.listdir(excel_dir) if f.lower().endswith(('.xlsx', '.xlsb'))]
    if not excel_files:
        print(f"No .xlsx or .xlsb files found in {excel_dir}")
        sys.exit(1)

    # Create out directory if it doesn't exist
    if not os.path.exists("out"):
        os.makedirs("out")

    for excel_file in excel_files:
        excel_path = os.path.join(excel_dir, excel_file)
        term = extract_term_from_filename(excel_file)
        if not term:
            print(f"Could not extract term from filename: {excel_file}")
            continue
        output_csv = os.path.join("out", f"grades_{term}.csv")
        print(f"Converting {excel_path} to {output_csv} for term {term}")
        excel_to_csv(excel_path, output_csv)
        print(f"Conversion complete for {excel_file}.")


if __name__ == "__main__":
    main()
