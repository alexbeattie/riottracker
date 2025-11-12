package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"rioters/backend/pkg/geocode"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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
}

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
	PhotoName *string  `json:"photo_name"`
	Longitude *float64 `json:"longitude"`
	Latitude  *float64 `json:"latitude"`
	Distance  float64  `json:"distance"`
}

func (nb *NullBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nb.Valid = false
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return errors.New("NullBool: unable to unmarshal data into bool")
	}
	nb.Bool = b
	nb.Valid = true
	return nil
}

func (nb NullBool) MarshalJSON() ([]byte, error) {
	if !nb.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nb.Bool)
}

var (
	ctx = context.Background()
	rdb *redis.Client
)

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

// Helper function to build cache keys for rioter queries
func buildCacheKey(c *gin.Context) string {
	params := []string{
		c.Query("state"),
		c.Query("charges"),
		c.Query("status"),
		c.Query("military_le"),
		c.Query("extremist"),
		c.Query("sentenced"),
		c.Query("page"),
		c.Query("page_size"),
	}
	return "rioters:" + strings.Join(params, "|")
}

var log = logrus.New()

func main() {
	log.SetFormatter(&logrus.JSONFormatter{})

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Using existing environment variables.")
	}
	connectionString := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Database ping error:", err)
	}

	if err := geocode.BatchUpdateCoordinates(db); err != nil {
		log.Fatal("Coordinate migration failed:", err)
	}
	if err := geocode.UpdateAllGeometries(db); err != nil {
		log.Fatal("Geometry update failed:", err)
	}
	log.Println("Coordinate migration completed successfully")

	router := gin.Default()
	router.Static("/photos", "./photos")
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/photos/") {
			requestedPath := filepath.Join("./photos", strings.TrimPrefix(c.Request.URL.Path, "/photos/"))
			c.Header("Cache-Control", "public, max-age=86400")
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

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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

	router.GET("/api/rioters/:id", func(c *gin.Context) {
		id := c.Param("id")
		cacheKey := fmt.Sprintf("rioter:%s", id)
		cached, err := getCachedData(cacheKey)
		if err == nil && cached != "" {
			log.Printf("Cache hit for rioter ID: %s", id)
			c.JSON(http.StatusOK, json.RawMessage(cached))
			return
		}

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
		err = db.QueryRow(query, id).Scan(
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

		jsonData, _ := json.Marshal(rioter)
		if err := setCachedData(cacheKey, string(jsonData), 10*time.Minute); err != nil {
			log.Printf("Failed to cache rioter %s: %v", id, err)
		}
		c.JSON(http.StatusOK, rioter)
	})

	router.GET("/api/search/suggestions", func(c *gin.Context) {
		term := c.Query("term")
		if term == "" {
			c.JSON(http.StatusOK, []string{})
			return
		}

		// Hard-coded suggestions for testing
		suggestions := []string{
			"Charles Adams",
			"Charles Baker",
			"Charles Smith",
			"Charles Walters",
			"Daniel Adams",
			"Daniel Miller",
			"Daniel Smith",
			"Rasha Abual-Ragheb",
		}

		// Filter suggestions based on the search term
		var filtered []string
		for _, suggestion := range suggestions {
			if strings.Contains(strings.ToLower(suggestion), strings.ToLower(term)) {
				filtered = append(filtered, suggestion)
			}
		}

		c.JSON(http.StatusOK, filtered)
	})

	router.GET("/api/rioters/count-by-state", func(c *gin.Context) {
		cacheKey := "count_by_state"
		cached, err := getCachedData(cacheKey)
		if err == nil && cached != "" {
			c.JSON(http.StatusOK, json.RawMessage(cached))
			return
		}
		rows, err := db.Query("SELECT state, COUNT(*) FROM rioters GROUP BY state")
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
		jsonData, _ := json.Marshal(stateCounts)
		setCachedData(cacheKey, string(jsonData), 10*time.Minute)
		c.JSON(http.StatusOK, stateCounts)
	})

	router.GET("/api/rioters/nearby", func(c *gin.Context) {
		latStr, lngStr := c.Query("lat"), c.Query("lng")
		radiusStr := c.DefaultQuery("radius", "50000")

		// Get filter parameters first to include in cache key
		state := c.Query("state")
		status := c.Query("status")
		searchText := c.Query("searchText")
		militaryLE := c.Query("military_le") == "true"
		extremist := c.Query("extremist") == "true"
		sentenced := c.Query("sentenced") == "true"
		commuted := c.Query("commuted") == "true"

		// Build cache key that includes all filter parameters
		cacheKey := fmt.Sprintf("nearby_%s_%s_%s_state:%s_status:%s_search:%s_mil:%t_ext:%t_sent:%t_comm:%t",
			latStr, lngStr, radiusStr, state, status, searchText, militaryLE, extremist, sentenced, commuted)
		cached, err := getCachedData(cacheKey)
		if err == nil && cached != "" {
			c.JSON(http.StatusOK, json.RawMessage(cached))
			return
		}

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

		log.Printf("Received request: lat=%f, lng=%f, radius=%f, filters: state=%s, status=%s, searchText=%s", lat, lng, radius, state, status, searchText)

		// Build WHERE clause with filters
		whereClause := "ST_DWithin(geom::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $3)"
		argCounter := 4
		args := []interface{}{lng, lat, radius}

		if state != "" {
			whereClause += fmt.Sprintf(" AND state = $%d", argCounter)
			args = append(args, state)
			argCounter++
		}

		if status != "" {
			whereClause += fmt.Sprintf(" AND case_status = $%d", argCounter)
			args = append(args, status)
			argCounter++
		}

		if searchText != "" {
			whereClause += fmt.Sprintf(" AND (search_vector @@ websearch_to_tsquery('english', $%d) OR first_name ILIKE $%d OR last_name ILIKE $%d)", argCounter, argCounter+1, argCounter+2)
			args = append(args, searchText, "%"+searchText+"%", "%"+searchText+"%")
			argCounter += 3
		}

		// Handle affiliation filters
		if militaryLE {
			whereClause += " AND military_le = true"
		}
		if extremist {
			whereClause += " AND extremist = true"
		}
		if sentenced {
			whereClause += " AND sentenced = true"
		}
		if commuted {
			whereClause += " AND commuted = true"
		}

		query := fmt.Sprintf(`
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
            WHERE %s
            ORDER BY distance;
        `, whereClause)
		log.Printf("Executing query: %s", query)

		rows, err := db.Query(query, args...)
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
				&r.PhotoName,
				&r.Longitude,
				&r.Latitude,
				&r.Distance,
			); err != nil {
				log.Printf("Row scan error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			nearbyRioters = append(nearbyRioters, r)
		}
		log.Printf("Successfully fetched %d nearby rioters", len(nearbyRioters))
		jsonData, _ := json.Marshal(nearbyRioters)
		setCachedData(cacheKey, string(jsonData), 10*time.Minute)
		c.JSON(http.StatusOK, nearbyRioters)
	})

	router.GET("/api/rioters/clusters", func(c *gin.Context) {
		zoomStr := c.DefaultQuery("zoom", "10")
		zoom, err := strconv.ParseFloat(zoomStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid zoom value"})
			return
		}

		gridSize := 0.01 / math.Pow(2, zoom)
		log.Printf("Zoom Level: %f, Grid Size: %f", zoom, gridSize)

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
			if err := rows.Scan(&center, &points, &pointCount); err != nil { // Fixed typo: "¢er" -> "&center"
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

	router.POST("/api/rioters/upload-photo", func(c *gin.Context) {
		idStr := c.PostForm("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		file, err := c.FormFile("photo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No photo file provided"})
			return
		}

		var rioter Rioter
		err = db.QueryRow("SELECT last_name, first_name, middle_name FROM rioters WHERE id = $1", id).Scan(
			&rioter.LastName, &rioter.FirstName, &rioter.MiddleName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rioter info"})
			return
		}

		photoName := strings.ToLower(rioter.LastName) + "-" + strings.ToLower(rioter.FirstName)
		if rioter.MiddleName != nil && *rioter.MiddleName != "" {
			photoName += "-" + strings.ToLower(*rioter.MiddleName)
		}
		photoName += ".jpg"
		const maxUploadSize = 10 << 20 // 10 MB
		allowedMimeTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
		}

		file, err = c.FormFile("photo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No photo file provided"})
			return
		}

		if file.Size > maxUploadSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10 MB"})
			return
		}

		if !allowedMimeTypes[file.Header.Get("Content-Type")] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPEG and PNG files are allowed"})
			return
		}

		dst := filepath.Join("./photos", photoName)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
			return
		}

		_, err = db.Exec("UPDATE rioters SET photo_name = $1 WHERE id = $2", photoName, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rioter photo"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     "success",
			"photo_name": photoName,
		})
	})

	router.POST("/api/rioters", func(c *gin.Context) {
		var newRioter NewRioterRequest
		if err := c.ShouldBindJSON(&newRioter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		city := ""
		if newRioter.City != nil {
			city = *newRioter.City
		}
		state := ""
		if newRioter.State != nil {
			state = *newRioter.State
		}

		photoName := strings.ToLower(newRioter.LastName) + "-" + strings.ToLower(newRioter.FirstName)
		if newRioter.MiddleName != nil && *newRioter.MiddleName != "" {
			photoName += "-" + strings.ToLower(*newRioter.MiddleName)
		}
		photoName += ".jpg"

		lat, lon, err := geocode.GeocodeWithMapbox(city, state)
		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Geocoding error: " + err.Error()})
			return
		}

		log.Printf("Geocoded coordinates for %s, %s: lat=%v, lon=%v", city, state, *lat, *lon)

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

	router.PUT("/api/rioters/:id", func(c *gin.Context) {
		id := c.Param("id")
		var updatedRioter Rioter
		if err := c.ShouldBindJSON(&updatedRioter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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
			updatedRioter.Sentenced, updatedRioter.Commuted, updatedRioter.Pardoned, updatedRioter.ArrestDate, updatedRioter.PhotoName, updatedRioter.Latitude, updatedRioter.Longitude, id,
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

	router.GET("/api/rioters", func(c *gin.Context) {
		cacheKey := buildCacheKey(c)

		// Skip cache if we're doing an exact name search
		useCache := c.Query("search_exact") != "true"

		if useCache {
			cached, err := getCachedData(cacheKey)
			if err == nil && cached != "" {
				log.Println("Cache hit for rioters list")
				c.JSON(http.StatusOK, json.RawMessage(cached))
				return
			}
		}

		whereClause := "WHERE 1=1"
		var args []interface{}
		argCounter := 1

		// Check for exact name search first
		searchExact := c.Query("search_exact") == "true"
		firstName := c.Query("first_name")
		lastName := c.Query("last_name")

		if searchExact && firstName != "" && lastName != "" {
			// Use ILIKE for case-insensitive matching
			whereClause += fmt.Sprintf(" AND first_name ILIKE $%d AND last_name ILIKE $%d", argCounter, argCounter+1)
			args = append(args, firstName, lastName)
			argCounter += 2
		} else {
			// Regular search filters
			if state := c.Query("state"); state != "" {
				whereClause += fmt.Sprintf(" AND state = $%d", argCounter)
				args = append(args, state)
				argCounter++
			}

			// Handle charges - support multiple charge types (OR logic) - COMMENTED OUT - Not working correctly
			/*
			   chargeConditions := []string{}
			   if c.Query("violence_assault") == "true" {
			       chargeConditions = append(chargeConditions, fmt.Sprintf("violence_assault = true"))
			   }
			   if c.Query("conspiracy") == "true" {
			       chargeConditions = append(chargeConditions, fmt.Sprintf("conspiracy = true"))
			   }
			   if c.Query("property") == "true" {
			       chargeConditions = append(chargeConditions, fmt.Sprintf("property = true"))
			   }
			   // Legacy support: if "charges" query param is used, treat it as a single charge type
			   if len(chargeConditions) == 0 {
			       if charges := c.Query("charges"); charges != "" {
			           switch charges {
			           case "violence_assault":
			               chargeConditions = append(chargeConditions, "violence_assault = true")
			           case "conspiracy":
			               chargeConditions = append(chargeConditions, "conspiracy = true")
			           case "property":
			               chargeConditions = append(chargeConditions, "property = true")
			           default:
			               // Fallback to text search if it's not a recognized charge type
			               whereClause += fmt.Sprintf(" AND charges ILIKE $%d", argCounter)
			               args = append(args, "%"+charges+"%")
			               argCounter++
			           }
			       }
			   }
			   // If we have charge conditions, combine them with OR (rioter has ANY of the selected charges)
			   if len(chargeConditions) > 0 {
			       whereClause += fmt.Sprintf(" AND (%s)", strings.Join(chargeConditions, " OR "))
			   }
			*/

			if status := c.Query("status"); status != "" {
				whereClause += fmt.Sprintf(" AND case_status = $%d", argCounter)
				args = append(args, status)
				argCounter++
			}
		}

		// Handle affiliations
		affMapping := map[string]string{
			"military_le": "military_le",
			"extremist":   "extremist",
			"sentenced":   "sentenced",
		}
		for param, column := range affMapping {
			if c.Query(param) == "true" {
				whereClause += fmt.Sprintf(" AND %s = true", column)
			}
		}

		// If we're doing a regular search (not exact name search), also check searchText
		if !searchExact {
			if searchText := c.Query("searchText"); searchText != "" {
				whereClause += fmt.Sprintf(" AND (search_vector @@ websearch_to_tsquery('english', $%d) OR first_name ILIKE $%d OR last_name ILIKE $%d)",
					argCounter, argCounter+1, argCounter+2)
				args = append(args, searchText, "%"+searchText+"%", "%"+searchText+"%")
				argCounter += 3
			}
		}

		var total int
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM rioters %s", whereClause)
		if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
			log.Printf("Count query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

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
            ORDER BY id
        `, whereClause)

		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}
		pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "50"))
		if err != nil || pageSize < 1 || pageSize > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size (1-1000)"})
			return
		}
		sortBy := c.DefaultQuery("sort_by", "id")
		sortOrder := c.DefaultQuery("sort_order", "asc")

		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "asc"
		}

		query = fmt.Sprintf(`
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
       ORDER BY %s %s
   `, whereClause, sortBy, sortOrder)

		offset := (page - 1) * pageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
		args = append(args, pageSize, offset)

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

		response := gin.H{
			"data":      rioters,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     int(math.Ceil(float64(total) / float64(pageSize))),
		}

		jsonData, _ := json.Marshal(response)
		if err := setCachedData(cacheKey, string(jsonData), 10*time.Minute); err != nil {
			log.Printf("Failed to cache data: %v", err)
		}
		c.JSON(http.StatusOK, response)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	router.Run(":" + port)
}
