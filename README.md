# January 6 Rioters Tracker

A web-based mapping application that tracks and visualizes individuals involved in the January 6, 2021 U.S. Capitol riots.

## 🚀 Quick Start

**For detailed startup instructions, see [START_HERE.md](./START_HERE.md)**

### Quick Commands

**Start Backend:**
```bash
cd backend
go run cmd/migrate/main.go
```

**Start Frontend:**
```bash
cd frontend
npm run serve
```

The application will be available at:
- Frontend: http://localhost:8080
- Backend API: http://localhost:8085

---

## Prerequisites

- **PostgreSQL** (with PostGIS extension)
- **Redis**
- **Go** 1.23.3+
- **Node.js** and npm/yarn
- **Mapbox Access Token** ([Get one here](https://account.mapbox.com/))

---

## Project Structure

```
riottracker/
├── backend/          # Go API server
│   ├── cmd/
│   │   └── migrate/  # Main application entry point
│   ├── handlers/    # HTTP handlers
│   ├── models/      # Data models
│   └── pkg/         # Packages (geocoding, etc.)
├── frontend/        # Vue.js frontend
│   ├── src/
│   │   ├── components/  # Vue components
│   │   ├── views/       # Route views
│   │   └── router/       # Vue Router config
└── SPECS.md         # Detailed project specifications
```

---

## Setup

### 1. Database Setup

Create PostgreSQL database and enable PostGIS:
```bash
createdb riotracker
psql riotracker -c "CREATE EXTENSION postgis;"
```

### 2. Backend Setup

```bash
cd backend
go mod tidy
```

Create `backend/.env`:
```env
DATABASE_URL=postgresql://user:password@localhost:5432/riottracker?sslmode=disable
MAPBOX_ACCESS_TOKEN=your_token_here
PORT=8085
REDIS_ADDR=localhost:6379
```

### 3. Frontend Setup

```bash
cd frontend
npm install
```

Create `frontend/.env`:
```env
VUE_APP_MAPBOX_ACCESS_TOKEN=your_token_here
VUE_APP_API_BASE_URL=http://localhost:8085/api
```

### 4. Start Redis

```bash
# macOS
brew services start redis

# Linux
sudo systemctl start redis
```

---

## Running the Application

### Development Mode

**Terminal 1 - Backend:**
```bash
cd backend
go run cmd/migrate/main.go
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run serve
```

### Production Build

**Frontend:**
```bash
cd frontend
npm run build
```

**Backend:**
```bash
cd backend
go build -o riotracker-server cmd/migrate/main.go
./riottracker-server
```

---

## Documentation

- **[START_HERE.md](./START_HERE.md)** - Detailed startup guide
- **[SPECS.md](./SPECS.md)** - Complete project specifications and API documentation

---

## Features

- 🗺️ Interactive map visualization with Mapbox
- 🔍 Advanced search and filtering
- 📍 Geographic clustering and nearby search
- 📸 Photo management
- 📊 State-based statistics
- 🔄 Real-time search suggestions
- 📱 Mobile-responsive design

---

## Tech Stack

**Backend:**
- Go 1.23.3
- Gin Web Framework
- PostgreSQL + PostGIS
- Redis (caching)

**Frontend:**
- Vue.js 3
- Mapbox GL JS
- Tailwind CSS
- Pinia (state management)

---

## API Endpoints

- `GET /api/health` - Health check
- `GET /api/rioters` - List rioters (with filtering)
- `GET /api/rioters/:id` - Get single rioter
- `POST /api/rioters` - Create rioter
- `PUT /api/rioters/:id` - Update rioter
- `GET /api/rioters/nearby` - Find nearby rioters
- `GET /api/rioters/count-by-state` - State statistics
- `POST /api/rioters/upload-photo` - Upload photo

See [SPECS.md](./SPECS.md) for complete API documentation.

---

## Troubleshooting

See [START_HERE.md](./START_HERE.md#troubleshooting) for common issues and solutions.

---

## License

[Add your license here]
