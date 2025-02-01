// models/rioter.go
package models

type Rioter struct {
    ID            int     `json:"id"`
    FirstName     string  `json:"first_name"`
    LastName      string  `json:"last_name"`
    City          string  `json:"city"`
    State         string  `json:"state"`
    Summary       string  `json:"summary"`
    Charges       string  `json:"charges"`
    CaseStatus    string  `json:"case_status"`
    CaseUpdates   string  `json:"case_updates"`
    PhotoName     string  `json:"photo_name"`
    Latitude      float64 `json:"latitude"`
    Longitude     float64 `json:"longitude"`
    ViolenceAssault bool  `json:"violence_assault"`
    Conspiracy    bool    `json:"conspiracy"`
    Property      bool    `json:"property"`
    MilitaryLE    bool    `json:"military_le"`
    Extremist     bool    `json:"extremist"`
    Sentenced     bool    `json:"sentenced"`
    Commuted      bool    `json:"commuted"`
}