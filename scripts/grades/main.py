from excel_to_csv import excel_to_csv
import os
import sys
import dotenv

dotenv.load_dotenv()

def get_term_name(term):
	# Converts TERM like '24s' to 'Spring 2024', etc.
	if len(term) < 2:
		return term
	year = '20' + term[:2]
	season_code = term[2].lower()
	season = {'s': 'Spring', 'u': 'Summer', 'f': 'Fall'}.get(season_code, season_code)
	return f"{season} {year}"

def main():
	# Check for environment variables
	if 'TERM' not in os.environ:
		print("TERM environmental variable not set.")
		exit(1)

	term = os.environ['TERM']
	term_name = get_term_name(term)

	excel_dir = os.path.join(os.path.dirname(__file__), 'put-excel-here')
	if not os.path.isdir(excel_dir):
		print(f"Excel input folder not found: {excel_dir}")
		sys.exit(1)

	# Find first .xlsx or .xlsb file in the folder
	excel_files = [f for f in os.listdir(excel_dir) if f.lower().endswith(('.xlsx', '.xlsb'))]
	if not excel_files:
		print(f"No .xlsx or .xlsb files found in {excel_dir}")
		sys.exit(1)
	excel_path = os.path.join(excel_dir, excel_files[0])

	# Output CSV path in old format: "Fall 2024.csv"
	output_csv = os.path.join(os.path.dirname(__file__), f"{term_name}.csv")

	print(f"Converting {excel_path} to {output_csv} for term {term_name}")
	excel_to_csv(excel_path, output_csv)
	print("Conversion complete.")

if __name__ == "__main__":
	main()
