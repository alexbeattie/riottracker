# Quick Setup: Google Custom Search Engine ID

You already have your Google API key: `AIzaSyD3zTZ22eKPzJaHQYAZxNyfjhvCyJSJgGk`

## Step 1: Create a Custom Search Engine

1. Go to: https://programmablesearchengine.google.com/controlpanel/create

2. Fill in the form:
   - **Search engine name**: `Jan 6 Court Cases` (or any name)
   - **Sites to search**: Leave this **EMPTY** (to search the entire web)
     - OR optionally add: `*.justice.gov`, `*.uscourts.gov`, `courtlistener.com`
   - Click **"Create"**

3. After creation, you'll see your **Search Engine ID** (CX)
   - It looks like: `017576662512468239146:abc123xyz`
   - Copy this ID

## Step 2: Run the Script

```bash
cd /Users/alexbeattie/Developer/riottracker/scripts

export DATABASE_URL="postgresql://alexbeattie:xela@localhost:5432/rioters?sslmode=disable"
export GOOGLE_API_KEY="AIzaSyD3zTZ22eKPzJaHQYAZxNyfjhvCyJSJgGk"
export GOOGLE_SEARCH_ENGINE_ID="a46dceb2ecaa34d02"

# Test with 5 records first
python3 add_charges_links.py --use-google --dry-run --limit 5

# If it works, run for all records
python3 add_charges_links.py --use-google --dry-run
python3 add_charges_links.py --use-google
```

## Alternative: Use DuckDuckGo (No API Key Needed)

If you don't want to set up Google Custom Search, you can use DuckDuckGo (slower, but free):

```bash
export DATABASE_URL="postgresql://alexbeattie:xela@localhost:5432/rioters?sslmode=disable"
python3 add_charges_links.py --dry-run --limit 5
```

Note: DuckDuckGo may have rate limits and connection issues.

