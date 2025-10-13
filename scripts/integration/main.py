# integration scraper will read from /in to retrieve coursebook data, grades data, and rmp data
# then will pretty much follow original professor scraper logic:
# 1. run aggregator.py logic to aggregate grades data across every semester
# 2. run professor main.py logic to match rmp data to aggregated grades data (matched data will use normalized names as the key)
# 3. match coursebook data to original grades data to get instructor ids
# 4. for remaining grades data without coursebook sections, try to match through the instructor name on the professor data
# 5. copy instructor data to a new set with id as the key
# 6. return everything to /out files for go driver to upload to the desired save environment

import json
import csv
import os
import time
import dotenv

dotenv_path = dotenv.find_dotenv()
if dotenv_path:
   dotenv.load_dotenv(dotenv_path)
else:
   dotenv.load_dotenv()

from aggregator import calculate_professor_ratings_from_grades
from professor_matcher import match_professor_names
from mapper import map_grades_to_instructors, create_instructor_id_lookup


def load_input_data():
   """Load all input data from the /in directory."""
   print("Loading input data...")
    
   # load all coursebook data
   coursebook_data = []
   coursebook_dir = "in/coursebook"
   print(f"Loading coursebook data from {coursebook_dir}...")
   if os.path.exists(coursebook_dir):
      for filename in os.listdir(coursebook_dir):
         if filename.endswith(".json"):
            filepath = os.path.join(coursebook_dir, filename)
            # print(f"Loading coursebook file: {filename}")
            with open(filepath, "r", encoding="utf-8") as f:
               coursebook_data.extend(json.load(f))
    
   # load all grades data
   grades_files = []
   grades_dir = "in/grades"
   print(f"Loading grades data from {grades_dir}...")
   if os.path.exists(grades_dir):
      for filename in os.listdir(grades_dir):
         if filename.endswith(".csv"):
               # print(f"Loading grades file: {filename}")
               grades_files.append(os.path.join(grades_dir, filename))
   
   # load rmp data
   rmp_filepath = "in/rmp-profiles/rmp_data.json"
   rmp_data = {}
   print(f"Loading RMP data from {rmp_filepath}...")
   if os.path.exists(rmp_filepath):
      with open(rmp_filepath, "r", encoding="utf-8") as f:
         rmp_data = json.load(f)
    
   print(f"Loaded {len(coursebook_data)} coursebook sections")
   print(f"Found {len(grades_files)} grades files")
   print(f"Loaded {len(rmp_data)} RMP professor entries")
   
   return coursebook_data, grades_files, rmp_data

def available_grade_semesters(grades_dir=None):
   """Return set of semester ids available from filenames like grades_25s.csv under in/grades."""
   base = grades_dir or os.path.join(os.path.dirname(__file__), "in", "grades")
   semesters = set()
   if os.path.exists(base):
      for filename in os.listdir(base):
         if filename.startswith("grades_") and filename.endswith(".csv"):
            sem = filename.replace("grades_", "").replace(".csv", "")
            semesters.add(sem)
   return semesters


def resolve_enhance_terms():
   """Resolve which semesters to enhance for integration.

   Priority:
     1) ENHANCE_TERMS env var. If value=='all' (case-insensitive) -> enhance all available semesters.
     2) CLASS_TERMS env var -> enhance those terms.
     3) Neither set -> enhance all available semesters.

   Returns:
     - None -> indicates 'all' (enhance all available semesters)
     - set(...) -> explicitly requested semesters
   """
   val = os.environ.get("ENHANCE_TERMS")
   if val:
      val = val.strip()
      if val.lower() == "all":
         return None
      terms = {t.strip() for t in val.split(',') if t.strip()}
      return terms if terms else None

   # Fallback to CLASS_TERMS
   val = os.environ.get("CLASS_TERMS")
   if val:
      terms = {t.strip() for t in val.split(',') if t.strip()}
      return terms if terms else None

   # Default: all
   return None


# determine which semesters are available and which to enhance based on env vars
# supports "all" for ENHANCE_TERMS to indicate every available semester
# supports CLASS_TERMS as a fallback for most cases, where ENHANCE_TERMS is the override
def determine_target_semesters():
   """High-level wrapper to decide which grade semesters to enhance.

   Returns a set of semester ids to enhance or None to indicate 'all available'.
   This will print helpful messages if requested semesters are missing.
   """
   requested = resolve_enhance_terms()
   available = available_grade_semesters()

   if requested is None:
      print(f"Enhancing all available semesters: {sorted(available)}")
      return available

   requested_set = set(requested)
   matched = requested_set & available
   missing = requested_set - available
   if not matched:
      print(f"No grade files found for requested semesters: {sorted(requested_set)}")
      print(f"Available semesters: {sorted(available)}. Example valid input: '{sorted(available)[-1] if available else '25s'}'")
      return set()
   if missing:
      print(f"Requested terms not present and will be skipped: {sorted(missing)}")
   return matched


def save_output_data(matched_professor_data, enhanced_grades_by_file, instructor_by_id):
   """Save all output data to the /out directory."""
   print("Saving output files...")
   
   os.makedirs("out", exist_ok=True)
   os.makedirs("out/professors", exist_ok=True)
   os.makedirs("out/grades", exist_ok=True)

   # Save matched professor data (by name)
   with open("out/professors/matched_professor_data_names.json", "w", encoding="utf-8") as f:
      json.dump(matched_professor_data, f, indent=4, ensure_ascii=False)
   
   # Save instructor lookup (by ID)
   with open("out/professors/matched_professor_data.json", "w", encoding="utf-8") as f:
      json.dump(instructor_by_id, f, indent=4, ensure_ascii=False)
   
   # Save enhanced grades files individually
   total_grades = 0
   for filepath, enhanced_grades in enhanced_grades_by_file.items():
      if enhanced_grades:  # Only save if we have data
         # Extract semester from original filename (e.g., "grades_25s.csv" -> "25s")
         basename = os.path.basename(filepath)
         semester = basename.replace("grades_", "").replace(".csv", "")
         output_filename = f"out/grades/enhanced_grades_{semester}.csv"
         
         with open(output_filename, "w", encoding="utf-8", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=enhanced_grades[0].keys())
            writer.writeheader()
            writer.writerows(enhanced_grades)
         
         total_grades += len(enhanced_grades)
         print(f"  Saved enhanced_grades_{semester}.csv with {len(enhanced_grades)} entries")
   
   return len(matched_professor_data), len(instructor_by_id), total_grades


def generate_stats(coursebook_data, matched_professor_data, instructor_by_id, enhanced_grades_by_file):
   """Generate and save summary statistics."""
   all_enhanced_grades = []
   for enhanced_grades in enhanced_grades_by_file.values():
      all_enhanced_grades.extend(enhanced_grades)
   
   stats = {
      "total_coursebook_sections": len(coursebook_data),
      "total_grade_entries": len(all_enhanced_grades),
      "matched_professor_names": len(matched_professor_data),
      "matched_professor_ids": len(instructor_by_id),
      "grades_with_instructor_ids": len([g for g in all_enhanced_grades if g["instructor_id"]]),
      "grades_without_instructor_ids": len([g for g in all_enhanced_grades if not g["instructor_id"]])
   }
   
   return stats


def main():
   """Main integration function that processes all data and creates output files."""
   print("Starting integration scraper...")
   start_time = time.time()
   
   # 1. Load input data
   coursebook_data, grades_files, rmp_data = load_input_data()
   
   # 2. Calculate professor ratings from grades
   print("Calculating professor ratings from grades...")
   professor_ratings = calculate_professor_ratings_from_grades(grades_files, coursebook_data)
   
   # 3. Match RMP data with professor ratings
   print("Matching RMP data with professor ratings...")
   matched_professor_data = match_professor_names(professor_ratings, rmp_data)
   
   # 4. Map grades to instructor IDs
   print("Mapping grades to instructor IDs...")
   target_semesters = determine_target_semesters()
   if not target_semesters:
      # nothing to enhance (either no available semesters or requested ones missing)
      print("No target semesters to enhance. Exiting.")
      return

   enhanced_grades_by_file = map_grades_to_instructors(grades_files, coursebook_data, matched_professor_data, target_semesters)
   
   # 5. Create instructor ID lookup
   print("Creating instructor ID lookup...")
   instructor_by_id = create_instructor_id_lookup(matched_professor_data)
   
   # 6. Save output files
   matched_count, instructor_count, grades_count = save_output_data(
      matched_professor_data, enhanced_grades_by_file, instructor_by_id
   )

   print(f"Saved data for {matched_count} matched professors.")
   
   # 7. Generate and save statistics
   stats = generate_stats(coursebook_data, matched_professor_data, instructor_by_id, enhanced_grades_by_file)
   
   # 8. Print summary
   end_time = time.time()
   print(f"\nIntegration complete in {end_time - start_time:.2f} seconds!")
   print(f"Results saved to /out directory:")
   print(f"  - matched_professor_data_names.json: {stats['matched_professor_names']} professors by name")
   print(f"  - matched_professor_data.json: {stats['matched_professor_ids']} professors by ID")
   print(f"  - total grade distributions: {grades_count} grade entries")
   print(f" NOTE: professor data keyed by names vs ids has a length discrepancy due to cases of duplicate names with different ids")
   print(f"\nMapping results:")
   print(f"  - Grades with instructor IDs: {stats['grades_with_instructor_ids']}")
   print(f"  - Grades without instructor IDs: {stats['grades_without_instructor_ids']}")


if __name__ == "__main__":
   main()