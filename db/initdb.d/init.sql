-- For postgres /docker-entrypoint-initdb.d

CREATE ROLE skabelon WITH NOINHERIT LOGIN PASSWORD 'skabelon_pwd';
GRANT CREATE ON DATABASE postgres TO skabelon;
