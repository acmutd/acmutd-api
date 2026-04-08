# Uploader

Reads scraped data from local files and writes it to Firestore. Controlled by the `UPLOADER` env var either `course` or `profs`.

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `UPLOADER` | Yes | `course` or `profs` |
| `SAVE_ENVIRONMENT` | Yes | `dev` or `prod` - selects which Firebase project |
| `CLASS_TERMS` | Course only | Comma-separated terms to upload (e.g. `24f,25s,26u`) |

Run with `--gather` to download input files from Cloud Storage before uploading.

## Input Files

All input is read from the `uploader/` directory at the repo root:

```
uploader/
├── coursebook/
│   └── classes_{term}.json           # one file per term
├── enhanced_grades/
│   └── enhanced_grades_{term}.csv    # one file per term (optional)
└── professors/
    └── matched_professor_data_ids.json
```

Grade files are optional — if missing for a term, courses are still uploaded without grade data.

## Files

### 1. Initialization (`uploader.go`)

Entry point and configuration. `NewUploaderHandler` validates env vars and builds the config. Optionally downloads input files from Cloud Storage (`--gather`), then routes to course or professor insertion based on `UPLOADER` env variable.

---

### 2. Data Aggregation (`aggregate.go`)

**Courses Aggregation**
- Loops through each term in CLASS_TERMS
    - Reads `classes_{term}.json` and `enhanced_grades_{term}.csv`.
    - Merges them by `section_address` and returns each course with all its sections.
    - Grade counts are aggregated from section level up to course level.
- **Section → General Course** relationship:
    - `loadCoursebookTerm` parses `RawCourse` entries from the scraper JSON into `SectionDoc` objects and builds a `CourseGeneralInfo` index keyed by `prefix/number`. After `mergeGrades` attaches grade data to sections
    - `loadAndCombine` iterates all sections and appends each to its parent `course.Sections`. 
    - Sections with no matching course log a warning and are skipped.

**Professors Aggregation**
- Reads `matched_professor_data_ids.json` and returns all professors with `InstructorID` as the key

---

### 3. Firestore Insertion (`insert.go`)

All writes for a term share a single `BulkWriter` (created in `insertTermData`, passed down) so the SDK can pipeline and batch them efficiently. `defer writer.End()` flushes everything when `insertTermData` returns.

- **Term Data Insertion**: top-level orchestrator. Creates the writer, calls insertion functions
    - **Course Insertion**: writes `CourseGeneralInfo` documents and their `SectionDoc` subcollections. For older terms, only `grades.{term}` is merged to avoid overwriting newer metadata.
    - **Term Insertion**: writes the `terms/{term}` document and a `prefixes` subcollection entry for each course prefix seen in that term.
- **Professor Insertion**: writes each professor as a full overwrite.

**Firestore structure written:**
```
courses/{prefix}/numbers/{number}                  ← CourseGeneralInfo
courses/{prefix}/numbers/{number}/sections/{addr}  ← SectionDoc
terms/{term}                                        ← term metadata
terms/{term}/prefixes/{prefix}                      ← prefix index
professors/{instructor_id}                          ← Professor
```
