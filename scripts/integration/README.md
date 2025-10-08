# Integration Scraper

The integration scraper processes coursebook data, grades data, and RMP (RateMyProfessors) data to create a unified dataset with instructor mappings and professor ratings.

## Architecture

The integration scraper is organized into modular components:

### Core Modules

- **`main.py`** - Main orchestration script that coordinates all components
- **`aggregator.py`** - Handles grades aggregation logic from multiple semesters
- **`matcher.py`** - Performs fuzzy matching between RMP and professor ratings data
- **`mapper.py`** - Maps grades data to coursebook sections and extracts instructor IDs
- **`utils.py`** - Shared utility functions for name normalization and data processing

### Data Flow

1. **Data Loading** - Loads coursebook JSON, grades CSV files, and RMP JSON data from `/in` directory
2. **Grades Aggregation** - Processes grades across semesters and calculates professor ratings
3. **RMP Matching** - Uses fuzzy matching with name variations and course overlap validation
4. **Instructor Mapping** - Maps grade entries to instructor IDs using coursebook data
5. **Output Generation** - Creates enhanced datasets and statistics in `/out` directory

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

This ensures professors with name discrepancies are still properly matched.

## Output Data

### Generated Files
```
out/
├── matched_professor_data.json    # Professors with RMP data (indexed by name)
├── instructor_by_id.json         # Instructors indexed by ID
├── enhanced_grades.csv           # Grades with instructor IDs and RMP flags
└── integration_stats.json       # Summary statistics
```

### Enhanced Grades CSV

The enhanced grades CSV includes all original grade data plus:
- `instructor_id` - UTD instructor ID from coursebook
- `instructor_name_normalized` - Normalized instructor name
- `has_rmp_data` - Boolean flag indicating RMP data availability

## Matching Logic

### Name Normalization
- Removes periods, extra spaces, and apostrophes
- Handles "Last, First" format conversion
- Converts to lowercase for consistent matching

### Fuzzy Matching Process
1. **Manual Matches** - Applies predefined name mappings from `manual_matches.json`
2. **Direct Matches** - Exact normalized name matches
3. **Name Variations** - Generates first/last name combinations
4. **Course Overlap Validation** - Ensures matched professors teach similar courses
5. **Fuzzy Threshold** - Uses 80% similarity threshold for name matching

## Usage

### Prerequisites
```bash
pip install -r requirements.txt
```

### Running the Integration
```bash
python main.py
```

### Dependencies
- `fuzzywuzzy` - Fuzzy string matching

## Statistics

The integration process tracks:
- Total coursebook sections processed
- Grade entries with instructor IDs mapped
- Professors successfully matched with RMP data
- Direct vs fuzzy match counts
- Unmatched entries for debugging

## Error Handling

- Missing files are handled gracefully with informative messages
- Invalid data entries are skipped with warnings
- Encoding issues are handled with UTF-8-sig for CSV files
- Empty or malformed data structures are validated

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