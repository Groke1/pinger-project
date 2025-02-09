CREATE TABLE pings (
    id SERIAL PRIMARY KEY,
    ip INET NOT NULL,
    duration INT NOT NULL,
    time_attempt TIMESTAMP NOT NULL,
    UNIQUE (ip)
)