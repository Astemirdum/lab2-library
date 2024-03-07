-- +goose Up
CREATE TABLE IF NOT EXISTS events
(
    timestamp TIMESTAMP NOT NULL,
    username VARCHAR(80) NOT NULL,
    reservation_uid uuid        NOT NULL,
    book_uid        uuid        NOT NULL,
    library_uid     uuid        NOT NULL,
    rating int NOT NULL,
    event_type text NOT NULL check ( length(event_type) <= 100 ),
    simplex text NOT NULL check (simplex = any(array['UP','DOWN']))
);

-- +goose Down
DROP TABLE IF EXISTS events CASCADE;