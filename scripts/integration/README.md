# Integration Scraper

The integration scraper combines data from the coursebook, grades, and RateMyProfessors (RMP) scrapers to create unified, enriched datasets. It maps instructor IDs, matches professor names across different data sources using fuzzy logic, validates matches through course overlap, and produces comprehensive professor profiles with both grade-based performance metrics and student ratings.

## Architecture

The integration scraper is organized into modular components that each handle a specific aspect of data integration:

### Core Modules

- **`main.py`** - Orchestrates the entire integration pipeline from data loading through output generation
- **`aggregator.py`** - Aggregates grade distributions across semesters and calculates professor ratings from grade data
- **`professor_matcher.py`** - Performs intelligent name matching between RMP and grade data using fuzzy logic and course overlap validation
- **`mapper.py`** - Maps grade CSV rows to coursebook sections using section addresses and extracts instructor IDs
- **`utils.py`** - Shared utility functions for name normalization, variation generation, and course overlap detection

### Data Flow

The integration process follows a sequential pipeline:

1. **Data Loading** (`load_input_data`)
   - Loads coursebook JSON files from `in/coursebook/`
   - Loads grade CSV files from `in/grades/`
   - Loads RMP data from `in/rmp-profiles/rmp_data.json`

2. **Grades Aggregation** (`calculate_professor_ratings_from_grades`)
   - Processes coursebook sections to build instructor name → ID mappings
   - Aggregates grade distributions across all semesters for each instructor
   - Calculates overall and per-course ratings based on grade point values
   - Groups professors by normalized names to handle name variations

3. **RMP Matching** (`match_professor_names`)
   - Applies manual name overrides from `manual_matches.json`
   - Performs direct name matches (normalized)
   - Uses fuzzy matching with name variations for remaining entries
   - Validates matches using course overlap to prevent false positives
   - Merges RMP data with grade-based ratings

4. **Instructor ID Mapping** (`map_grades_to_instructors`)
   - Creates section address lookup (e.g., `acct2301.001.25s`)
   - Maps each grade CSV row to coursebook section via section address
   - Extracts instructor IDs for matched sections
   - Falls back to name-based matching when section lookup fails
   - Produces enhanced grade CSVs with instructor IDs and normalized names

5. **Output Generation** (`save_output_data`, `generate_stats`)
   - Creates professor data indexed by name (`matched_professor_data_names.json`)
   - Creates professor data indexed by ID (`matched_professor_data.json`)
   - Writes per-semester enhanced grade CSVs with instructor metadata
   - Generates integration statistics (match rates, coverage metrics)

## How It Works

### 1. Environment Configuration (`determine_target_semesters`)

**ENHANCE_TERMS Priority System**
- Checks `ENHANCE_TERMS` environment variable first
- If set to `"all"` (case-insensitive), processes all available grade files
- If set to comma-separated terms (e.g., `"25s,25u"`), processes only those semesters
- Falls back to `CLASS_TERMS` if `ENHANCE_TERMS` is not set
- Defaults to processing all semesters if neither variable is set

**Validation and Guidance**
- Compares requested semesters against available grade files in `in/grades/`
- Prints warnings for requested semesters that don't have grade files
- Suggests valid semester identifiers when no matches are found
- Prevents processing when no valid target semesters exist

### 2. Coursebook Section Processing (`process_section_data`)

**Instructor-to-Course Mapping**
- Parses `instructors` and `instructor_ids` fields from coursebook sections
- Extracts the first instructor from comma-separated lists
- Normalizes instructor names for consistent matching
- Groups sections by instructor, tracking which courses each teaches
- Handles instructors who teach multiple courses or have multiple IDs

**Duplicate ID Detection**
- Identifies cases where the same normalized name maps to multiple instructor IDs
- Prints warnings for ambiguous mappings (e.g., "John Smith" with IDs "abc123" and "def456")
- Preserves all ID variations to avoid data loss

### 3. Grades Aggregation (`calculate_professor_ratings_from_grades`)

**Grade Point Conversion**
- Maps letter grades to 4.0 scale: A+/A=4.0, A-=3.67, B+=3.33, ..., F=0.0
- Special handling for W (withdrawal)=0.67, P (pass)=4.0, NP (no pass)=0.0
- Handles missing or malformed grade values gracefully

**Rating Calculation**
- Aggregates grade distributions across all semesters per instructor per course
- Calculates weighted average: `(Σ grade_value × count) / total_students`
- Converts 4.0-scale to 5.0-scale to match RMP: `(GPA / 4.0) × 5.0`
- Generates both overall ratings and per-course ratings

**Data Structure**
```json
{
    "john smith": [
        {
            "instructor_id": "abc123",
            "overall_grade_rating": 4.35,
            "total_grade_count": 1247,
            "course_ratings": {
                "CS1337": 4.50,
                "CS2336": 4.20
            }
        }
    ]
}
```

### 4. Professor Matching (`match_professor_names`)

**Phase A: Manual Overrides** (`apply_manual_matches`)
- Loads `manual_matches.json` if present
- Applies explicit name mappings before other matching
- Example: Maps "Yu Chung Ng" (grades) → "Vincent Ng" (RMP)
- Normalizes both sides before matching to handle spacing/formatting differences
- Removes successfully matched entries from further processing

**Phase B: Direct Matching**
- Compares normalized names between grade data and RMP data
- Exact matches are processed immediately
- When multiple profiles exist for the same name, selects best match based on:
  1. Course overlap (do they teach the same subjects?)
  2. Ratings count (higher = more reliable RMP data)

**Phase C: Fuzzy Matching with Name Variations** (`generate_name_variations`)
- For unmatched entries, generates name variations:
  - Swap first/last: "John Smith" ↔ "Smith John"
  - Skip middle names: "John Paul Smith" → "John Smith"
  - First + last only: "Carlos Busso Recabarren" → "Carlos Recabarren"
  - Remove first/last name variations
  - Handles 3+ part names with multiple combinations
- Compares all variations using fuzzy string matching (80% threshold)
- Example: "Bhadrachalam Chitturi" matches "Chitturi Bhadrachalam" at 85% similarity

**Phase D: Course Overlap Validation** (`check_course_overlap`)
Three validation levels (OR logic - any match is sufficient):
1. **Exact course code match**: "CS1337" in both RMP and grades
2. **Department match**: Both teach "CS" courses (extracts prefix from course codes)
3. **Course number match**: Both teach "1337" (extracts numbers from course codes)

This prevents matching professors with similar names who teach different subjects (e.g., "John Smith" in CS vs. "John Smith" in MATH).

**Unmatched Data Handling**
- Professors in grades without RMP matches are still included in output
- RMP entries without grade matches are logged but not included in final data
- Statistics report match rates and unmatched counts for debugging

### 5. Grade-to-Instructor Mapping (`map_grades_to_instructors`)

**Section Address Lookup** (`find_instructor_id_by_section_address`)
- Builds a lookup dictionary keyed by `section_address` from coursebook
- Section address format: `{prefix}{number}.{section}.{term}` (e.g., `acct2301.001.25s`)
- For each grade row, constructs the section address from Subject + Catalog Nbr + Section
- Performs lowercase prefix matching to find the coursebook section
- Extracts the first instructor ID from the section's `instructor_ids` field

**Fallback Name Matching**
- When section address lookup fails (section not found in coursebook):
  - Extracts "Instructor 1" from the grade CSV row
  - Normalizes the instructor name
  - Searches matched professor data for the normalized name
  - If found, uses the first instructor ID from the match
- This catches cases where coursebook data is incomplete or section addresses don't align

**Enhanced Row Generation**
Each grade CSV row is enhanced with:
- `instructor_id`: UTD instructor ID (empty string if not found)
- `instructor_name_normalized`: Lowercased, standardized name for matching

**Per-File Processing**
- Respects `target_semesters` filter (skips non-target files)
- Returns a dictionary: `{filepath: [enhanced_rows]}`
- Preserves all original CSV columns plus new fields

**Statistics Tracking**
- Section address matches: Found via coursebook section lookup
- Fallback name matches: Found via instructor name matching
- No matches: Neither method succeeded
- Prints match percentages and totals

**Zero-Division Protection**
- If no grades are processed (e.g., selected an ongoing semester with no data yet):
  - Prints helpful guidance with possible causes
  - Lists available semesters from `in/grades/`
  - Suggests a valid example semester to try
  - Returns early to avoid division-by-zero errors

### 6. Instructor ID Lookup Creation (`create_instructor_id_lookup`)

**Purpose**
- Converts name-keyed professor data to ID-keyed for faster lookups
- Enables direct access to professor profiles via instructor ID

**Transformation**
- Input: `{"john smith": [{"instructor_id": "abc123", ...}]}`
- Output: `{"abc123": {"normalized_coursebook_name": "john smith", ...}}`

**Key Changes**
- Removes `instructor_id` from value (now the dict key)
- Adds `normalized_coursebook_name` to preserve the original name mapping
- Handles multiple profiles per name by creating separate ID entries

### 7. Output Generation and Statistics

**File Outputs**
1. `out/professors/matched_professor_data_names.json`
   - Professors keyed by normalized name
   - Contains grade ratings + RMP data merged
   - Supports name-based lookups

2. `out/professors/matched_professor_data.json`
   - Professors keyed by instructor ID
   - Includes `normalized_coursebook_name` for reverse lookups
   - Supports ID-based lookups (faster, more reliable)

3. `out/grades/enhanced_grades_{semester}.csv`
   - One file per semester (e.g., `enhanced_grades_25s.csv`)
   - All original CSV columns preserved
   - Additional fields: `instructor_id`, `instructor_name_normalized`

**Statistics Generated**
- Total coursebook sections loaded
- Total grade entries processed
- Matched professors (by name and by ID counts)
- Grades with instructor IDs vs. without
- Match rates and coverage percentages

**Execution Summary**
- Total execution time
- File counts and paths
- Warnings about name vs. ID count discrepancies
- Mapping success rates

### 8. Semester Sorting and Academic Ordering

**Custom Sort Logic**
- Academic year ordering: Spring (s) → Summer (u) → Fall (f)
- Not alphabetical: Prevents "24f, 25s, 25u" from sorting incorrectly
- Two-digit year extraction: "25s" → year=25, term=s (rank 0)
- Used in semester suggestions and error messages

**Example Ordering**
```
24s (Spring 2024)
24u (Summer 2024)
24f (Fall 2024)
25s (Spring 2025)
25u (Summer 2025)
```

## Input Data

### Expected Directory Structure
```
in/
├── coursebook/
│   └── classes_25f.json       # Coursebook section data
├── grades/
│   ├── grades_24u.csv         # Grade distribution files
│   ├── grades_25s.csv
│   └── grades_25u.csv
└── rmp-profiles/
    └── rmp_data.json          # RateMyProfessors data
```

### Data Formats

- **Coursebook Data**: JSON array with section objects containing `instructor_ids`, `instructors`, `course_prefix`, `course_number`, `section_number`
- **Grades Data**: CSV files with columns: `Subject`, `Catalog Nbr`, `Section`, `Instructor 1`, grade distributions (`A+`, `A`, `B+`, etc.)
- **RMP Data**: JSON object with professor names as keys, arrays of professor profile objects as values
- **Manual Matches** (optional): `manual_matches.json` file with explicit name mappings for edge cases

### Manual Matching

For cases where professors have different names in grades vs RMP data, create a `manual_matches.json` file:

```json
[
    {
        "ratings_name": "Yu Chung Ng",
        "rmp_name": "Vincent Ng"
    }
]
```

This ensures professors with known name discrepancies are still properly matched.

## Output Data

### Generated Files
```
out/
├── professors/  
│   ├── matched_professor_data_names.json   # Professors with RMP data (indexed by name)  
│   └── matched_professor_data.json         # Professors with RMP data (indexed by ID)  
└── grades/  
    └── enhanced_grades_<semester>.csv      # Enhanced grades per semester (with instructor IDs and RMP flags)  
```

### Enhanced Grades CSVs

Each enhanced grades CSV (one per semester) includes all original grade data plus:
- `instructor_id` - UTD instructor ID from coursebook
- `instructor_name_normalized` - Normalized instructor name

## Matching Logic

### Name Normalization (`normalize_name`)

**Transformations Applied**:
1. **Comma handling**: Standardizes spacing around commas
2. **Middle initial removal**: Strips trailing initials like "A.", "B.C."
3. **Initial spacing**: Adds space between consecutive initials ("A.B" → "A B")
4. **Period removal**: Removes all periods from names
5. **Apostrophe removal**: Strips apostrophes, single quotes, backticks
6. **Hyphen conversion**: Replaces hyphens with spaces
7. **Last, First swap**: Converts "Smith, John" to "john smith"
8. **Lowercase conversion**: Ensures case-insensitive matching
9. **Space normalization**: Collapses multiple spaces to single space

**Examples**:
- "O'Brien, Patrick J." → "patrick j obrien"
- "Smith-Jones, Mary Ann" → "mary ann smith jones"
- "Ng, Yu-Chung" → "yu chung ng"

### Fuzzy Matching Process

**1. Manual Matches** (`manual_matches.json`)
- Highest priority - applied before any automated matching
- Explicit mappings for known edge cases
- Example use case: Professor goes by different names in different systems
```json
[
    {
        "ratings_name": "Yu Chung Ng",
        "rmp_name": "Vincent Ng"
    }
]
```

**2. Direct Matches**
- Exact normalized name matches
- No fuzzy logic needed - 100% confidence
- Fastest and most reliable matching method

**3. Name Variations** (`generate_name_variations`)
Handles complex name formats:
- 2-part names: "John Smith"
  - Swap: "Smith John"
- 3-part names: "John Paul Smith"
  - First + last: "John Smith"
  - First + middle: "John Paul"
  - Swap first/last: "Smith John"
  - Skip first: "Paul Smith"
  - Skip last: "John Paul"
- 4+ part names: Multiple combinations tested
  - Useful for hyphenated or multi-cultural names

**4. Fuzzy Threshold**
- Uses FuzzyWuzzy's ratio algorithm (Levenshtein distance-based)
- Threshold: 80% similarity required for match
- Prevents false positives from completely different names

**5. Course Overlap Validation** (`check_course_overlap`)

Three levels of validation (any one match is sufficient):

**Level 1: Exact Course Match**
- RMP courses: `["CS1337", "CS2336"]`
- Grades courses: `["CS1337", "CS3345"]`
- Match: `CS1337` appears in both → Valid ✓

**Level 2: Department Match**
- Extracts course prefixes (letters before numbers)
- RMP: `["CS1337", "CS2336"]` → departments: `{CS}`
- Grades: `["MATH2413"]` → departments: `{MATH}`
- Match: No common department → Invalid ✗

**Level 3: Course Number Match**
- Extracts numeric portions of course codes
- RMP: `["CS1337", "EE1337"]` → numbers: `{1337}`
- Grades: `["MATH1337"]` → numbers: `{1337}`
- Match: `1337` appears in both → Valid ✓

**Why Course Overlap Matters**:
- Prevents matching professors with common names who teach different subjects
- Example: "John Smith" in Computer Science ≠ "John Smith" in Mathematics
- Increases match confidence by validating domain alignment

### Multiple Profiles Per Name

**When It Occurs**:
- Same normalized name maps to different instructor IDs
- Common with very common names (e.g., "John Smith", "Michael Johnson")
- Different people, or same person with multiple IDs in the system

**Handling Strategy**:
- Preserve all profiles as array entries
- Use course overlap to disambiguate when matching
- Select profile with highest RMP ratings count if multiple course-matched profiles exist
- Output both name-keyed and ID-keyed data for flexibility

**Example**:
```json
{
    "john smith": [
        {
            "instructor_id": "abc123",
            "department": "Computer Science",
            "overall_grade_rating": 4.5
        },
        {
            "instructor_id": "def456",
            "department": "Mathematics",
            "overall_grade_rating": 3.8
        }
    ]
}
```

## Usage

### Prerequisites
```bash
pip install -r requirements.txt
```

Required dependencies:
- `fuzzywuzzy` - Fuzzy string matching for name variations
- `python-dotenv` - Environment variable management

### Environment Variables

Create a `.env` file in the project root or set environment variables:

```env
# Primary control (takes precedence over CLASS_TERMS)
ENHANCE_TERMS=all              # Process all available semesters
# OR
ENHANCE_TERMS=25s,25u,25f      # Process specific semesters only

# Fallback control (used if ENHANCE_TERMS not set)
CLASS_TERMS=25s,25u,25f        # Legacy semester selection
```

**Behavior**:
1. If `ENHANCE_TERMS` is set:
   - `"all"` → processes all grade files found in `in/grades/`
   - Comma-separated list → processes only those semesters
2. If `ENHANCE_TERMS` not set, falls back to `CLASS_TERMS`
3. If neither set, defaults to processing all semesters

### Running the Integration
```bash
cd scripts/integration
python main.py
```

### Expected Output
```
Loading input data...
Loaded 73017 coursebook sections
Found 24 grades files
Loaded 1842 RMP professor entries

Calculating professor ratings from grades...
Matching RMP data with professor ratings...
Applying 5 manual matches...
Direct Matches: 847
Matched Professors: 1653

Mapping grades to instructor IDs...
Enhancing all available semesters: ['24s', '24u', '24f', '25s']
Mapping grades in grades_24s.csv...
...

--- Instructor Mapping Statistics ---
Total grades processed: 49340
Section address matches: 28350 (57.5%)
Fallback name matches: 19087 (38.7%)
No matches found: 1903 (3.8%)
Total matched: 47437 (96.2%)

Integration complete in 12.34 seconds!
Results saved to /out directory
```

## Statistics

The integration process tracks comprehensive metrics:

**Data Volume**:
- Total coursebook sections loaded
- Total grade entries processed across all semesters
- RMP professor entries loaded

**Matching Performance**:
- Direct name matches (exact normalized matches)
- Fuzzy matches (via name variations)
- Manual matches applied
- Unmatched ratings (professors in grades but not RMP)
- Unmatched RMP (profiles in RMP but not grades)

**Instructor Mapping**:
- Section address matches (via coursebook section lookup)
- Fallback name matches (when section address fails)
- No matches (neither method succeeded)
- Total match rate percentage

**Output Counts**:
- Matched professors by name (can be > matched by ID due to duplicate names)
- Matched professors by ID (unique instructor count)
- Grades with instructor IDs
- Grades without instructor IDs

**Why Name Count ≠ ID Count**:
- Same normalized name can map to multiple instructor IDs
- Example: "Daniel Griffith" has 2 IDs in the system
- Name-keyed: 1 entry (array with 2 profiles)
- ID-keyed: 2 entries (separate instructor records)

Example statistics output:
```
Total coursebook sections: 73017
Grade entries: 49340
Matched professor names: 1653
Matched professor IDs: 1689
Grades with instructor IDs: 47437 (96.2%)
Grades without instructor IDs: 1903 (3.8%)

Section address matches: 28350 (57.5%)
Fallback name matches: 19087 (38.7%)
No matches: 1903 (3.8%)
```

## Error Handling

**Missing Input Files**:
- Gracefully handles missing directories (`in/coursebook/`, `in/grades/`, etc.)
- Prints informative messages about which data sources are missing
- Continues processing with available data rather than failing completely

**Invalid or Malformed Data**:
- Skips grade CSV rows with missing required fields (Subject, Catalog Nbr, Instructor)
- Ignores rows with zero total grades (empty grade distributions)
- Handles both quoted and unquoted CSV column names ("Catalog Nbr" vs Catalog Nbr)

**Encoding Issues**:
- Uses `UTF-8-sig` encoding for CSV files to handle BOM (Byte Order Mark)
- Strips extra whitespace from parsed values
- Handles various name formats and special characters

**Empty or Missing Semesters**:
- Detects when requested semesters have no grade files
- Prints available semester list with academic ordering
- Suggests valid example semesters to try
- Prevents division-by-zero when no grades are processed

**Name Matching Failures**:
- Professors without RMP matches are still included in output (with grade data only)
- Unmatched RMP entries are logged for review
- Statistics clearly show match rates for debugging

**Duplicate Detection**:
- Identifies and warns about professors with multiple instructor IDs
- Preserves all profiles to avoid data loss
- Logs specific examples for manual review

## Integration with Firebase

After integration, the output files are uploaded to Firebase Cloud Storage when invoked via the Go driver with `SAVE_ENVIRONMENT=dev` or `SAVE_ENVIRONMENT=prod`. The upload is handled by `internal/scraper/integration.go` (if implemented) which would:
- Read JSON and CSV files from `out/professors/` and `out/grades/`
- Upload to appropriate paths in Firebase Storage
- Set correct content types (`application/json` for JSON, `text/csv` for CSV)
- Clean up local output files after successful upload

The integrated data is then available for:
- API endpoints serving professor profiles
- Course search features with instructor ratings
- Student tools for course planning and professor selection
- Analytics and reporting dashboards

## Troubleshooting

**"No grade files found for requested semesters"**
- Check `ENHANCE_TERMS` or `CLASS_TERMS` environment variable for typos
- Verify grade files exist in `in/grades/` with naming format `grades_25s.csv`
- Review available semesters list printed in error message

**"Could not extract term from filename"** (from grades scraper)
- Ensure Excel files in `scripts/grades/put-excel-here/` have recognizable semester names
- Supported formats: "Spring 2025", "Fall 2024", "25s", "24f", etc.
- Run grades scraper to generate missing CSV files

**Low match rates (< 90%)**
- Check if coursebook data is complete and recent
- Verify section_address format in coursebook JSON
- Review unmatched professors in statistics output
- Consider adding manual matches for known edge cases

**Name vs ID count discrepancy**
- This is expected when professors share normalized names
- Check terminal output for "Instructor name 'X' has multiple associated IDs"
- Review both name-keyed and ID-keyed output files for correctness

**"Manual match failed"**
- Verify both names exist in grades and RMP data
- Check for typos in `manual_matches.json`
- Ensure names are in original format (not normalized)

**Missing RMP data in enhanced grades**
- Verify RMP scraper ran successfully and `in/rmp-profiles/rmp_data.json` exists
- Check RMP match statistics for low match rates
- Instructor may not have RMP profile (new professors, adjuncts)

**Fuzzy matching too aggressive/conservative**
- Adjust `fuzzy_threshold` parameter in `match_professor_names()` (default: 80)
- Higher threshold = more conservative (fewer false positives)
- Lower threshold = more aggressive (more matches, but possible false positives)

## Development

### Adding New Data Sources
1. Extend `load_input_data()` in `main.py`
2. Add processing logic to appropriate module
3. Update output generation if needed

### Modifying Matching Logic
- Edit `matcher.py` for fuzzy matching improvements
- Modify `utils.py` for name normalization changes
- Update `aggregator.py` for grades calculation adjustments

### Testing
- Ensure input data is properly formatted
- Check output files for expected structure
- Validate statistics against expected ranges