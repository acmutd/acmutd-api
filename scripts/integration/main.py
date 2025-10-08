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

from aggregator import calculate_professor_ratings_from_grades
from professor_matcher import match_professor_names
from mapper import map_grades_to_instructors, create_instructor_id_lookup


def load_input_data():
   """Load all input data from the /in directory."""
   print("Loading input data...")
    
   # load all coursebook data
   coursebook_data = []
   coursebook_dir = "in/coursebook"
   if os.path.exists(coursebook_dir):
      for filename in os.listdir(coursebook_dir):
         if filename.endswith(".json"):
            filepath = os.path.join(coursebook_dir, filename)
            print(f"Loading coursebook file: {filename}")
            with open(filepath, "r", encoding="utf-8") as f:
               coursebook_data.extend(json.load(f))
    
   # load all grades data
   grades_files = []
   grades_dir = "in/grades"
   if os.path.exists(grades_dir):
      for filename in os.listdir(grades_dir):
         if filename.endswith(".csv"):
               print(f"Loading grades file: {filename}")
               grades_files.append(os.path.join(grades_dir, filename))
   
   # load rmp data
   rmp_filepath = "in/rmp-profiles/rmp_data.json"
   rmp_data = {}
   if os.path.exists(rmp_filepath):
      with open(rmp_filepath, "r", encoding="utf-8") as f:
         rmp_data = json.load(f)
    
   print(f"Loaded {len(coursebook_data)} coursebook sections")
   print(f"Found {len(grades_files)} grades files")
   print(f"Loaded {len(rmp_data)} RMP professor entries")
   
   return coursebook_data, grades_files, rmp_data


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
   with open("out/professors/matched_professor_data_ids.json", "w", encoding="utf-8") as f:
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
   # Flatten all enhanced grades into a single list for stats
   all_enhanced_grades = []
   for enhanced_grades in enhanced_grades_by_file.values():
      all_enhanced_grades.extend(enhanced_grades)
   
   stats = {
      "total_coursebook_sections": len(coursebook_data),
      "total_grade_entries": len(all_enhanced_grades),
      "matched_professors": len(matched_professor_data),
      "instructors_by_id": len(instructor_by_id),
      "grades_with_instructor_ids": len([g for g in all_enhanced_grades if g["instructor_id"]]),
      "grades_with_rmp_data": len([g for g in all_enhanced_grades if g["has_rmp_data"]])
   }
   
   with open("out/integration_stats.json", "w", encoding="utf-8") as f:
      json.dump(stats, f, indent=4, ensure_ascii=False)
   
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
   enhanced_grades_by_file = map_grades_to_instructors(grades_files, coursebook_data, matched_professor_data)
   
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
   print(f"  - matched_professor_data.json: {matched_count} professors")
   print(f"  - instructor_by_id.json: {instructor_count} instructors by ID")
   print(f"  - enhanced_grades.csv: {grades_count} grade entries")
   print(f"  - integration_stats.json: Summary statistics")
   print(f"\nMapping results:")
   print(f"  - Grades with instructor IDs: {stats['grades_with_instructor_ids']}")
   print(f"  - Grades with RMP data: {stats['grades_with_rmp_data']}")


if __name__ == "__main__":
   main()