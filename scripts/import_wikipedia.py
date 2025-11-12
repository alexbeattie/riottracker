#!/usr/bin/env python3
"""
Import January 6 Rioters data from Wikipedia into the database.

This script scrapes data from the 4 Wikipedia pages:
- A-F: https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(A-F)
- G-L: https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(G-L)
- M-S: https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(M-S)
- T-Z: https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(T-Z)
"""

import os
import re
import sys
import psycopg2
from datetime import datetime
from urllib.parse import urljoin, quote
import requests
from bs4 import BeautifulSoup
import time

# Wikipedia URLs (properly encoded)
WIKIPEDIA_URLS = [
    "https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(A-F)",
    "https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(G-L)",
    "https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(M-S)",
    "https://en.wikipedia.org/wiki/List_of_cases_of_the_January_6_United_States_Capitol_attack_(T-Z)",
]

def parse_name(full_name):
    """Parse full name into first, last, and optional middle name."""
    if not full_name or not full_name.strip():
        return None, None, None
    
    # Remove Wikipedia links and formatting
    name = re.sub(r'\[\[.*?\|(.*?)\]\]', r'\1', full_name)  # [[Link|Text]] -> Text
    name = re.sub(r'\[\[(.*?)\]\]', r'\1', name)  # [[Text]] -> Text
    name = re.sub(r'<.*?>', '', name)  # Remove HTML tags
    
    parts = name.strip().split()
    
    if len(parts) == 0:
        return None, None, None
    elif len(parts) == 1:
        return parts[0], None, None
    elif len(parts) == 2:
        return parts[0], parts[1], None
    else:
        # First name, middle name(s), last name
        return parts[0], parts[-1], ' '.join(parts[1:-1]) if len(parts) > 2 else None

def parse_date(date_str):
    """Parse date string to date object."""
    if not date_str:
        return None
    
    # Common date formats
    formats = [
        "%B %d, %Y",  # January 18, 2023
        "%B %d %Y",   # January 18 2023
        "%m/%d/%Y",   # 01/18/2023
        "%Y-%m-%d",   # 2023-01-18
    ]
    
    for fmt in formats:
        try:
            return datetime.strptime(date_str.strip(), fmt).date()
        except ValueError:
            continue
    
    return None

def extract_location(text):
    """Try to extract city and state from text."""
    if not text:
        return None, None
    
    # Common patterns: "City, State" or "from City, State"
    patterns = [
        r'from\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*),\s+([A-Z]{2}|[A-Z][a-z]+)',
        r'([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*),\s+([A-Z]{2}|[A-Z][a-z]+)',
        r'resident of\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*),\s+([A-Z]{2}|[A-Z][a-z]+)',
    ]
    
    for pattern in patterns:
        match = re.search(pattern, text)
        if match:
            city = match.group(1)
            state = match.group(2)
            # Normalize state abbreviations
            if len(state) == 2:
                state = state.upper()
            return city, state
    
    return None, None

def parse_charges(charges_text):
    """Parse charges and extract flags."""
    if not charges_text:
        return "", False, False, False, False
    
    charges = charges_text
    
    # Extract charge type flags
    violence_assault = bool(re.search(r'assault|violence|resisting|impeding|obstruction', charges_text, re.I))
    conspiracy = bool(re.search(r'conspiracy', charges_text, re.I))
    theft = bool(re.search(r'theft|steal', charges_text, re.I))
    property = bool(re.search(r'property|damage|destruction', charges_text, re.I))
    
    return charges, violence_assault, conspiracy, theft, property

def parse_case_status(pleas_text, judgment_text, notes_text):
    """Parse case status from pleas, judgment, and notes."""
    combined = f"{pleas_text or ''} {judgment_text or ''} {notes_text or ''}"
    combined_lower = combined.lower()
    
    # Determine case status
    if 'pardoned' in combined_lower or 'pardon' in combined_lower:
        return 'pardoned', True, False, True
    elif 'commuted' in combined_lower or 'commutation' in combined_lower:
        return 'commuted', False, True, False
    elif 'sentenced' in combined_lower:
        return 'sentenced', False, False, True
    elif 'pleaded guilty' in combined_lower or 'pleaded not guilty' in combined_lower:
        if 'guilty' in combined_lower:
            return 'pleaded-guilty', False, False, False
        else:
            return 'pleaded-not-guilty', False, False, False
    elif 'convicted' in combined_lower:
        return 'convicted', False, False, True
    elif 'acquitted' in combined_lower:
        return 'acquitted', False, False, False
    elif 'dismissed' in combined_lower:
        return 'dismissed', False, False, False
    
    return None, False, False, False

def geocode_location(city, state, mapbox_token):
    """Geocode city and state using Mapbox API."""
    if not city or not state:
        return None, None
    
    if not mapbox_token:
        print(f"Warning: No Mapbox token, skipping geocoding for {city}, {state}")
        return None, None
    
    try:
        url = f"https://api.mapbox.com/geocoding/v5/mapbox.places/{city},{state}.json"
        params = {
            'access_token': mapbox_token,
            'limit': 1
        }
        response = requests.get(url, params=params, timeout=5)
        response.raise_for_status()
        data = response.json()
        
        if data.get('features'):
            coords = data['features'][0]['geometry']['coordinates']
            return coords[1], coords[0]  # lat, lng
    except Exception as e:
        print(f"Geocoding error for {city}, {state}: {e}")
    
    return None, None

def scrape_wikipedia_page(url):
    """Scrape a Wikipedia page and extract rioter data."""
    print(f"Scraping {url}...")
    
    try:
        # Add headers to avoid 403 errors (Wikipedia blocks requests without proper User-Agent)
        headers = {
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
            'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
            'Accept-Language': 'en-US,en;q=0.5',
            'Accept-Encoding': 'gzip, deflate',
            'Connection': 'keep-alive',
            'Referer': 'https://www.google.com/',
        }
        
        # Use a session to maintain cookies
        session = requests.Session()
        session.headers.update(headers)
        
        response = session.get(url, timeout=30)
        response.raise_for_status()
        soup = BeautifulSoup(response.content, 'html.parser')
        
        # Find the main table
        table = soup.find('table', {'class': 'wikitable'})
        if not table:
            print(f"No table found on {url}")
            return []
        
        rows = table.find_all('tr')[1:]  # Skip header row
        rioters = []
        
        for row in rows:
            cells = row.find_all(['td', 'th'])
            if len(cells) < 6:
                continue
            
            try:
                # Extract data from cells
                arrest_date_str = cells[0].get_text(strip=True)
                name_cell = cells[1]
                charges_text = cells[2].get_text(strip=True)
                pleas_text = cells[3].get_text(strip=True)
                judgment_text = cells[4].get_text(strip=True)
                notes_text = cells[5].get_text(strip=True) if len(cells) > 5 else ""
                
                # Parse name
                name_link = name_cell.find('a')
                if name_link:
                    full_name = name_link.get_text(strip=True)
                else:
                    full_name = name_cell.get_text(strip=True)
                
                first_name, last_name, middle_name = parse_name(full_name)
                if not first_name or not last_name:
                    continue
                
                # Parse dates
                arrest_date = parse_date(arrest_date_str)
                
                # Extract location from notes
                city, state = extract_location(notes_text)
                
                # Parse charges
                charges, violence_assault, conspiracy, theft, property = parse_charges(charges_text)
                
                # Parse case status
                case_status, pardoned, commuted, sentenced = parse_case_status(
                    pleas_text, judgment_text, notes_text
                )
                
                # Combine case updates
                case_updates = f"Pleas: {pleas_text}\nJudgment: {judgment_text}\nNotes: {notes_text}".strip()
                
                rioter = {
                    'first_name': first_name,
                    'last_name': last_name,
                    'middle_name': middle_name,
                    'arrest_date': arrest_date,
                    'charges': charges,
                    'case_status': case_status,
                    'case_updates': case_updates if case_updates else None,
                    'city': city,
                    'state': state,
                    'violence_assault': violence_assault,
                    'conspiracy': conspiracy,
                    'theft': theft,
                    'property': property,
                    'pardoned': pardoned,
                    'commuted': commuted,
                    'sentenced': sentenced,
                    'notes': notes_text,
                }
                
                rioters.append(rioter)
                
            except Exception as e:
                print(f"Error parsing row: {e}")
                continue
        
        print(f"Extracted {len(rioters)} rioters from {url}")
        return rioters
        
    except Exception as e:
        print(f"Error scraping {url}: {e}")
        return []

def import_rioters(rioters, db_conn, mapbox_token=None, dry_run=False, table_name='rioters', update_existing=True):
    """Import rioters into the database.
    
    Args:
        rioters: List of rioter dictionaries
        db_conn: Database connection
        mapbox_token: Mapbox API token for geocoding
        dry_run: If True, don't actually insert/update
        table_name: Name of table to import into (default: 'rioters', can use 'rioters_import' for testing)
        update_existing: If True, update existing records; if False, skip duplicates
    """
    cursor = db_conn.cursor()
    
    imported = 0
    updated = 0
    skipped = 0
    errors = 0
    
    for rioter in rioters:
        try:
            # Check if rioter already exists (more flexible matching)
            cursor.execute(f"""
                SELECT id FROM {table_name} 
                WHERE LOWER(TRIM(first_name)) = LOWER(TRIM(%s)) 
                AND LOWER(TRIM(last_name)) = LOWER(TRIM(%s))
            """, (rioter['first_name'], rioter['last_name']))
            
            existing = cursor.fetchone()
            
            # Geocode location
            lat, lng = None, None
            if rioter['city'] and rioter['state']:
                lat, lng = geocode_location(rioter['city'], rioter['state'], mapbox_token)
                if lat and lng:
                    time.sleep(0.1)  # Rate limiting for Mapbox API
            
            # Generate photo name
            photo_name = f"{rioter['last_name'].lower()}-{rioter['first_name'].lower()}"
            if rioter['middle_name']:
                photo_name += f"-{rioter['middle_name'].lower().replace(' ', '-')}"
            photo_name += ".jpg"
            
            if existing:
                # Existing rioter found
                if not update_existing:
                    if dry_run:
                        print(f"Would skip (exists): {rioter['first_name']} {rioter['last_name']}")
                    skipped += 1
                    continue
                
                # Update existing rioter
                if dry_run:
                    print(f"Would update: {rioter['first_name']} {rioter['last_name']}")
                    updated += 1
                else:
                    cursor.execute(f"""
                        UPDATE {table_name} SET
                            middle_name = COALESCE(%s, middle_name),
                            arrest_date = COALESCE(%s, arrest_date),
                            charges = COALESCE(%s, charges),
                            case_status = COALESCE(%s, case_status),
                            case_updates = COALESCE(%s, case_updates),
                            city = COALESCE(%s, city),
                            state = COALESCE(%s, state),
                            violence_assault = %s,
                            conspiracy = %s,
                            theft = %s,
                            property = %s,
                            pardoned = %s,
                            commuted = %s,
                            sentenced = %s,
                            latitude = COALESCE(%s, latitude),
                            longitude = COALESCE(%s, longitude),
                            geom = CASE 
                                WHEN %s IS NOT NULL AND %s IS NOT NULL 
                                THEN ST_SetSRID(ST_MakePoint(%s, %s), 4326)
                                ELSE geom
                            END
                        WHERE id = %s
                    """, (
                        rioter['middle_name'], rioter['arrest_date'], rioter['charges'],
                        rioter['case_status'], rioter['case_updates'], rioter['city'],
                        rioter['state'], rioter['violence_assault'], rioter['conspiracy'],
                        rioter['theft'], rioter['property'], rioter['pardoned'],
                        rioter['commuted'], rioter['sentenced'], lat, lng,
                        lng, lat, lng, lat, existing[0]
                    ))
                    updated += 1
            else:
                # Insert new rioter
                if dry_run:
                    print(f"Would import: {rioter['first_name']} {rioter['last_name']} ({rioter['city']}, {rioter['state']})")
                    imported += 1
                else:
                    if not lat or not lng:
                        # Use default coordinates if geocoding failed
                        lat, lng = 39.8283, -98.5795  # Center of US
                    
                    cursor.execute(f"""
                        INSERT INTO {table_name} (
                            first_name, last_name, middle_name, arrest_date,
                            charges, case_status, case_updates,
                            city, state, violence_assault, conspiracy,
                            theft, property, pardoned, commuted, sentenced,
                            latitude, longitude, photo_name, geom
                        ) VALUES (
                            %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s,
                            ST_SetSRID(ST_MakePoint(%s, %s), 4326)
                        ) RETURNING id
                    """, (
                        rioter['first_name'], rioter['last_name'], rioter['middle_name'],
                        rioter['arrest_date'], rioter['charges'], rioter['case_status'],
                        rioter['case_updates'], rioter['city'], rioter['state'],
                        rioter['violence_assault'], rioter['conspiracy'], rioter['theft'],
                        rioter['property'], rioter['pardoned'], rioter['commuted'],
                        rioter['sentenced'], lat, lng, photo_name, lng, lat
                    ))
                    imported += 1
            
        except Exception as e:
            print(f"Error importing {rioter['first_name']} {rioter['last_name']}: {e}")
            errors += 1
            skipped += 1
    
    if not dry_run:
        db_conn.commit()
    
    cursor.close()
    return imported, updated, skipped, errors

def create_import_table(db_conn, table_name='rioters_import'):
    """Create a separate table for testing imports."""
    cursor = db_conn.cursor()
    
    # Check if table exists
    cursor.execute("""
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_name = %s
        )
    """, (table_name,))
    
    if cursor.fetchone()[0]:
        print(f"Table {table_name} already exists. Use --drop-table to recreate it.")
        cursor.close()
        return False
    
    # Create table with same structure as rioters
    cursor.execute(f"""
        CREATE TABLE {table_name} (LIKE rioters INCLUDING ALL);
    """)
    
    db_conn.commit()
    cursor.close()
    print(f"Created import table: {table_name}")
    return True

def main():
    """Main function."""
    import argparse
    
    parser = argparse.ArgumentParser(description='Import January 6 Rioters from Wikipedia')
    parser.add_argument('--dry-run', action='store_true', help='Dry run - do not insert into database')
    parser.add_argument('--db-url', help='Database URL (or use DATABASE_URL env var)')
    parser.add_argument('--mapbox-token', help='Mapbox token (or use MAPBOX_ACCESS_TOKEN env var)')
    parser.add_argument('--test-table', action='store_true', help='Import into separate test table (rioters_import)')
    parser.add_argument('--table-name', default='rioters', help='Table name to import into (default: rioters)')
    parser.add_argument('--no-update', action='store_true', help='Skip updating existing records (only insert new ones)')
    parser.add_argument('--create-test-table', action='store_true', help='Create a test table for imports')
    parser.add_argument('--drop-table', action='store_true', help='Drop and recreate test table (use with caution!)')
    args = parser.parse_args()
    
    # Get database URL
    db_url = args.db_url or os.getenv('DATABASE_URL')
    if not db_url:
        print("Error: DATABASE_URL not set. Use --db-url or set DATABASE_URL environment variable.")
        sys.exit(1)
    
    # Get Mapbox token (optional)
    mapbox_token = args.mapbox_token or os.getenv('MAPBOX_ACCESS_TOKEN')
    if not mapbox_token:
        print("Warning: MAPBOX_ACCESS_TOKEN not set. Geocoding will be skipped.")
    
    # Connect to database
    try:
        conn = psycopg2.connect(db_url)
        print("Connected to database")
    except Exception as e:
        print(f"Error connecting to database: {e}")
        sys.exit(1)
    
    # Determine table name
    if args.test_table or args.create_test_table:
        table_name = 'rioters_import'
    else:
        table_name = args.table_name
    
    # Create test table if requested
    if args.create_test_table:
        if args.drop_table:
            cursor = conn.cursor()
            cursor.execute(f"DROP TABLE IF EXISTS {table_name} CASCADE;")
            conn.commit()
            cursor.close()
            print(f"Dropped existing table: {table_name}")
        create_import_table(conn, table_name)
    
    # Scrape all Wikipedia pages
    all_rioters = []
    for url in WIKIPEDIA_URLS:
        rioters = scrape_wikipedia_page(url)
        all_rioters.extend(rioters)
        time.sleep(1)  # Be nice to Wikipedia
    
    print(f"\nTotal rioters extracted: {len(all_rioters)}")
    
    # Import into database
    print(f"\n{'[DRY RUN] ' if args.dry_run else ''}Importing rioters into table: {table_name}")
    imported, updated, skipped, errors = import_rioters(
        all_rioters, conn, mapbox_token, args.dry_run, 
        table_name=table_name, update_existing=not args.no_update
    )
    
    print(f"\nResults:")
    print(f"  Imported: {imported}")
    print(f"  Updated: {updated}")
    print(f"  Skipped: {skipped}")
    print(f"  Errors: {errors}")
    
    conn.close()

if __name__ == '__main__':
    main()

