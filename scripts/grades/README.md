# Grades Scraper

This script converts UTD grade distribution Excel files (`.xlsx` or `.xlsb` format) into CSV files for processing and upload to Firebase. The scraper automatically detects the semester from the filename and outputs standardized CSV files.

## Setup

Install the required Python dependencies:

```bash
pip install -r requirements.txt
```

The following packages are required:
- `openpyxl` - for `.xlsx` file handling
- `pyxlsb` - for `.xlsb` file handling

## Usage

1. **Add Excel Files**: Place your grade distribution Excel files in the `put-excel-here/` folder
   - Supported formats: `.xlsx`, `.xlsb`
   - Files should be named with the semester included (e.g., "Spring 2024 Grades.xlsb", "Fall 2024.xlsx", or "24f.xlsb")

2. **Run the Scraper**:
   ```bash
   python main.py
   ```

3. **Output**: CSV files will be generated in the `out/` directory with the naming convention `grades_{term}.csv` (e.g., `grades_24s.csv`, `grades_24f.csv`)

## How It Works

### 1. File Discovery (`main.py`)
- Scans the `put-excel-here/` folder for all `.xlsx` and `.xlsb` files
- Clears any existing files in the `out/` directory to ensure clean output

### 2. Term Detection (`extract_term_from_filename`)
The scraper uses two strategies to extract the semester term from filenames:

**Strategy A: Natural Language Parsing**
- Looks for patterns like "Spring 2024", "Fall 2025", "Summer 2024"
- Converts season names to term codes:
  - `Spring` → `s`
  - `Summer` → `u`
  - `Fall` → `f`
- Extracts the year and converts to 2-digit format (e.g., `2024` → `24`)
- Example: "Spring 2024 Grades.xlsb" → `24s`

**Strategy B: Term Code Detection (Fallback)**
- Searches for explicit term codes in the filename (e.g., `25s`, `24u`, `24f`)
- Example: "grades_24f.xlsx" → `24f`

If neither strategy succeeds, the file is skipped with a warning message.

### 3. Excel to CSV Conversion (`excel_to_csv.py`)

**Phase A: File Format Handling**
- Detects file extension (`.xlsx` or `.xlsb`)
- Uses appropriate library:
  - `openpyxl` for `.xlsx` files
  - `pyxlsb` for `.xlsb` files (binary Excel format)

**Phase B: Data Extraction**
- Reads from the "GradeDist" worksheet
- Fixes column header: Changes column B to "Catalog Nbr" (standardization)
- For data rows: Ensures empty cells in columns 22-26 are explicitly set to empty strings (prevents data inconsistency)

**Phase C: Data Cleaning**
- Trims trailing empty columns from all rows
- Converts float values to integers where appropriate (e.g., `2301.0` → `2301`)
- Pads short rows with empty strings to match header length
- Truncates rows that are longer than the header

**Phase D: CSV Writing**
- Removes any existing output file with the same name
- Writes cleaned data to CSV with minimal quoting
- Outputs to `out/grades_{term}.csv`

### 4. Output Generation
- Each Excel file produces one corresponding CSV file
- Files are named using the detected term code (e.g., `grades_25s.csv`)
- Progress is logged for each file conversion

## Output

CSV files are saved to the `out/` directory with the following naming convention:

```
out/grades_{term}.csv
```

Where `{term}` is the semester code (e.g., `24s` for Spring 2024, `24f` for Fall 2024, `25u` for Summer 2025).

### Output Format

The CSV contains grade distribution data with the following columns:
- Subject (e.g., "ACCT", "CS", "MATH")
- Catalog Nbr (course number, e.g., "2301", "1337")
- Section (e.g., "001", "002")
- Instructor 1, Instructor 2 (instructor names)
- Grade distribution columns (A+, A, A-, B+, B, etc.)
- Enrollment statistics
- Other metadata columns

Example row:
```csv
Subject,Catalog Nbr,Section,Instructor 1,A+,A,A-,B+,B,B-,...
ACCT,2301,001,John Smith,5,12,8,10,7,4,...
```

## Integration with Firebase

After conversion, the CSV files are uploaded to Firebase Cloud Storage when the scraper is invoked via the Go driver with `SAVE_ENVIRONMENT=dev` or `SAVE_ENVIRONMENT=prod`. The upload is handled by `internal/scraper/grades.go` which:
- Reads all CSV files from the `out/` directory
- Uploads each file to `grades/{filename}` in Firebase Storage
- Sets the content type to `text/csv`
- Cleans up local output files after successful upload

## Troubleshooting

**"Excel input folder not found"**
- Ensure the `put-excel-here/` directory exists in the `scripts/grades/` folder

**"No .xlsx or .xlsb files found"**
- Verify that Excel files are placed in `put-excel-here/`
- Check that files have the correct extension (`.xlsx` or `.xlsb`)

**"Could not extract term from filename"**
- Rename your file to include the semester (e.g., "Spring 2024.xlsx" or "24s.xlsx")
- Supported formats: "Spring/Summer/Fall YYYY" or "YYs/u/f"

**"pyxlsb is required for .xlsb support"**
- Install the missing dependency: `pip install pyxlsb`