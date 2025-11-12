# Add Charges Links Script

This script searches for court case links for each rioter and updates the `charges_link` field in the database.

## How It Works

1. Finds all rioters without `charges_link`
2. Searches for: `"[Full Name]" "Jan 6" "riot" court case charges`
3. Prioritizes links from known court sources:
   - justice.gov (DOJ)
   - courtlistener.com
   - dcd.uscourts.gov (DC District Court)
   - cases.justia.com
   - lawandcrime.com
4. Updates the `charges_link` field with the best match

## Usage

### Option 1: Using DuckDuckGo (No API Key Required)

```bash
cd scripts
export DATABASE_URL="postgresql://alexbeattie:xela@localhost:5432/rioters?sslmode=disable"

# Dry run first
python3 add_charges_links.py --dry-run

# Actual update
python3 add_charges_links.py
```

**Note:** DuckDuckGo has rate limits, so this will be slower. The script includes delays between requests.

### Option 2: Using Google Custom Search (Faster, Requires API Key)

1. Get Google Custom Search API key:
   - Go to https://developers.google.com/custom-search/v1/overview
   - Create a project and enable Custom Search API
   - Create credentials (API key)

2. Create a Custom Search Engine:
   - Go to https://programmablesearchengine.google.com/
   - Create a new search engine
   - Add sites: `*.justice.gov`, `*.uscourts.gov`, `courtlistener.com`
   - Get your Search Engine ID

3. Run the script:

```bash
export DATABASE_URL="postgresql://alexbeattie:xela@localhost:5432/rioters?sslmode=disable"
export GOOGLE_API_KEY="your_google_api_key"
export GOOGLE_SEARCH_ENGINE_ID="your_search_engine_id"

python3 add_charges_links.py --use-google --dry-run
python3 add_charges_links.py --use-google
```

## Options

- `--dry-run` - Preview what would be updated without making changes
- `--use-google` - Use Google Custom Search API (requires API key)
- `--google-api-key KEY` - Google API key (or use GOOGLE_API_KEY env var)
- `--limit N` - Process only first N rioters (for testing)
- `--db-url URL` - Database URL (or use DATABASE_URL env var)

## Notes

- The script includes rate limiting to be respectful to search APIs
- It prioritizes official court sources over news articles
- If no suitable link is found, the rioter is skipped (you can manually add links later)
- Processing all rioters may take a while due to rate limits

## Manual Review

After running, you can review the links:

```bash
psql -U alexbeattie -d rioters -c "SELECT first_name, last_name, charges_link FROM rioters WHERE charges_link IS NOT NULL LIMIT 20;"
```

