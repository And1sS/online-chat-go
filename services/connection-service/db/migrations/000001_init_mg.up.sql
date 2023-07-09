CREATE TABLE IF NOT EXISTS users
(
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password TEXT        NOT NULL
);

CREATE INDEX users_username_idx ON users (username);

INSERT INTO users (username, password)
VALUES ('user', '$2a$12$DdaVkx1Q8ckdfUg90Ep6T.a7cvEXdPKjyzHqb5GLg.0yN0yRLCx0y') --! password = user