CREATE SCHEMA IF NOT EXISTS skabelon AUTHORIZATION skabelon;

CREATE TABLE IF NOT EXISTS skabelon.resource (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    rkey TEXT NOT NULL UNIQUE,
    is_fun BOOLEAN NULL,
    my_int INT,
    description TEXT NOT NULL
);

INSERT INTO skabelon.resource (rkey, description) VALUES
('RSK-1', 'High risk'),
('RSK-2', 'Medium risk'),
('RSK-3', 'Low risk')
ON CONFLICT DO NOTHING;
