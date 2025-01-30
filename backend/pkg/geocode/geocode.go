package geocode

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "net/url"
    "os"
    "time"
)

type MapboxGeocodeResponse struct {
    Features []struct {
        Geometry struct {
            Coordinates []float64 `json:"coordinates"`
        } `json:"geometry"`
    } `json:"features"`
}

func GeocodeWithMapbox(city, state string) (*float64, *float64, error) {
    mapboxToken := os.Getenv("MAPBOX_ACCESS_TOKEN")
    if mapboxToken == "" {
        return nil, nil, fmt.Errorf("MAPBOX_ACCESS_TOKEN not set")
    }

    baseURL := "https://api.mapbox.com/geocoding/v5/mapbox.places/"
    var query string
    if city == "" {
        // If only state is available, append "state" to improve accuracy
        query = fmt.Sprintf("%s state.json", url.QueryEscape(state))
    } else {
        query = fmt.Sprintf("%s,%s.json", url.QueryEscape(city), url.QueryEscape(state))
    }    
    requestURL := fmt.Sprintf("%s%s?access_token=%s&country=US&limit=1", 
        baseURL, 
        query, 
        mapboxToken,
    )

    req, err := http.NewRequest("GET", requestURL, nil)
    if err != nil {
        return nil, nil, err
    }
    req.Header.Set("User-Agent", "RiotersDatabase/1.0")

    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    resp, err := client.Do(req)
    if err != nil {
        return nil, nil, err
    }
    defer resp.Body.Close()

    var geocodeResp MapboxGeocodeResponse
    if err := json.NewDecoder(resp.Body).Decode(&geocodeResp); err != nil {
        return nil, nil, err
    }

    if len(geocodeResp.Features) == 0 {
        return nil, nil, fmt.Errorf("no coordinates found for %s, %s", city, state)
    }

    // Mapbox returns [longitude, latitude]
    lon := geocodeResp.Features[0].Geometry.Coordinates[0]
    lat := geocodeResp.Features[0].Geometry.Coordinates[1]

    return &lat, &lon, nil
}

// BatchUpdateCoordinates updates coordinates for rioters without lat/lon
func BatchUpdateCoordinates(db *sql.DB) error {
    // Fetch rioters without coordinates
    rows, err := db.Query("SELECT id, city, state FROM rioters WHERE latitude IS NULL OR longitude IS NULL")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var city, state string
        
        if err := rows.Scan(&id, &city, &state); err != nil {
            log.Println("Error scanning row:", err)
            continue
        }

        // Skip if city or state is empty
        if city == "" || state == "" {
            continue
        }

        // Geocode the address
        lat, lon, err := GeocodeWithMapbox(city, state)
        if err != nil {
            log.Printf("Geocoding error for %s, %s: %v", city, state, err)
            continue
        }

        // Update the database with both lat/long and geometry
        _, err = db.Exec(`
            UPDATE rioters 
            SET 
                latitude = $1, 
                longitude = $2,
                geom = ST_SetSRID(ST_MakePoint($2, $1), 4326)
            WHERE id = $3`, 
            lat, lon, id)
        if err != nil {
            log.Printf("Error updating coordinates for ID %d: %v", id, err)
        }

        // Be kind to the API
        time.Sleep(250 * time.Millisecond)
    }

    return nil
}
func UpdateAllGeometries(db *sql.DB) error {
    _, err := db.Exec(`
        UPDATE rioters 
        SET geom = ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
        WHERE latitude IS NOT NULL 
        AND longitude IS NOT NULL 
        AND geom IS NULL`)
    return err
}

