CREATE SCHEMA IF NOT EXISTS henhold AUTHORIZATION skabelon;

CREATE TABLE IF NOT EXISTS henhold.risk (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    key TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL
);

INSERT INTO henhold.risk (key, description) VALUES
('RSK-1', 'High risk'),
('RSK-2', 'Medium risk'),
('RSK-3', 'Low risk')
ON CONFLICT DO NOTHING;
