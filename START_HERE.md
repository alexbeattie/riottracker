# 🚀 Quick Start Guide

## Prerequisites

Before starting, make sure you have:

1. **PostgreSQL** (with PostGIS extension) installed and running
2. **Redis** installed and running
3. **Go** 1.23.3+ installed
4. **Node.js** and npm/yarn installed
5. **Mapbox Access Token** (get one at https://account.mapbox.com/)

---

## Step 1: Set Up Database

1. Create a PostgreSQL database:
   ```bash
   createdb riotracker
   ```

2. Enable PostGIS extension:
   ```bash
   psql riotracker -c "CREATE EXTENSION postgis;"
   ```

3. Create the `rioters` table (you'll need to run your migration script or create the table manually)

---

## Step 2: Configure Environment Variables

### Backend Setup

1. Navigate to backend directory:
   ```bash
   cd backend
   ```

2. Create a `.env` file (or set environment variables):
   ```bash
   # backend/.env
   DATABASE_URL=postgresql://username:password@localhost:5432/riottracker?sslmode=disable
   MAPBOX_ACCESS_TOKEN=your_mapbox_token_here
   PORT=8085
   REDIS_ADDR=localhost:6379
   REDIS_PASSWORD=
   REDIS_DB=0
   ```

   **Note:** The frontend expects the backend on port `8085`, so make sure `PORT=8085` in your `.env` file.

### Frontend Setup

1. Navigate to frontend directory:
   ```bash
   cd frontend
   ```

2. Create a `.env` file:
   ```bash
   # frontend/.env
   VUE_APP_MAPBOX_ACCESS_TOKEN=your_mapbox_token_here
   VUE_APP_API_BASE_URL=http://localhost:8085/api
   ```

---

## Step 3: Install Dependencies

### Backend
```bash
cd backend
go mod tidy
```

### Frontend
```bash
cd frontend
npm install
# OR
yarn install
```

---

## Step 4: Start Redis

Make sure Redis is running:

```bash
# macOS (Homebrew)
brew services start redis

# Linux
sudo systemctl start redis

# Or run directly
redis-server
```

---

## Step 5: Start the Backend Server

```bash
cd backend
go run cmd/migrate/main.go
```

The backend will:
- Connect to Redis
- Connect to PostgreSQL
- Run geocoding migrations (if needed)
- Start the API server on port 8085 (or the port specified in PORT env var)

You should see:
```
Connected to Redis
Coordinate migration completed successfully
```

The server will be running at: **http://localhost:8085**

---

## Step 6: Start the Frontend Development Server

Open a **new terminal window** and run:

```bash
cd frontend
npm run serve
# OR
yarn serve
```

The frontend will start on: **http://localhost:8080** (or another port if 8080 is busy)

---

## Step 7: Access the Application

Open your browser and navigate to:
- **Frontend:** http://localhost:8080
- **Backend API:** http://localhost:8085/api/health (to test)

---

## Troubleshooting

### Backend won't start

1. **Database connection error:**
   - Check that PostgreSQL is running
   - Verify `DATABASE_URL` is correct
   - Ensure the database exists

2. **Redis connection error:**
   - Check that Redis is running: `redis-cli ping` (should return `PONG`)
   - Verify Redis address in `.env`

3. **Port already in use:**
   - Change `PORT` in backend `.env` file
   - Update frontend `VUE_APP_API_BASE_URL` to match

### Frontend won't connect to backend

1. **CORS errors:**
   - Backend CORS is configured to allow all origins (check `main.go`)

2. **API connection refused:**
   - Make sure backend is running on port 8085
   - Check `VUE_APP_API_BASE_URL` in frontend `.env`

3. **Mapbox errors:**
   - Verify `VUE_APP_MAPBOX_ACCESS_TOKEN` is set correctly
   - Check Mapbox token is valid at https://account.mapbox.com/

### Database issues

1. **PostGIS not found:**
   ```bash
   # Install PostGIS
   # macOS
   brew install postgis
   
   # Ubuntu/Debian
   sudo apt-get install postgis
   ```

2. **Missing table:**
   - Run your database migration script
   - Or create the table manually using the schema in SPECS.md

---

## Running Both Servers (Quick Reference)

### Terminal 1 - Backend:
```bash
cd backend
go run cmd/migrate/main.go
```

### Terminal 2 - Frontend:
```bash
cd frontend
npm run serve
```

---

## Production Build

### Build Frontend:
```bash
cd frontend
npm run build
# Output will be in frontend/dist/
```

### Build Backend:
```bash
cd backend
go build -o riotracker-server cmd/migrate/main.go
./riottracker-server
```

---

## Health Check

Test if everything is working:

1. **Backend health:** http://localhost:8085/api/health
2. **Get rioters:** http://localhost:8085/api/rioters?page=1&page_size=10
3. **Frontend:** http://localhost:8080

---

## Need Help?

- Check `SPECS.md` for detailed documentation
- Review error messages in terminal
- Check browser console for frontend errors
- Check backend logs (JSON formatted)

