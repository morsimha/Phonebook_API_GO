-- מאפשר חיפוש עם GIN + pg_trgm
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- יצירת הטבלה (אם לא קיימת)
CREATE TABLE IF NOT EXISTS contacts (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- אינדקסים לשאילתות ILIKE
CREATE INDEX IF NOT EXISTS idx_contacts_first_name ON contacts USING GIN (first_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_contacts_last_name ON contacts USING GIN (last_name gin_trgm_ops);
