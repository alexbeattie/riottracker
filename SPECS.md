# January 6 Rioters Tracker - Project Specifications

## 1. Project Overview

### 1.1 Purpose
A web-based mapping application that tracks and visualizes individuals involved in the January 6, 2021 U.S. Capitol riots. The application provides an interactive map interface with search and filtering capabilities to explore rioter data including charges, case status, affiliations, and geographic locations.

### 1.2 Key Features
- **Interactive Map Visualization**: Mapbox-powered map displaying rioter locations across the United States
- **Advanced Search & Filtering**: Search by name, location, charges, case status, and affiliations
- **Detailed Rioter Profiles**: Comprehensive information including charges, case updates, photos, and legal status
- **Geographic Clustering**: Visual clustering of rioters by location
- **CRUD Operations**: Create, read, update, and delete rioter records
- **Photo Management**: Upload and display rioter photos
- **Real-time Search Suggestions**: Autocomplete search functionality

---

## 2. Architecture

### 2.1 Technology Stack

#### Backend
- **Language**: Go 1.23.3
- **Web Framework**: Gin (gin-gonic/gin)
- **Database**: PostgreSQL with PostGIS extension (for geospatial data)
- **Caching**: Redis (go-redis/v9)
- **Geocoding**: Mapbox Geocoding API
- **Logging**: Logrus (sirupsen/logrus)
- **Environment**: godotenv

#### Frontend
- **Framework**: Vue.js 3.2.13
- **Router**: Vue Router 4
- **State Management**: Pinia 3.0.1
- **HTTP Client**: Axios with caching (axios-cache-interceptor)
- **Mapping**: Mapbox GL JS 3.9.4
- **Styling**: Tailwind CSS 3.3.5
- **Build Tool**: Vue CLI 5.0

### 2.2 System Architecture
```
┌─────────────┐
│   Browser   │
│  (Vue.js)   │
└──────┬──────┘
       │ HTTP/REST API
       │
┌──────▼──────┐         ┌──────────┐
│   Go API    │◄────────┤  Redis   │
│   Server    │  Cache  │  Cache   │
└──────┬──────┘         └──────────┘
       │
       │ SQL Queries
       │
┌──────▼──────┐
│ PostgreSQL  │
│  + PostGIS  │
└─────────────┘
```

---

## 3. Data Model

### 3.1 Rioter Entity

```go
type Rioter struct {
    ID              int       `json:"id"`
    FirstName       string    `json:"first_name"`
    LastName        string    `json:"last_name"`
    MiddleName      *string   `json:"middle_name"`
    City            *string   `json:"city"`
    State           *string   `json:"state"`
    Age             *int      `json:"age"`
    Summary         *string   `json:"summary"`
    Jurisdiction    *string   `json:"jurisdiction"`
    Charges         *string   `json:"charges"`
    ChargesLink     *string   `json:"charges_link"`
    CaseStatus      *string   `json:"case_status"`
    CaseUpdates     *string   `json:"case_updates"`
    PhotoName       *string   `json:"photo_name"`
    Latitude        float64   `json:"latitude"`
    Longitude       float64   `json:"longitude"`
    
    // Boolean flags
    ViolenceAssault bool      `json:"violence_assault"`
    Conspiracy      bool      `json:"conspiracy"`
    Theft           bool      `json:"theft"`
    Property        bool      `json:"property"`
    MilitaryLE      bool      `json:"military_le"`
    Extremist       bool      `json:"extremist"`
    InspiredTrump   bool      `json:"inspired_trump"`
    Sentenced       bool      `json:"sentenced"`
    Commuted        bool      `json:"commuted"`
    Pardoned        bool      `json:"pardoned"`
    
    // Dates
    ArrestDate      *time.Time `json:"arrest_date"`
    
    // Geospatial (PostGIS)
    Geom            geometry.Point `json:"-"` // Internal use only
}
```

### 3.2 Database Schema (PostgreSQL)

**Table: `rioters`**
- `id` (SERIAL PRIMARY KEY)
- `first_name` (VARCHAR)
- `last_name` (VARCHAR)
- `middle_name` (VARCHAR, nullable)
- `city` (VARCHAR, nullable)
- `state` (VARCHAR, nullable)
- `age` (INTEGER, nullable)
- `summary` (TEXT, nullable)
- `jurisdiction` (VARCHAR, nullable)
- `charges` (TEXT, nullable)
- `charges_link` (VARCHAR, nullable)
- `case_status` (VARCHAR, nullable)
- `case_updates` (TEXT, nullable)
- `photo_name` (VARCHAR, nullable)
- `latitude` (DOUBLE PRECISION)
- `longitude` (DOUBLE PRECISION)
- `geom` (GEOGRAPHY(POINT, 4326)) - PostGIS geometry column
- `violence_assault` (BOOLEAN)
- `conspiracy` (BOOLEAN)
- `theft` (BOOLEAN)
- `property` (BOOLEAN)
- `military_le` (BOOLEAN)
- `extremist` (BOOLEAN)
- `inspired_trump` (BOOLEAN)
- `sentenced` (BOOLEAN)
- `commuted` (BOOLEAN)
- `pardoned` (BOOLEAN)
- `arrest_date` (DATE, nullable)
- `search_vector` (TSVECTOR) - Full-text search index

**Indexes:**
- Primary key on `id`
- GIST index on `geom` for spatial queries
- GIN index on `search_vector` for full-text search
- Indexes on `state`, `case_status`, `charges` for filtering

---

## 4. API Endpoints

### 4.1 Health Check
- **GET** `/api/health`
  - Returns API health status
  - Response: `{"status": "OK"}` or `{"status": "DOWN"}`

### 4.2 Rioter CRUD Operations

#### Get All Rioters (with filtering)
- **GET** `/api/rioters`
  - **Query Parameters:**
    - `page` (int, default: 1) - Page number
    - `page_size` (int, default: 10, max: 100) - Results per page
    - `search` (string) - Full-text search query
    - `state` (string) - Filter by state
    - `city` (string) - Filter by city
    - `charges` (string) - Filter by charge type
    - `status` (string) - Filter by case status
    - `military_le` (bool) - Filter by military/LE affiliation
    - `extremist` (bool) - Filter by extremist affiliation
    - `sentenced` (bool) - Filter by sentenced status
    - `commuted` (bool) - Filter by commuted status
    - `search_exact` (bool) - Exact name search mode
    - `first_name` (string) - First name for exact search
    - `last_name` (string) - Last name for exact search
  - **Response:**
    ```json
    {
      "total": 1567,
      "page": 1,
      "page_size": 50,
      "data": [/* array of Rioter objects */]
    }
    ```
  - **Caching:** Redis cache with 10-minute TTL

#### Get Single Rioter
- **GET** `/api/rioters/:id`
  - Returns complete rioter details including geospatial data
  - **Response:** Single Rioter object
  - **Caching:** Redis cache with 10-minute TTL

#### Create Rioter
- **POST** `/api/rioters`
  - **Request Body:** `NewRioterRequest` JSON
  - Automatically geocodes city/state using Mapbox
  - Creates PostGIS geometry point
  - **Response:** `{"status": "success", "id": <new_id>}`

#### Update Rioter
- **PUT** `/api/rioters/:id`
  - **Request Body:** Complete Rioter object
  - Updates geospatial data if coordinates change
  - **Response:** `{"status": "success", "id": <id>}`

#### Delete Rioter
- **DELETE** `/api/rioters/:id`
  - **Response:** `{"status": "success", "id": <id>}`

### 4.3 Search & Filtering

#### Search Suggestions
- **GET** `/api/search/suggestions`
  - **Query Parameters:**
    - `term` (string) - Search term
  - **Response:** Array of suggestion strings
  - **Note:** Currently returns hardcoded suggestions (needs implementation)

#### Count by State
- **GET** `/api/rioters/count-by-state`
  - Returns count of rioters grouped by state
  - **Response:** `{"Alabama": 45, "California": 123, ..., "total": 1567}`
  - **Caching:** Redis cache with 10-minute TTL

### 4.4 Geographic Queries

#### Nearby Rioters
- **GET** `/api/rioters/nearby`
  - **Query Parameters:**
    - `lat` (float, required) - Latitude
    - `lng` (float, required) - Longitude
    - `radius` (float, default: 50000) - Radius in meters (max: 1,000,000)
  - Returns rioters within specified radius, ordered by distance
  - Uses PostGIS `ST_DWithin` for efficient spatial queries
  - **Response:** Array of `NearbyRioter` objects with distance field
  - **Caching:** Redis cache with 10-minute TTL

#### Clusters
- **GET** `/api/rioters/clusters`
  - **Query Parameters:**
    - `zoom` (float, default: 10) - Map zoom level
  - Returns clustered rioter data for map visualization
  - Uses PostGIS grid-based clustering
  - **Response:** Array of cluster objects with center, points, and count

### 4.5 Photo Management

#### Upload Photo
- **POST** `/api/rioters/upload-photo`
  - **Form Data:**
    - `id` (int, required) - Rioter ID
    - `photo` (file, required) - Image file (JPEG/PNG, max 10MB)
  - Automatically names photo: `{lastname}-{firstname}-{middlename}.jpg`
  - **Response:** `{"status": "success", "photo_name": "<filename>"}`

#### Serve Photos
- **GET** `/photos/:filename`
  - Serves rioter photos
  - Falls back to `placeholder.jpg` if photo not found
  - **Cache Headers:** `Cache-Control: public, max-age=86400`

---

## 5. Frontend Components

### 5.1 Views (Routes)

#### HomeView (`/`)
- Main landing page with map and search interface

#### MapView (`/map`)
- Full-screen map view with sidebar filters

#### NewRioterView (`/new`)
- Form to create new rioter entries

#### EditRioterView (`/rioter/:id/edit`)
- Form to edit existing rioter records

### 5.2 Core Components

#### `App.vue`
- Main application container
- Manages global state (rioters list, selected rioter, filters)
- Handles API calls and data fetching
- Mobile-responsive sidebar with search filters
- Flyout detail panel for selected rioter

#### `RiotersMap.vue`
- Mapbox GL map component
- Displays rioter markers with clustering
- Handles marker clicks and popups
- Supports fly-to-marker animation
- Jitter algorithm for overlapping markers

#### `SearchFilters.vue`
- Search input with autocomplete suggestions
- State dropdown filter
- Charges dropdown filter
- Case status dropdown filter
- Affiliation toggle buttons (Military/LEO, Extremist, Sentenced)
- State count display
- Reset filters functionality

#### `RiotersList.vue`
- Displays paginated list of rioters
- Shows rioter photos, names, locations
- Highlights selected rioter
- Click handler to select rioter and show details

#### `RioterForm.vue` / `NewRioterForm.vue` / `EditRioterForm.vue`
- Form components for creating/editing rioter data
- Handles all rioter fields
- Photo upload capability
- Form validation

#### `RioterImage.vue`
- Reusable component for displaying rioter photos
- Handles image loading errors with placeholder fallback

#### `BasePagination.vue`
- Pagination controls for rioter lists

#### `Navigation.vue`
- Top navigation bar with links to different views

### 5.3 State Management

#### `stores/rioters.js` (Pinia Store)
- Centralized state for rioter data
- Actions for fetching, creating, updating rioters
- Getters for filtered/computed rioter lists

### 5.4 Utilities

#### `api.js`
- Axios instance configured with base URL
- Request/response interceptors
- Error handling

#### `imageHandling.js`
- Helper functions for image URL generation
- Placeholder image handling

---

## 6. Key Features & Functionality

### 6.1 Map Features
- **Interactive Map**: Mapbox-powered map with zoom, pan, and navigation controls
- **Marker Clustering**: Visual clustering of rioters at same/similar locations
- **Marker Jittering**: Slight offset for overlapping markers to improve visibility
- **Popup Information**: Click markers to see rioter name, photo, and location
- **Fly-to Animation**: Smooth map animation when selecting a rioter
- **Bounds Fitting**: Automatically fits map bounds to show all filtered rioters

### 6.2 Search & Filter Features
- **Full-Text Search**: PostgreSQL full-text search across names and locations
- **Autocomplete Suggestions**: Real-time search suggestions as user types
- **Multi-Filter Support**: Combine multiple filters (state, charges, status, affiliations)
- **State Count Display**: Shows count of rioters in selected state
- **Reset Filters**: One-click reset to clear all filters

### 6.3 Data Display Features
- **Rioter Detail Panel**: Slide-out panel showing complete rioter information
- **Photo Display**: Profile photos with fallback to placeholder
- **Tag System**: Visual tags for affiliations and charge types
- **Case Status Tracking**: Display current legal status
- **Charges Link**: External link to source documentation

### 6.4 Performance Features
- **Redis Caching**: API responses cached for 10 minutes
- **Pagination**: Efficient pagination for large result sets
- **Lazy Loading**: Images loaded on demand
- **Debounced Search**: Reduces API calls during typing

---

## 7. Environment Configuration

### 7.1 Backend Environment Variables
```env
DATABASE_URL=postgresql://user:password@localhost:5432/riottracker?sslmode=disable
MAPBOX_ACCESS_TOKEN=your_mapbox_token_here
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### 7.2 Frontend Environment Variables
```env
VUE_APP_MAPBOX_ACCESS_TOKEN=your_mapbox_token_here
VUE_APP_API_BASE_URL=http://localhost:8085/api
```

---

## 8. Database Setup

### 8.1 PostgreSQL Requirements
- PostgreSQL 12+ with PostGIS extension
- Enable PostGIS: `CREATE EXTENSION postgis;`
- Full-text search support (built-in)

### 8.2 Migration Scripts
- Coordinate migration: Geocodes existing records and creates PostGIS geometries
- Batch geocoding: Updates all records with missing coordinates

---

## 9. Deployment Considerations

### 9.1 Backend
- Port: 8085 (configurable)
- Static file serving: `/photos` directory
- CORS: Configured for cross-origin requests
- Health check endpoint for monitoring

### 9.2 Frontend
- Build command: `npm run build`
- Output: `dist/` directory
- Serve static files via web server (nginx, Apache, etc.)

### 9.3 Infrastructure Requirements
- PostgreSQL database with PostGIS
- Redis server for caching
- Mapbox account and API token
- Web server for static file serving
- SSL certificate for HTTPS (recommended)

---

## 10. Security Considerations

### 10.1 Current Implementation
- CORS configured (currently allows all origins - should be restricted in production)
- File upload size limits (10MB max)
- MIME type validation for uploads (JPEG/PNG only)
- SQL injection prevention via parameterized queries

### 10.2 Recommended Enhancements
- Authentication/authorization for admin operations
- Rate limiting on API endpoints
- Input validation and sanitization
- HTTPS enforcement
- API key authentication for Mapbox
- Database connection pooling limits
- Audit logging for data modifications

---

## 11. Known Issues & Limitations

### 11.1 Current Limitations
- Search suggestions endpoint returns hardcoded data (needs database implementation)
- No user authentication system
- No audit trail for data changes
- No bulk import functionality
- Limited error handling in some edge cases
- Mobile responsiveness could be improved

### 11.2 Technical Debt
- Some duplicate code between components
- Inconsistent error handling patterns
- Missing unit tests
- Missing integration tests
- No API documentation (Swagger/OpenAPI)

---

## 12. Future Enhancements

### 12.1 Planned Features
- [ ] User authentication and role-based access control
- ] Advanced analytics dashboard
- [ ] Export functionality (CSV, JSON, PDF)
- [ ] Bulk import from CSV/Excel
- [ ] Timeline visualization of arrests/sentences
- [ ] Network graph showing connections between rioters
- [ ] Advanced filtering (date ranges, charge severity)
- [ ] Data visualization charts (state distribution, charge types)
- [ ] Email notifications for case updates
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Unit and integration test coverage
- [ ] Docker containerization
- [ ] CI/CD pipeline

### 12.2 Performance Improvements
- [ ] Implement database connection pooling
- [ ] Add CDN for static assets
- [ ] Optimize map marker rendering for large datasets
- [ ] Implement virtual scrolling for rioter lists
- [ ] Add database query optimization and indexing review

### 12.3 UX Improvements
- [ ] Improved mobile navigation
- [ ] Keyboard shortcuts
- [ ] Advanced map controls (heatmap, density view)
- [ ] Print-friendly views
- [ ] Accessibility improvements (ARIA labels, keyboard navigation)
- [ ] Dark mode support

---

## 13. Development Workflow

### 13.1 Backend Development
```bash
cd backend
go mod tidy
go run cmd/migrate/main.go
```

### 13.2 Frontend Development
```bash
cd frontend
npm install
npm run serve
```

### 13.3 Database Migrations
- Run migration script: `go run cmd/migrate/main.go`
- This will geocode existing records and create PostGIS geometries

---

## 14. API Response Formats

### 4.1 Success Response
```json
{
  "status": "success",
  "data": { /* response data */ }
}
```

### 4.2 Error Response
```json
{
  "error": "Error message description"
}
```

### 4.3 Paginated Response
```json
{
  "total": 1567,
  "page": 1,
  "page_size": 50,
  "data": [ /* array of items */ ]
}
```

---

## 15. Code Organization

### 15.1 Backend Structure
```
backend/
├── cmd/
│   └── migrate/
│       └── main.go          # Main application entry point
├── handlers/
│   └── search.go            # Search handler (partial implementation)
├── models/
│   └── rioter.go            # Rioter data model
├── pkg/
│   └── geocode/
│       └── geocode.go       # Geocoding utilities
├── photos/                  # Photo storage directory
├── go.mod
└── go.sum
```

### 15.2 Frontend Structure
```
frontend/
├── public/
│   ├── index.html
│   └── marker-icons/        # Map marker icons
├── src/
│   ├── components/          # Vue components
│   ├── views/               # Route views
│   ├── router/              # Vue Router configuration
│   ├── stores/              # Pinia stores
│   ├── utils/               # Utility functions
│   ├── assets/              # Static assets
│   ├── api.js               # API client
│   ├── App.vue              # Root component
│   └── main.js              # Application entry
├── package.json
└── vue.config.js
```

---

## 16. Testing Strategy (Recommended)

### 16.1 Backend Tests
- Unit tests for geocoding functions
- Unit tests for database queries
- Integration tests for API endpoints
- Load testing for search endpoints

### 16.2 Frontend Tests
- Component unit tests (Vue Test Utils)
- E2E tests (Cypress/Playwright)
- Visual regression tests
- Accessibility tests

---

## 17. Monitoring & Logging

### 17.1 Current Logging
- Structured JSON logging via Logrus
- Request/response logging
- Error logging with context

### 17.2 Recommended Monitoring
- Application performance monitoring (APM)
- Database query performance tracking
- Cache hit/miss rates
- API endpoint response times
- Error rate tracking
- User activity analytics

---

## 18. Data Privacy & Compliance

### 18.1 Considerations
- Public data (court records, news articles)
- Ensure compliance with data retention policies
- Consider GDPR/privacy implications if deployed internationally
- Implement data export/deletion capabilities
- Document data sources and update frequency

---

## 19. Documentation Requirements

### 19.1 Code Documentation
- [ ] Add GoDoc comments to all exported functions
- [ ] Add JSDoc comments to Vue components
- [ ] Document API endpoints with examples
- [ ] Create architecture decision records (ADRs)

### 19.2 User Documentation
- [ ] User guide for searching and filtering
- [ ] Admin guide for data management
- [ ] API documentation for developers
- [ ] Deployment guide

---

## 20. Version History

### Current Version: 1.0.0 (Development)

**Features:**
- Basic CRUD operations
- Map visualization
- Search and filtering
- Photo management
- Geographic queries
- Redis caching

---

## Appendix A: Case Status Values

- `pleaded-not-guilty`
- `pleaded-guilty`
- `acquitted`
- `convicted`
- `dismissed`

## Appendix B: Charge Types

- `violence_assault` - Violence/Assault charges
- `conspiracy` - Conspiracy charges
- `property` - Property damage charges
- `theft` - Theft charges

## Appendix C: Affiliation Types

- `military_le` - Military or Law Enforcement background
- `extremist` - Extremist organization affiliation
- `sentenced` - Has been sentenced
- `commuted` - Sentence has been commuted
- `pardoned` - Has been pardoned
- `inspired_trump` - Inspired by Trump (boolean flag)

---

**Document Version:** 1.0  
**Last Updated:** 2024  
**Maintained By:** Development Team

