CREATE TABLE IF NOT EXISTS provinces (
    id          INTEGER PRIMARY KEY,
    external_id TEXT UNIQUE NOT NULL,
    code        TEXT,
    name        TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cities (
    id          INTEGER PRIMARY KEY,
    external_id TEXT UNIQUE NOT NULL,
    province_id INTEGER NOT NULL REFERENCES provinces (id),
    name        TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_cities_province_id ON cities (province_id);

CREATE TABLE IF NOT EXISTS districts (
    id          INTEGER PRIMARY KEY,
    external_id TEXT UNIQUE NOT NULL,
    city_id     INTEGER NOT NULL REFERENCES cities (id),
    name        TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_districts_city_id ON districts (city_id);

CREATE TABLE IF NOT EXISTS zips (
    id          INTEGER PRIMARY KEY,
    external_id TEXT UNIQUE NOT NULL,
    code        TEXT NOT NULL,
    city_id     INTEGER NOT NULL REFERENCES cities (id),
    district_id INTEGER NOT NULL REFERENCES districts (id)
);
CREATE INDEX IF NOT EXISTS idx_zips_district_id ON zips (district_id);

CREATE TABLE IF NOT EXISTS villages (
    id          INTEGER PRIMARY KEY,
    external_id TEXT UNIQUE NOT NULL,
    district_id INTEGER NOT NULL REFERENCES districts (id),
    zip_id      INTEGER NULL REFERENCES zips (id),
    name        TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_villages_district_id ON villages (district_id);
CREATE INDEX IF NOT EXISTS idx_villages_name ON villages (name);
