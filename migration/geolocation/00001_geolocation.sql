-- +goose Up

-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = now();
RETURN NEW;
END;
$$ language 'plpgsql';
-- +goose StatementEnd

CREATE TABLE geolocations (
                       id                          UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
                       ip                          TEXT NOT NULL UNIQUE DEFAULT '',
                       country_code                TEXT NOT NULL DEFAULT '',
                       country                     TEXT NOT NULL DEFAULT '',
                       city                        TEXT NOT NULL DEFAULT '',
                       latitude                    TEXT NOT NULL DEFAULT '',
                       longitude                   TEXT NOT NULL DEFAULT '',
                       mystery_value               TEXT NOT NULL DEFAULT '',
                       created_at                  TIMESTAMP with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       modified_at                  TIMESTAMP with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX index_ip ON geolocations(ip);

CREATE TRIGGER update_geolocation_modified BEFORE UPDATE ON geolocations FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
-- +goose Down
DROP TRIGGER IF EXISTS update_geolocation_modified on geolocations;

DROP INDEX index_ip;
DROP TABLE geolocations;

DROP FUNCTION IF EXISTS update_modified_column;
