#!/usr/bin/env python3
"""
Add charges_link for rioters by searching for their name + "Jan 6 riot" and finding court case links.

This script searches for each rioter and attempts to find their court case/documentation links.
"""

import os
import re
import sys
import psycopg2
import time
import requests
from bs4 import BeautifulSoup
from urllib.parse import quote, urlparse

# Common sources for Jan 6 court documents
COURT_SOURCES = [
    'justice.gov',
    'courtlistener.com',
    'dcd.uscourts.gov',
    'ecf.dcd.uscourts.gov',
    'cases.justia.com',
    'lawandcrime.com',
]

def search_google(query, api_key=None):
    """Search Google using Custom Search API (requires API key and Search Engine ID)."""
    if not api_key:
        return None
    
    search_engine_id = os.getenv('GOOGLE_SEARCH_ENGINE_ID')
    if not search_engine_id:
        print("  ⚠️  GOOGLE_SEARCH_ENGINE_ID not set. To use Google search:")
        print("     1. Go to https://programmablesearchengine.google.com/controlpanel/create")
        print("     2. Create a search engine (search entire web)")
        print("     3. Copy the Search Engine ID (CX)")
        print("     4. Set: export GOOGLE_SEARCH_ENGINE_ID='your_cx_id'")
        return None
    
    try:
        url = "https://www.googleapis.com/customsearch/v1"
        params = {
            'key': api_key,
            'cx': search_engine_id,
            'q': query,
            'num': 5  # Get top 5 results
        }
        response = requests.get(url, params=params, timeout=10)
        
        # Handle rate limiting
        if response.status_code == 429:
            print(f"  ⚠️  Rate limited by Google API (429). Waiting 60 seconds...")
            time.sleep(60)  # Wait 60 seconds before retrying
            # Retry once
            response = requests.get(url, params=params, timeout=10)
            if response.status_code == 429:
                print(f"  ⚠️  Still rate limited. Skipping this search.")
                return None
        
        response.raise_for_status()
        data = response.json()
        
        results = []
        for item in data.get('items', []):
            results.append({
                'title': item.get('title', ''),
                'link': item.get('link', ''),
                'snippet': item.get('snippet', '')
            })
        return results
    except Exception as e:
        if '429' in str(e):
            print(f"  ⚠️  Rate limited by Google API. Waiting before next request...")
            time.sleep(60)
        else:
            print(f"Google API error: {e}")
        return None

def search_duckduckgo(query):
    """Search DuckDuckGo using their Lite HTML interface."""
    try:
        # Use DuckDuckGo Lite (more reliable)
        url = "https://lite.duckduckgo.com/lite/"
        data = {
            'q': query
        }
        headers = {
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
            'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
            'Content-Type': 'application/x-www-form-urlencoded',
        }
        response = requests.post(url, data=data, headers=headers, timeout=15)
        response.raise_for_status()
        soup = BeautifulSoup(response.content, 'html.parser')
        
        results = []
        # DuckDuckGo Lite uses different structure
        for result in soup.find_all('div', class_='result'):
            link_elem = result.find('a', {'class': 'result-link'})
            if link_elem:
                link = link_elem.get('href', '')
                title = link_elem.get_text(strip=True)
                snippet_elem = result.find('td', class_='result-snippet')
                snippet = snippet_elem.get_text(strip=True) if snippet_elem else ''
                
                results.append({
                    'title': title,
                    'link': link,
                    'snippet': snippet
                })
        
        # If no results from lite, try regular HTML
        if not results:
            url = "https://html.duckduckgo.com/html/"
            params = {'q': query}
            response = requests.get(url, params=params, headers=headers, timeout=15)
            response.raise_for_status()
            soup = BeautifulSoup(response.content, 'html.parser')
            
            for result in soup.find_all('div', class_='result'):
                link_elem = result.find('a', class_='result__a')
                if link_elem:
                    link = link_elem.get('href', '')
                    # DuckDuckGo uses relative URLs, need to extract actual URL
                    if link.startswith('/l/?kh='):
                        # Extract actual URL from DuckDuckGo redirect
                        try:
                            actual_url = link.split('uddg=')[1].split('&')[0] if 'uddg=' in link else None
                            if actual_url:
                                from urllib.parse import unquote
                                link = unquote(actual_url)
                        except:
                            continue
                    
                    snippet_elem = result.find('a', class_='result__snippet')
                    results.append({
                        'title': link_elem.get_text(strip=True),
                        'link': link,
                        'snippet': snippet_elem.get_text(strip=True) if snippet_elem else ''
                    })
        
        return results[:5]  # Top 5 results
    except Exception as e:
        print(f"DuckDuckGo search error: {e}")
        return None

def try_doj_direct_link(first_name, last_name, middle_name=None):
    """Try to construct a direct DOJ link based on naming pattern."""
    # DOJ URLs follow pattern: https://www.justice.gov/usao-dc/defendants/lastname-firstname-middleinitial
    # Examples: holdridge-brent-john, boughner-tim-lavon, baker-stephen-maury
    last_lower = last_name.lower().replace(' ', '-').replace("'", '').replace('.', '')
    first_lower = first_name.lower().replace(' ', '-').replace("'", '').replace('.', '')
    
    # Try multiple DOJ URL patterns
    possible_urls = []
    
    # Pattern 1: lastname-firstname-middleinitial (if middle name exists)
    if middle_name:
        middle_initial = middle_name[0].lower() if middle_name else ''
        middle_lower = middle_name.lower().replace(' ', '-').replace("'", '').replace('.', '')
        possible_urls.append(f"https://www.justice.gov/usao-dc/defendants/{last_lower}-{first_lower}-{middle_lower}")
        possible_urls.append(f"https://www.justice.gov/usao-dc/defendants/{last_lower}-{first_lower}-{middle_initial}")
    
    # Pattern 2: lastname-firstname
    possible_urls.append(f"https://www.justice.gov/usao-dc/defendants/{last_lower}-{first_lower}")
    
    # Pattern 3: lastname-firstinitial
    if first_lower:
        possible_urls.append(f"https://www.justice.gov/usao-dc/defendants/{last_lower}-{first_lower[0]}")
    
    # Check if URL exists (use GET since DOJ blocks HEAD requests)
    for url in possible_urls:
        try:
            response = requests.get(url, timeout=5, allow_redirects=True)
            # DOJ pages return 200, redirects are fine too
            if response.status_code in [200, 301, 302]:
                # Verify it's actually a DOJ page (not a 404 page)
                if 'justice.gov' in response.url and 'defendants' in response.url:
                    return url
        except:
            continue
    
    return None

def find_best_court_link(results, rioter_name, first_name, last_name, middle_name=None):
    """Find the best court case link from search results."""
    if not results:
        return None
    
    # First, try direct DOJ link
    doj_link = try_doj_direct_link(first_name, last_name, middle_name)
    if doj_link:
        print(f"  ✓ Found direct DOJ link")
        return doj_link
    
    # Exclude news sites and prioritize court sources
    news_domains = ['washingtonpost.com', 'nytimes.com', 'cnn.com', 'foxnews.com', 
                    'nbcnews.com', 'cbsnews.com', 'abcnews.com', 'reuters.com',
                    'theguardian.com', 'usatoday.com', 'apnews.com', 'politico.com',
                    'lawandcrime.com', 'planetprinceton.com', 'albanylawreview.org']
    
    # Prioritize results from known court sources
    for result in results:
        link = result.get('link', '')
        if not link:
            continue
        
        # Skip news articles
        if any(domain in link.lower() for domain in news_domains):
            continue
        
        # Highest priority: DOJ defendant pages
        if 'justice.gov/usao-dc/defendants' in link.lower():
            print(f"  ✓ Found DOJ defendant page")
            return link
        
        # High priority: courtlistener.com
        if 'courtlistener.com' in link.lower():
            print(f"  ✓ Found courtlistener.com link")
            return link
        
        # Medium priority: other court sources
        for source in ['dcd.uscourts.gov', 'ecf.dcd.uscourts.gov', 'cases.justia.com', 'justia.com']:
            if source in link.lower():
                print(f"  ✓ Found court source link")
                return link
        
        # Check if it mentions court case, charges, or DOJ (but not news)
        snippet = (result.get('snippet', '') + ' ' + result.get('title', '')).lower()
        if any(keyword in snippet for keyword in ['court case', 'indictment', 'complaint', 'doj', 'justice.gov', 'charges filed', 'criminal complaint']):
            # Verify the name appears in the result
            name_parts = rioter_name.lower().split()
            if any(part in snippet for part in name_parts if len(part) > 2) or any(part in result.get('title', '').lower() for part in name_parts if len(part) > 2):
                # Double-check it's not a news article
                if not any(domain in link.lower() for domain in news_domains):
                    print(f"  ✓ Found court-related link")
                    return link
    
    # Don't use fallback - only accept verified court sources
    # If we can't find a proper court link, return None
    return None

def update_charges_links(db_conn, use_google_api=False, google_api_key=None, dry_run=False, limit=None, update_all=False):
    """Update charges_link for all rioters missing links (or all rioters if update_all=True)."""
    cursor = db_conn.cursor()
    
    # Get all rioters - either without links or all if update_all is True
    if update_all:
        query = """
            SELECT id, first_name, last_name, middle_name, charges_link
            FROM rioters
            ORDER BY id
        """
    else:
        query = """
            SELECT id, first_name, last_name, middle_name, charges_link
            FROM rioters
            WHERE charges_link IS NULL OR charges_link = ''
            ORDER BY id
        """
    if limit:
        query += f" LIMIT {limit}"
    
    cursor.execute(query)
    
    rioters = cursor.fetchall()
    if update_all:
        print(f"Found {len(rioters)} rioters to check/update")
    else:
        print(f"Found {len(rioters)} rioters without charges_link")
    
    updated = 0
    skipped = 0
    errors = 0
    
    for rioter_id, first_name, last_name, middle_name, current_link in rioters:
        try:
            # Handle malformed names where suffixes might be in wrong fields
            # Clean up names - remove common suffixes from last_name
            clean_last = last_name.strip()
            clean_middle = (middle_name or '').strip()
            clean_first = first_name.strip()
            
            # If last_name looks like a suffix (III, Sr., Jr., etc), use middle_name as last_name
            suffix_patterns = ['III', 'II', 'Sr.', 'Sr', 'Jr.', 'Jr', 'IV', 'V']
            if any(clean_last.endswith(suffix) for suffix in suffix_patterns) and clean_middle:
                # Likely the actual last name is in middle_name
                parts = clean_middle.split()
                if len(parts) > 1:
                    clean_last = parts[-1]  # Last part of middle_name is likely the real last name
                    clean_middle = ' '.join(parts[:-1]) if len(parts) > 1 else ''
                elif len(parts) == 1:
                    clean_last = parts[0]
                    clean_middle = ''
            
            # Clean up any remaining issues (commas, parentheses, etc)
            clean_last = clean_last.rstrip('),').strip()
            clean_middle = clean_middle.rstrip('),').strip()
            
            # Build full name for display
            full_name = f"{clean_first} {clean_last}".strip()
            if clean_middle:
                full_name = f"{clean_first} {clean_middle} {clean_last}".strip()
            
            # Use cleaned names for DOJ link checking
            actual_first = clean_first
            actual_last = clean_last
            actual_middle = clean_middle if clean_middle else None
            
            # Build search query - prioritize court sources
            # Try DOJ-specific search first
            search_query = f'"{full_name}" site:justice.gov/usao-dc/defendants "Jan 6"'
            
            # If that doesn't work, we'll try broader search with court site filters
            # But we'll be strict about what we accept
            
            print(f"\nSearching for: {full_name}")
            print(f"  Query: {search_query}")
            
            # Search for court case link - try DOJ-specific first
            results = None
            if use_google_api and google_api_key:
                results = search_google(search_query, google_api_key)
                time.sleep(2)  # Increased delay to avoid rate limits (Google allows ~100 queries per 100 seconds)
                
                # If no DOJ results, try broader search but only on court sites
                if not results or len(results) == 0:
                    search_query_broad = f'"{full_name}" "Jan 6" (site:justice.gov OR site:courtlistener.com OR site:uscourts.gov OR site:justia.com) charges indictment'
                    print(f"  Trying broader court-only search...")
                    results = search_google(search_query_broad, google_api_key)
                    time.sleep(2)
                    
                    # If still no results, try one more time with just the name on court sites
                    if not results or len(results) == 0:
                        search_query_name_only = f'"{full_name}" site:justice.gov/usao-dc/defendants'
                        print(f"  Trying name-only DOJ search...")
                        results = search_google(search_query_name_only, google_api_key)
                        time.sleep(2)
            else:
                results = search_duckduckgo(search_query)
                time.sleep(3)  # Rate limiting for DuckDuckGo (be respectful)
            
            if not results:
                print(f"  ⚠️  No search results found")
                skipped += 1
                continue
            
            # Show what results were found (for debugging)
            if len(results) > 0:
                print(f"  Found {len(results)} search results")
            
            # Find best court link (use cleaned names for DOJ checking)
            court_link = find_best_court_link(results, full_name, actual_first, actual_last, actual_middle)
            
            if not court_link and len(results) > 0:
                print(f"  ⚠️  No suitable court link found in results")
                for i, result in enumerate(results[:3], 1):
                    title = result.get('title', 'No title')[:60]
                    link = result.get('link', '')[:80]
                    print(f"    {i}. {title}...")
                    print(f"       {link}...")
            
            if court_link:
                # Verify the link is actually accessible before updating
                try:
                    verify_response = requests.get(court_link, timeout=5, allow_redirects=True)
                    if verify_response.status_code == 200:
                        # Check if it's actually a valid court page (not a 404 page)
                        content_lower = verify_response.text.lower()
                        if '404' not in content_lower[:500] and 'not found' not in content_lower[:500]:
                            print(f"  ✅ Found verified link: {court_link}")
                            
                            if dry_run:
                                print(f"  [DRY RUN] Would update charges_link")
                                updated += 1
                            else:
                                cursor.execute("""
                                    UPDATE rioters 
                                    SET charges_link = %s 
                                    WHERE id = %s
                                """, (court_link, rioter_id))
                                updated += 1
                        else:
                            print(f"  ⚠️  Link appears to be 404/invalid: {court_link}")
                            skipped += 1
                    else:
                        print(f"  ⚠️  Link returned status {verify_response.status_code}: {court_link}")
                        skipped += 1
                except Exception as verify_error:
                    print(f"  ⚠️  Could not verify link: {str(verify_error)[:50]}")
                    skipped += 1
            else:
                print(f"  ⚠️  No suitable court link found in results")
                # Show what we found
                for i, result in enumerate(results[:3], 1):
                    print(f"    {i}. {result.get('title', 'No title')[:60]}...")
                    print(f"       {result.get('link', '')[:80]}...")
                skipped += 1
            
        except Exception as e:
            print(f"  ❌ Error processing {full_name}: {e}")
            errors += 1
            continue
    
    if not dry_run:
        db_conn.commit()
    
    cursor.close()
    
    print(f"\n{'='*60}")
    print(f"Results:")
    print(f"  Updated: {updated}")
    print(f"  Skipped: {skipped}")
    print(f"  Errors: {errors}")
    
    return updated, skipped, errors

def main():
    """Main function."""
    import argparse
    
    parser = argparse.ArgumentParser(description='Add charges_link for rioters')
    parser.add_argument('--dry-run', action='store_true', help='Dry run - do not update database')
    parser.add_argument('--db-url', help='Database URL (or use DATABASE_URL env var)')
    parser.add_argument('--google-api-key', help='Google Custom Search API key (or use GOOGLE_API_KEY env var)')
    parser.add_argument('--use-google', action='store_true', help='Use Google Custom Search API (requires API key)')
    parser.add_argument('--limit', type=int, help='Limit number of rioters to process (for testing)')
    parser.add_argument('--update-all', action='store_true', help='Update ALL rioters (including those with existing links) - use to fix broken links')
    args = parser.parse_args()
    
    # Get database URL
    db_url = args.db_url or os.getenv('DATABASE_URL')
    if not db_url:
        print("Error: DATABASE_URL not set. Use --db-url or set DATABASE_URL environment variable.")
        sys.exit(1)
    
    # Get Google API key if using Google
    google_api_key = None
    if args.use_google:
        google_api_key = args.google_api_key or os.getenv('GOOGLE_API_KEY')
        if not google_api_key:
            print("Warning: GOOGLE_API_KEY not set. Will use DuckDuckGo instead.")
            print("To use Google: Get API key from https://developers.google.com/custom-search/v1/overview")
            print("And create a Custom Search Engine at https://programmablesearchengine.google.com/")
    
    # Connect to database
    try:
        conn = psycopg2.connect(db_url)
        print("Connected to database")
    except Exception as e:
        print(f"Error connecting to database: {e}")
        sys.exit(1)
    
    # Update charges links
    mode = "DRY RUN" if args.dry_run else ("UPDATING ALL" if args.update_all else "UPDATING")
    print(f"[{mode}] Updating charges links...")
    updated, skipped, errors = update_charges_links(
        conn, 
        use_google_api=args.use_google,
        google_api_key=google_api_key,
        dry_run=args.dry_run,
        limit=args.limit,
        update_all=args.update_all
    )
    
    conn.close()

if __name__ == '__main__':
    main()

