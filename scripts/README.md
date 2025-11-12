# Wikipedia Import Script

This script imports January 6 Rioters data from Wikipedia into your database.

## Prerequisites

Install required Python packages:

```bash
pip install psycopg2-binary beautifulsoup4 requests
```

Or if using a virtual environment:

```bash
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install psycopg2-binary beautifulsoup4 requests
```

## Usage

### Option 1: Import into Test Table (Recommended First)

This creates a separate `rioters_import` table so you can test without affecting your main data:

```bash
cd scripts

# Create test table and import
python3 import_wikipedia.py \
  --create-test-table \
  --test-table \
  --dry-run  # Remove this for actual import
```

### Option 2: Import into Main Table (Updates Existing)

This will update existing records and add new ones:

```bash
cd scripts

# Set environment variables
export DATABASE_URL="postgresql://user:password@localhost:5432/rioters?sslmode=disable"
export MAPBOX_ACCESS_TOKEN="your_mapbox_token_here"

# Dry run first
python3 import_wikipedia.py --dry-run

# Actual import
python3 import_wikipedia.py
```

### Option 3: Import Only New Records (No Updates)

This will skip existing records and only add new ones:

```bash
python3 import_wikipedia.py --no-update
```

## Command Line Options

- `--dry-run` - Preview what would be imported without making changes
- `--test-table` - Import into `rioters_import` table instead of `rioters`
- `--create-test-table` - Create a test table before importing
- `--drop-table` - Drop and recreate test table (use with `--create-test-table`)
- `--no-update` - Skip updating existing records (only insert new ones)
- `--table-name NAME` - Specify custom table name (default: `rioters`)
- `--db-url URL` - Database URL (or use DATABASE_URL env var)
- `--mapbox-token TOKEN` - Mapbox token (or use MAPBOX_ACCESS_TOKEN env var)

## What It Does

1. **Scrapes 4 Wikipedia pages:**
   - A-F: https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(A-F)
   - G-L: https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(G-L)
   - M-S: https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(M-S)
   - T-Z: https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(T-Z)

2. **Extracts data:**
   - Name (first, last, middle)
   - Arrest date
   - Charges
   - Case status (from pleas/judgment)
   - Location (city, state) from notes
   - Pardon/commutation status
   - Charge type flags (violence, conspiracy, etc.)

3. **Geocodes locations** using Mapbox API (if token provided)

4. **Imports/updates** records in the database:
   - Creates new records for rioters not in database
   - Updates existing records if name matches

## Notes

- The script will skip geocoding if Mapbox token is not provided
- It uses rate limiting to be respectful to Wikipedia and Mapbox APIs
- Existing records are matched by first name + last name (case-insensitive)
- The script handles Wikipedia's special formatting and links

