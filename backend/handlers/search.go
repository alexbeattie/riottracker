package handlers

import (
    "database/sql"
    "net/http"
    "strings"
    "encoding/json"
		"rioters/backend/models" // Add this import
)

type SearchFilters struct {
    SearchText   string                 `json:"searchText"`
    State        string                 `json:"state"`
		City         string                 `json:"city"`
    Charges      string                 `json:"charges"`
    Status       string                 `json:"status"`
    Affiliations map[string]bool        `json:"affiliations"`
}

func SearchRioters(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse query parameters
        search := r.URL.Query().Get("search")
        state := r.URL.Query().Get("state")
        charges := r.URL.Query().Get("charges")
        status := r.URL.Query().Get("status")
        affiliations := r.URL.Query()["affiliations"]

        query := `
            SELECT id, first_name, last_name, city, state, 
                   ts_headline(first_name || ' ' || last_name, 
                               websearch_to_tsquery('english', $1)) as highlight,
                   ts_rank(search_vector, websearch_to_tsquery('english', $1)) as rank
            FROM rioters
            WHERE (
                $1 = '' OR 
                search_vector @@ websearch_to_tsquery('english', $1)
            )
            AND ($2 = '' OR state ILIKE $2)
            AND ($3 = '' OR charges = $3)
            AND ($4 = '' OR case_status = $4)
            AND ($5::boolean[] IS NULL OR military_le = ANY($5))
            ORDER BY rank DESC
            LIMIT 50
        `

        // Convert affiliations to boolean array
        var affilArray pq.BoolArray
        for _, a := range affiliations {
            affilArray = append(affilArray, a == "military_le" || a == "extremist")
        }

        rows, err := db.Query(query, 
            search,
            state,
            charges,
            status,
            affilArray,
        )
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var rioters []models.Rioter
        for rows.Next() {
            var r models.Rioter
            if err := rows.Scan(/* add all your fields here */); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            rioters = append(rioters, r)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(rioters)
    }
		
}
