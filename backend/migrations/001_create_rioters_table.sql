-- Create rioters table
CREATE TABLE IF NOT EXISTS rioters (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    middle_name VARCHAR(255),
    city VARCHAR(255),
    state VARCHAR(100),
    age INTEGER,
    summary TEXT,
    jurisdiction VARCHAR(255),
    charges TEXT,
    charges_link VARCHAR(500),
    case_status VARCHAR(100),
    case_updates TEXT,
    photo_name VARCHAR(255),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    geom GEOGRAPHY(POINT, 4326),
    violence_assault BOOLEAN DEFAULT FALSE,
    conspiracy BOOLEAN DEFAULT FALSE,
    theft BOOLEAN DEFAULT FALSE,
    property BOOLEAN DEFAULT FALSE,
    military_le BOOLEAN DEFAULT FALSE,
    extremist BOOLEAN DEFAULT FALSE,
    inspired_trump BOOLEAN DEFAULT FALSE,
    sentenced BOOLEAN DEFAULT FALSE,
    commuted BOOLEAN DEFAULT FALSE,
    pardoned BOOLEAN DEFAULT FALSE,
    arrest_date DATE,
    search_vector TSVECTOR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_rioters_state ON rioters(state);
CREATE INDEX IF NOT EXISTS idx_rioters_case_status ON rioters(case_status);
CREATE INDEX IF NOT EXISTS idx_rioters_geom ON rioters USING GIST(geom);
CREATE INDEX IF NOT EXISTS idx_rioters_search_vector ON rioters USING GIN(search_vector);

-- Create function to update search_vector
CREATE OR REPLACE FUNCTION update_rioters_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := 
        setweight(to_tsvector('english', COALESCE(NEW.first_name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.last_name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.city, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.state, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.charges, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.summary, '')), 'C');
    NEW.updated_at := CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search_vector
CREATE TRIGGER update_rioters_search_vector_trigger
    BEFORE INSERT OR UPDATE ON rioters
    FOR EACH ROW
    EXECUTE FUNCTION update_rioters_search_vector();

-- Update existing rows (if any) to populate search_vector
UPDATE rioters SET search_vector = 
    setweight(to_tsvector('english', COALESCE(first_name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(last_name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(city, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(state, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(charges, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(summary, '')), 'C');

