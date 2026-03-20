# Riottracker AI Enhancement Plan

## Current State

Full-stack Jan 6th Capitol rioter tracking application:
- **Backend**: Go (Gin), PostgreSQL + PostGIS, Redis
- **Frontend**: Vue.js 3, Mapbox GL JS, Tailwind CSS, Pinia
- **Data pipeline**: Wikipedia scraper → PostgreSQL, DOJ court link resolver
- **Search**: PostgreSQL full-text search (`tsvector`/`websearch_to_tsquery`)
- **~1,500+ rioter records** with charges, case status, location, affiliations

---

## Phase 1: RAG Chat Interface (Priority — Start Here)

**Goal**: Add a chat panel where users ask natural language questions about Jan 6 rioters and get sourced answers.

**Examples**:
- "How many pardoned rioters were from Texas?"
- "Who was sentenced to the longest prison term?"
- "Show me extremist-affiliated rioters who assaulted police"

### Architecture

```
User question
    │
    ▼
┌──────────────┐     ┌──────────────────┐
│  Vue.js Chat │────▶│  Go API endpoint │
│    Panel     │◀────│  POST /api/chat  │
└──────────────┘     └───────┬──────────┘
                             │
                    ┌────────▼─────────┐
                    │  Python RAG      │
                    │  service (FastAPI)│
                    └────────┬─────────┘
                             │
                 ┌───────────┼───────────┐
                 ▼           ▼           ▼
          ┌──────────┐ ┌──────────┐ ┌─────────┐
          │ ChromaDB │ │ Postgres │ │  Ollama  │
          │ (vectors)│ │ (SQL)    │ │  (LLM)  │
          └──────────┘ └──────────┘ └─────────┘
```

### Tasks

- [ ] **1a. Embed rioter data into ChromaDB**
  - Create `ai/ingest_rioters.py`
  - For each rioter, build a text document: name + city + state + charges + case_status + case_updates + affiliations
  - Embed with `nomic-embed-text` via Ollama (same as obsidian-rag)
  - Store in ChromaDB collection `riottracker`
  - Include metadata: `rioter_id`, `state`, `case_status`, `pardoned`, `sentenced`

- [ ] **1b. Build Python RAG query service**
  - Create `ai/query.py` — search ChromaDB, retrieve top-K rioter docs, pass to LLM
  - Create `ai/server.py` — FastAPI server with `POST /chat` endpoint
  - Use `mistral-nemo:12b` or `llama3.1:8b` via Ollama for generation
  - Prompt template grounds the LLM on retrieved rioter records
  - Return: answer text + list of cited rioter IDs

- [ ] **1c. Go API proxy endpoint**
  - Add `POST /api/chat` handler in Go backend
  - Proxies request to Python FastAPI service
  - Returns JSON: `{ answer: string, cited_rioters: int[] }`

- [ ] **1d. Vue.js Chat Panel component**
  - Create `frontend/src/components/ChatPanel.vue`
  - Slide-out panel or bottom drawer UI
  - Message input + streaming response display
  - When answer cites rioter IDs, render clickable names that fly-to-map + open popup
  - Chat history within session (Pinia store)

- [ ] **1e. Integration & polish**
  - Chat citations click → map flies to rioter marker + opens popup
  - Loading states, error handling
  - "Ask AI" button in the navigation bar

### Tech Decisions
- **Why ChromaDB over pgvector?** Keeps the AI layer decoupled; same stack as obsidian-rag for portfolio consistency. Can swap to pgvector later.
- **Why separate Python service?** Go doesn't have good Ollama/ChromaDB client libraries. FastAPI sidecar is clean and lets us reuse obsidian-rag patterns.
- **Model**: `llama3.1:8b` for speed, `mistral-nemo:12b` for quality. Can use model-router to pick.

---

## Phase 2: AI Case Summarizer

**Goal**: Generate plain-English summaries of each rioter's case from raw charges/notes.

### Tasks

- [ ] **2a. Batch summarization script**
  - Create `ai/summarize.py`
  - For each rioter, send charges + case_updates + notes to LLM
  - Prompt: "Summarize this person's Jan 6 case in 2-3 sentences for a general audience"
  - Store result in new `ai_summary` column in PostgreSQL

- [ ] **2b. Display summaries in frontend**
  - Show AI summary in rioter popup/card
  - Badge: "AI-generated summary"
  - Lazy-generate on first view if not yet cached

---

## Phase 3: Charge Severity Classifier

**Goal**: Classify each rioter's charges into severity tiers using NLP.

### Tiers
- **Low**: Trespassing, parading, disorderly conduct
- **Medium**: Obstruction, civil disorder, property destruction
- **High**: Assault on officers, conspiracy, weapons charges
- **Extreme**: Seditious conspiracy, assault with deadly weapon

### Tasks

- [ ] **3a. Build classifier**
  - Create `ai/classify_severity.py`
  - Heuristic-first approach (keyword matching), with LLM fallback for ambiguous cases
  - Store result in new `charge_severity` column (enum: low/medium/high/extreme)

- [ ] **3b. Color-coded map markers**
  - Green (low) → Yellow (medium) → Orange (high) → Red (extreme)
  - Legend overlay on map
  - Filter by severity in SearchFilters.vue

---

## Phase 4: Semantic Search Upgrade

**Goal**: Replace PostgreSQL full-text search with vector similarity search.

### Tasks

- [ ] **4a. Dual search path**
  - Keep existing `tsvector` search as fallback
  - Add semantic search endpoint: embed user query → ChromaDB similarity → return rioter IDs
  - Merge results (hybrid search)

- [ ] **4b. Frontend toggle**
  - "Smart Search" toggle in SearchFilters.vue
  - When enabled, queries go through semantic path
  - Show relevance scores

---

## Phase 5: Insights & Pattern Detection

**Goal**: Surface statistical patterns and anomalies.

### Tasks

- [ ] **5a. Insights API endpoint**
  - `GET /api/insights` — precomputed stats
  - Geographic clusters, sentencing disparities, pardon patterns
  - LLM-generated narrative for each insight

- [ ] **5b. Insights dashboard panel**
  - Cards showing key findings
  - "12 rioters from Maricopa County were all pardoned"
  - "Extremist-affiliated rioters received 3.2x longer sentences"

---

## File Structure (New)

```
riottracker/
├── ai/                      # NEW — AI/ML layer
│   ├── requirements.txt     # chromadb, ollama, fastapi, uvicorn
│   ├── ingest_rioters.py    # Embed rioter data → ChromaDB
│   ├── query.py             # RAG query logic
│   ├── server.py            # FastAPI service (POST /chat)
│   ├── summarize.py         # Batch case summarizer
│   └── classify_severity.py # Charge severity classifier
├── backend/                 # Existing Go backend
│   ├── handlers/
│   │   ├── search.go
│   │   └── chat.go          # NEW — proxy to Python AI service
│   └── ...
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── ChatPanel.vue    # NEW — AI chat interface
│   │   │   └── InsightsPanel.vue # NEW (Phase 5)
│   │   └── ...
│   └── ...
└── ...
```

---

## Dependencies (New)

### Python (`ai/requirements.txt`)
```
chromadb
ollama
fastapi
uvicorn
psycopg2-binary
rich
```

### Go (add to `go.mod`)
- No new deps — uses `net/http` to proxy to Python service

### Frontend (add to `package.json`)
- No new deps — uses existing Vue.js + Tailwind

---

## Running the Full Stack

```bash
# Terminal 1: PostgreSQL + Redis (already running)

# Terminal 2: Go backend
cd backend && go run cmd/migrate/main.go

# Terminal 3: Python AI service (NEW)
cd ai && uvicorn server:app --port 8090

# Terminal 4: Vue.js frontend
cd frontend && npm run serve
```

---

## Portfolio Positioning

This project demonstrates:
- **RAG pipeline** on structured data (not just documents)
- **Multi-service architecture** (Go API + Python AI sidecar + Vue.js frontend)
- **Applied NLP** (classification, summarization, semantic search)
- **Real-world data engineering** (Wikipedia scraping, geocoding, court record linking)
- **Production patterns** (caching with Redis, spatial indexing with PostGIS, full-text search)

Blog post idea: "Adding AI to a Full-Stack Tracking App: RAG Chat Over 1,500 Court Records"
