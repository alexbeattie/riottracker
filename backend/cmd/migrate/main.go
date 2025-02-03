package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"rioters/backend/pkg/geocode"
	"strconv" // ✅ New import
	"strings"
	"time"

	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
	ViolenceAssault bool       `json:"violence_assault"`
	Conspiracy      bool       `json:"conspiracy"`
	Theft           bool       `json:"theft"`
	Property        bool       `json:"property"`
	Age             *int       `json:"age"`
	City            *string    `json:"city"`
	State           *string    `json:"state"`
	MilitaryLE      bool       `json:"military_le"`
	Extremist       bool       `json:"extremist"`
	InspiredTrump   bool       `json:"inspired_trump"`
	Sentenced       bool       `json:"sentenced"`
	Commuted        bool       `json:"commuted"`
	Pardoned        bool       `json:"pardoned"`
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

func main() {
	// Connect to PostgreSQL
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Using existing environment variables.")
	}
	connectionString := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		log.Fatal("Database connection error:", err)
	}
	defer db.Close()

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
		AllowOrigins:     []string{"http://localhost:8081","http://192.168.1.82:8081"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
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
		// Pagination parameters
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "20")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size (1-1000)"})
			return
		}

		offset := (page - 1) * pageSize

		// Get total count
		var total int
		countQuery := `SELECT COUNT(*) FROM rioters`
		if err := db.QueryRow(countQuery).Scan(&total); err != nil {
			log.Printf("Count query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Build the base query
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
        WHERE 1=1
    `

		// Add filters based on query parameters
		var args []interface{}
		argCounter := 1

		if state := c.Query("state"); state != "" {
			query += fmt.Sprintf(" AND state = $%d", argCounter)
			args = append(args, state)
			argCounter++
		}

		// Add pagination
		query += fmt.Sprintf(" ORDER BY id LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
		args = append(args, pageSize, offset)

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

		c.JSON(http.StatusOK, gin.H{
			"data":      rioters,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     int(math.Ceil(float64(total) / float64(pageSize))),
		})
	})
	router.GET("/api/search/suggestions", func(c *gin.Context) {
		term := c.Query("term")
		if term == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search term is required"})
			return
		}

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
		LIMIT 5
	`

		rows, err := db.Query(query, term)
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

		c.JSON(http.StatusOK, suggestions)
	})

	router.GET("/api/rioters/count-by-state", func(c *gin.Context) {
		rows, err := db.Query("SELECT location, COUNT(*) FROM rioters GROUP BY location")
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

		c.JSON(http.StatusOK, stateCounts)
	})
	router.GET("/api/rioters/nearby", func(c *gin.Context) {
		latStr := c.Query("lat")
		lngStr := c.Query("lng")
		radiusStr := c.DefaultQuery("radius", "50000") // Default: 50km

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

		gridSize := fmt.Sprintf("%f", 0.01/zoom) // ✅ Now zoom is correctly converted

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

	// Start server
	router.Run(":8080")
}
