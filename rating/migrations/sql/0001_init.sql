-- +goose Up
CREATE TABLE IF NOT EXISTS rating
(
    id       SERIAL PRIMARY KEY,
    username VARCHAR(80) NOT NULL,
    stars    INT         NOT NULL
        CHECK (stars BETWEEN 0 AND 100)
);

insert into rating (username, stars) VALUES ('Test Max', 90);

-- +goose Down
DROP TABLE IF EXISTS rating CASCADE;