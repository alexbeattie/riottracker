package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"rioters/backend/pkg/geocode"
	"strconv" // ✅ New import
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// NullBool wraps sql.NullBool to support JSON unmarshaling.
type NullBool struct {
	sql.NullBool
}
type NewRioterRequest struct {
	LastName        string  `json:"last_name" binding:"required"`
	FirstName       string  `json:"first_name" binding:"required"`
	MiddleName      *string `json:"middle_name"`
	Summary         *string `json:"summary"`
	Jurisdiction    *string `json:"jurisdiction"`
	Charges         *string `json:"charges"`
	ChargesLink     *string `json:"charges_link"`
	CaseStatus      *string `json:"case_status"`
	CaseUpdates     *string `json:"case_updates"`
	ViolenceAssault bool    `json:"violence_assault"`
	Conspiracy      bool    `json:"conspiracy"`
	Theft           bool    `json:"theft"`
	Property        bool    `json:"property"`
	Age             *int    `json:"age"`
	City            *string `json:"city"`
	State           *string `json:"state"`
	MilitaryLE      bool    `json:"military_le"`
	Extremist       bool    `json:"extremist"`
	InspiredTrump   bool    `json:"inspired_trump"`
	// You can add additional fields such as timestamps if needed.
}

// Define the Rioter struct
type Rioter struct {
	ID              int        `json:"id"`
	LastName        string     `json:"last_name"`
	FirstName       string     `json:"first_name"`
	MiddleName      *string    `json:"middle_name"`
	Summary         *string    `json:"summary"`
	Jurisdiction    *string    `json:"jurisdiction"`
	Charges         *string    `json:"charges"`
	ChargesLink     *string    `json:"charges_link"`
	CaseStatus      *string    `json:"case_status"`
	CaseUpdates     *string    `json:"case_updates"`
	ViolenceAssault NullBool   `json:"violence_assault"`
	Conspiracy      NullBool   `json:"conspiracy"`
	Theft           NullBool   `json:"theft"`
	Property        NullBool   `json:"property"`
	Age             *int       `json:"age"`
	City            *string    `json:"city"`
	State           *string    `json:"state"`
	MilitaryLE      NullBool   `json:"military_le"`
	Extremist       NullBool   `json:"extremist"`
	InspiredTrump   NullBool   `json:"inspired_trump"`
	Sentenced       NullBool   `json:"sentenced"`
	Commuted        NullBool   `json:"commuted"`
	Pardoned        NullBool   `json:"pardoned"`
	ArrestDate      *time.Time `json:"arrest_date"`
	PhotoName       *string    `json:"photo_name"`
	Latitude        *float64   `json:"latitude"`
	Longitude       *float64   `json:"longitude"`
}

type NearbyRioter struct {
	ID        int      `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	City      *string  `json:"city"`
	State     *string  `json:"state"`
	PhotoName *string  `json:"photo_name"` // ✅ Added this field
	Longitude *float64 `json:"longitude"`
	Latitude  *float64 `json:"latitude"`
	Distance  float64  `json:"distance"`
}

func (nb *NullBool) UnmarshalJSON(data []byte) error {
	// If the value is null, mark it as invalid.
	if string(data) == "null" {
		nb.Valid = false
		return nil
	}

	// Try unmarshaling into a boolean.
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return errors.New("NullBool: unable to unmarshal data into bool")
	}
	nb.Bool = b
	nb.Valid = true
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (nb NullBool) MarshalJSON() ([]byte, error) {
	if !nb.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nb.Bool)
}

// var ctx = context.Background()

func getCachedData(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return val, err
}

func setCachedData(key string, value string, ttl time.Duration) error {
	return rdb.Set(ctx, key, value, ttl).Err()
}

var (
	ctx = context.Background()
	rdb *redis.Client // Declare Redis client globally
)

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Update if Redis is remote
		Password: "",               // No password by default
		DB:       0,                // Default DB
	})

	// Test Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis") // Connect to PostgreSQL
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Using existing environment variables.")
	}
	connectionString := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatal("Database ping error:", err)
	}
	// Run coordinate migration
	if err := geocode.BatchUpdateCoordinates(db); err != nil {
		log.Fatal("Coordinate migration failed:", err)
	}
	// Add this line to update geometries for existing coordinates
	if err := geocode.UpdateAllGeometries(db); err != nil {
		log.Fatal("Geometry update failed:", err)
	}

	log.Println("Coordinate migration completed successfully")
	// Set up Gin
	router := gin.Default()
	router.Static("/photos", "./photos")
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/photos/") {
			requestedPath := filepath.Join("./photos", strings.TrimPrefix(c.Request.URL.Path, "/photos/"))

			// Add headers
			c.Header("Cache-Control", "public, max-age=86400")

			// Check if file exists
			if _, err := os.Stat(requestedPath); os.IsNotExist(err) {
				c.File("./photos/placeholder.jpg")
				c.Abort()
				return
			}
		}
		c.Next()
	})

	if err := os.MkdirAll("photos", os.ModePerm); err != nil {
		log.Fatal("Failed to create photos directory:", err)
	}

	placeholderPath := "./photos/placeholder.jpg"
	if _, err := os.Stat(placeholderPath); os.IsNotExist(err) {
		log.Printf("Warning: placeholder.jpg not found in photos directory")
	}

	// Set up Gin

	// Set up CORS first
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8081", "http://localhost:8080", "http://192.168.1.82:8081", "http://192.168.1.158:8081"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Add OPTIONS here
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Cache-Control"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/api/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})
	// API endpoint to fetch all rioters
	router.GET("/api/rioters", func(c *gin.Context) {
		// Check for "all" parameter first
		fetchAll := c.Query("all") == "true"

		// Build base WHERE clause and args for both count and data queries
		whereClause := "WHERE 1=1"
		var args []interface{}
		argCounter := 1

		if state := c.Query("state"); state != "" {
			whereClause += fmt.Sprintf(" AND state = $%d", argCounter)
			args = append(args, state)
			argCounter++
		}

		if city := c.Query("city"); city != "" {
			whereClause += fmt.Sprintf(" AND city = $%d", argCounter)
			args = append(args, city)
			argCounter++
		}

		// Get filtered count
		var total int
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM rioters %s", whereClause)
		if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
			log.Printf("Count query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Build the data query using same WHERE clause
		query := fmt.Sprintf(`
        SELECT id, last_name, first_name, middle_name, summary, 
               jurisdiction, charges, charges_link, case_status, 
               case_updates, violence_assault, conspiracy, theft, 
               property, age, city, state, military_le, extremist, 
               inspired_trump, sentenced, commuted, pardoned, 
               arrest_date, photo_name,
               ST_X(geom::geometry) as longitude,
               ST_Y(geom::geometry) as latitude
        FROM rioters
        %s
    `, whereClause)

		// Add ORDER BY
		query += " ORDER BY id"

		// Add pagination only if not fetching all
		if !fetchAll {
			pageStr := c.DefaultQuery("page", "1")
			pageSizeStr := c.DefaultQuery("page_size", "500")

			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
				return
			}

			pageSize, err := strconv.Atoi(pageSizeStr)
			if err != nil || pageSize < 1 || pageSize > 1601 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size (1-1000)"})
				return
			}

			offset := (page - 1) * pageSize
			args = append(args, pageSize, offset)
			query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))

		}

		// Execute the query
		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var rioters []Rioter
		for rows.Next() {
			var r Rioter
			if err := rows.Scan(
				&r.ID, &r.LastName, &r.FirstName, &r.MiddleName,
				&r.Summary, &r.Jurisdiction, &r.Charges, &r.ChargesLink,
				&r.CaseStatus, &r.CaseUpdates, &r.ViolenceAssault,
				&r.Conspiracy, &r.Theft, &r.Property, &r.Age, &r.City,
				&r.State, &r.MilitaryLE, &r.Extremist, &r.InspiredTrump,
				&r.Sentenced, &r.Commuted, &r.Pardoned, &r.ArrestDate,
				&r.PhotoName, &r.Longitude, &r.Latitude,
			); err != nil {
				log.Printf("Row scan error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			rioters = append(rioters, r)
		}

		if fetchAll {
			c.JSON(http.StatusOK, gin.H{
				"data":  rioters,
				"total": total,
			})
		} else {
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
			c.JSON(http.StatusOK, gin.H{
				"data":      rioters,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
				"pages":     int(math.Ceil(float64(total) / float64(pageSize))),
			})
		}
	})

	// SEARCH DATABASE FOR RIOTERS with "term" query parameter
	router.GET("/api/search/suggestions", func(c *gin.Context) {
		term := c.Query("term")
		if term == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search term is required"})
			return
		}

		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}

		pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
		if err != nil || pageSize < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size"})
			return
		}

		offset := (page - 1) * pageSize

		query := `
		SELECT DISTINCT
			ts_headline(
				first_name || ' ' || last_name || ' ' || city,
				websearch_to_tsquery('english', $1),
				'StartSel=<mark>, StopSel=</mark>'
			) AS suggestion,
			ts_rank(search_vector, websearch_to_tsquery('english', $1)) AS rank
		FROM rioters
	WHERE search_vector @@ websearch_to_tsquery('english', $1)
	ORDER BY rank DESC
	LIMIT $2 OFFSET $3
	`

		rows, err := db.Query(query, term, pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var suggestions []string
		for rows.Next() {
			var suggestion string
			var rank float64
			if err := rows.Scan(&suggestion, &rank); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			suggestions = append(suggestions, suggestion)
		}

		if len(suggestions) == 0 {
			suggestions = []string{}
		}

		c.JSON(http.StatusOK, suggestions)
	})

	router.GET("/api/rioters/count-by-state", func(c *gin.Context) {
		cacheKey := "count_by_state"

		// Check Redis cache
		cached, err := getCachedData(cacheKey)
		if err == nil && cached != "" {
			c.JSON(http.StatusOK, json.RawMessage(cached)) // Serve cached data
			return
		}
		rows, err := db.Query("SELECT state, COUNT(*) FROM rioters GROUP BY state") // ✅ Use 'state' instead of 'location'
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		stateCounts := make(map[string]int)
		for rows.Next() {
			var state string
			var count int
			if err := rows.Scan(&state, &count); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			stateCounts[state] = count
		}
		// Cache the result for 10 minutes
		jsonData, _ := json.Marshal(stateCounts)
		setCachedData(cacheKey, string(jsonData), 10*time.Minute)

		c.JSON(http.StatusOK, stateCounts)
	})
	router.GET("/api/rioters/nearby", func(c *gin.Context) {
		latStr, lngStr := c.Query("lat"), c.Query("lng")
		radiusStr := c.DefaultQuery("radius", "50000")

		cacheKey := fmt.Sprintf("nearby_%s_%s_%s", latStr, lngStr, radiusStr)
		// Check Redis cache first
		cached, err := getCachedData(cacheKey)
		if err == nil && cached != "" {
			c.JSON(http.StatusOK, json.RawMessage(cached))
			return
		}
		// Convert query parameters to float64
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			log.Printf("Invalid latitude: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
			return
		}

		lng, err := strconv.ParseFloat(lngStr, 64)
		if err != nil {
			log.Printf("Invalid longitude: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
			return
		}

		radius, err := strconv.ParseFloat(radiusStr, 64)
		if err != nil {
			log.Printf("Invalid radius: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid radius"})
			return
		}

		log.Printf("Received request: lat=%f, lng=%f, radius=%f", lat, lng, radius)

		query := `
    SELECT 
        id, first_name, last_name, city, state,
        photo_name,
        ST_X(geom::geometry) as longitude,
        ST_Y(geom::geometry) as latitude,
        ST_Distance(
            geom::geography,
            ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
        ) as distance
    FROM rioters
    WHERE ST_DWithin(
        geom::geography,
        ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
        $3
    )
    ORDER BY distance;
`
		log.Printf("Executing query: %s", query) // Log the SQL query

		rows, err := db.Query(query, lng, lat, radius) // Correct parameter order
		if err != nil {
			log.Printf("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var nearbyRioters []NearbyRioter
		if math.Abs(lng) > 180 || math.Abs(lat) > 90 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coordinates"})
			return
		}

		if radius <= 0 || radius > 1000000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Radius must be between 1 and 1,000,000 meters"})
			return
		}
		for rows.Next() {
			var r NearbyRioter
			if err := rows.Scan(
				&r.ID,
				&r.FirstName,
				&r.LastName,
				&r.City,
				&r.State,
				&r.PhotoName, // ✅ Added this
				&r.Longitude,
				&r.Latitude,
				&r.Distance,
			); err != nil {

				log.Printf("Row scan error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			nearbyRioters = append(nearbyRioters, r) // ✅ Fix: Append to slice
		}
		if err != nil {
			log.Printf("Query error: %v\nQuery: %s\nParams: %v", err, query, []interface{}{lng, lat, radius})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		log.Printf("Successfully fetched %d nearby rioters", len(nearbyRioters))
		jsonData, _ := json.Marshal(nearbyRioters)
		setCachedData(cacheKey, string(jsonData), 10*time.Minute)

		c.JSON(http.StatusOK, nearbyRioters)
	})

	router.GET("/api/rioters/clusters", func(c *gin.Context) {
		// Get zoom as a string from the query parameters
		zoomStr := c.DefaultQuery("zoom", "10")

		// Convert zoom from string to float64
		zoom, err := strconv.ParseFloat(zoomStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid zoom value"})
			return
		}

		gridSize := 0.01 / math.Pow(2, zoom) // Properly scale grid size by zoom

		log.Printf("Zoom Level: %f, Grid Size: %s", zoom, gridSize)

		query := `
        SELECT 
            ST_AsGeoJSON(ST_Centroid(cluster_geom)) as center,
            json_agg(json_build_object(
                'id', id,
                'first_name', first_name,
                'last_name', last_name,
                'city', city,
                'state', state
            )) as points,
            COUNT(*) as point_count
        FROM (
            SELECT 
                id, first_name, last_name, city, state,
                ST_CollectionExtract(ST_Collect(geom)) as cluster_geom
            FROM rioters
            WHERE geom IS NOT NULL
            GROUP BY ST_SnapToGrid(geom, $1)
        ) clusters
        GROUP BY cluster_geom;
    `

		rows, err := db.Query(query, gridSize)
		if err != nil {
			log.Printf("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var clusters []map[string]interface{}
		for rows.Next() {
			var center string
			var points string
			var pointCount int

			if err := rows.Scan(&center, &points, &pointCount); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			cluster := map[string]interface{}{
				"center":      center,
				"points":      points,
				"point_count": pointCount,
			}
			clusters = append(clusters, cluster)
		}

		c.JSON(http.StatusOK, clusters)
	})

	router.GET("/api/rioters/by-state", func(c *gin.Context) {
		state := c.Query("state")
		if state == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "State parameter is required"})
			return
		}

		query := `
        SELECT 
            city,
            json_agg(json_build_object(
                'id', id,
                'first_name', first_name,
                'last_name', last_name,
                'latitude', ST_Y(geom::geometry),
                'longitude', ST_X(geom::geometry)
            )) AS markers
        FROM rioters
        WHERE state = $1
        GROUP BY city
    `

		rows, err := db.Query(query, state)
		if err != nil {
			log.Printf("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		defer rows.Close()

		var results []map[string]interface{}
		for rows.Next() {
			var city string
			var markers string
			if err := rows.Scan(&city, &markers); err != nil {
				log.Printf("Row scan error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}

			result := map[string]interface{}{
				"city":    city,
				"markers": json.RawMessage(markers),
			}
			results = append(results, result)
		}

		c.JSON(http.StatusOK, results)
	})

	// POST endpoint to add a new rioter record
	// POST endpoint to add a new rioter record
	router.POST("/api/rioters", func(c *gin.Context) {
		var newRioter NewRioterRequest
		if err := c.ShouldBindJSON(&newRioter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Safely extract city and state (using empty string if nil)
		city := ""
		if newRioter.City != nil {
			city = *newRioter.City
		}
		state := ""
		if newRioter.State != nil {
			state = *newRioter.State
		}

		// Compute the photo name based on last_name, first_name, and middle_name.
		// For example: "alam-zachary-jordan.jpg"
		photoName := strings.ToLower(newRioter.LastName) + "-" + strings.ToLower(newRioter.FirstName)
		if newRioter.MiddleName != nil && *newRioter.MiddleName != "" {
			photoName += "-" + strings.ToLower(*newRioter.MiddleName)
		}
		photoName += ".jpg"

		// Call your Mapbox geocoding function to get coordinates.
		lat, lon, err := geocode.GeocodeWithMapbox(city, state)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Geocoding error: " + err.Error()})
			return
		}

		log.Printf("Geocoded coordinates for %s, %s: lat=%v, lon=%v", city, state, *lat, *lon)

		// Insert the new record using the computed lat/lon and photoName.
		var newID int
		err = db.QueryRow(`
        INSERT INTO rioters (
            last_name, first_name, middle_name, summary, jurisdiction, charges,
            charges_link, case_status, case_updates, violence_assault, conspiracy,
            theft, property, age, city, state, military_le, extremist, inspired_trump,
            latitude, longitude, photo_name, geom
        )
        VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, $10, $11,
            $12, $13, $14, $15, $16, $17, $18, $19,
            $20, $21, $22,
            ST_SetSRID(ST_MakePoint($21, $20), 4326)
        )
        RETURNING id
    `,
			newRioter.LastName, newRioter.FirstName, newRioter.MiddleName, newRioter.Summary, newRioter.Jurisdiction, newRioter.Charges,
			newRioter.ChargesLink, newRioter.CaseStatus, newRioter.CaseUpdates, newRioter.ViolenceAssault, newRioter.Conspiracy,
			newRioter.Theft, newRioter.Property, newRioter.Age, newRioter.City, newRioter.State, newRioter.MilitaryLE, newRioter.Extremist, newRioter.InspiredTrump,
			lat, lon, photoName,
		).Scan(&newID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status": "success",
			"id":     newID,
		})
	})
	// PUT endpoint to update a rioter record
	router.PUT("/api/rioters/:id", func(c *gin.Context) {
		id := c.Param("id")
		var updatedRioter Rioter
		if err := c.ShouldBindJSON(&updatedRioter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Recompute photo_name from the updated fields.
		// (Assuming updatedRioter has the new values.)
		photoName := strings.ToLower(updatedRioter.LastName) + "-" + strings.ToLower(updatedRioter.FirstName)
		if updatedRioter.MiddleName != nil && *updatedRioter.MiddleName != "" {
			photoName += "-" + strings.ToLower(*updatedRioter.MiddleName)
		}
		photoName += ".jpg"

		// Update the record in the database.
		_, err := db.Exec(`
        UPDATE rioters SET
            last_name = $1,
            first_name = $2,
            middle_name = $3,
            summary = $4,
            jurisdiction = $5,
            charges = $6,
            charges_link = $7,
            case_status = $8,
            case_updates = $9,
            violence_assault = $10,
            conspiracy = $11,
            theft = $12,
            property = $13,
            age = $14,
            city = $15,
            state = $16,
            military_le = $17,
            extremist = $18,
            inspired_trump = $19,
            sentenced = $20,
            commuted = $21,
            pardoned = $22,
            arrest_date = $23,
            photo_name = $24,
            latitude = $25,
            longitude = $26,
            geom = ST_SetSRID(ST_MakePoint($26, $25), 4326)
        WHERE id = $27
    `,
			updatedRioter.LastName, updatedRioter.FirstName, updatedRioter.MiddleName, updatedRioter.Summary, updatedRioter.Jurisdiction, updatedRioter.Charges,
			updatedRioter.ChargesLink, updatedRioter.CaseStatus, updatedRioter.CaseUpdates, updatedRioter.ViolenceAssault, updatedRioter.Conspiracy,
			updatedRioter.Theft, updatedRioter.Property, updatedRioter.Age, updatedRioter.City, updatedRioter.State, updatedRioter.MilitaryLE, updatedRioter.Extremist, updatedRioter.InspiredTrump,
			updatedRioter.Sentenced, updatedRioter.Commuted, updatedRioter.Pardoned, updatedRioter.ArrestDate, photoName, updatedRioter.Latitude, updatedRioter.Longitude, id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"id":     id,
		})
	})
	// Start server
	router.GET("/api/rioters/:id", func(c *gin.Context) {
		id := c.Param("id")

		log.Printf("Fetching rioter with ID: %s", id)

		// Query to get a single rioter
		query := `
        SELECT id, last_name, first_name, middle_name, summary, 
               jurisdiction, charges, charges_link, case_status, 
               case_updates, violence_assault, conspiracy, theft, 
               property, age, city, state, military_le, extremist, 
               inspired_trump, sentenced, commuted, pardoned, 
               arrest_date, photo_name,
               ST_X(geom::geometry) as longitude,
               ST_Y(geom::geometry) as latitude
        FROM rioters
        WHERE id = $1
    `

		var rioter Rioter
		err := db.QueryRow(query, id).Scan(
			&rioter.ID, &rioter.LastName, &rioter.FirstName, &rioter.MiddleName,
			&rioter.Summary, &rioter.Jurisdiction, &rioter.Charges, &rioter.ChargesLink,
			&rioter.CaseStatus, &rioter.CaseUpdates, &rioter.ViolenceAssault,
			&rioter.Conspiracy, &rioter.Theft, &rioter.Property, &rioter.Age, &rioter.City,
			&rioter.State, &rioter.MilitaryLE, &rioter.Extremist, &rioter.InspiredTrump,
			&rioter.Sentenced, &rioter.Commuted, &rioter.Pardoned, &rioter.ArrestDate,
			&rioter.PhotoName, &rioter.Longitude, &rioter.Latitude,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Rioter with ID %s not found", id)
				c.JSON(http.StatusNotFound, gin.H{"error": "Rioter not found"})
				return
			}
			log.Printf("Database error for ID %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query error"})
			return
		}

		log.Printf("Successfully fetched rioter: %+v", rioter)
		c.JSON(http.StatusOK, rioter)
	})

	router.Run(":8080")
}
