-- +goose Up

CREATE TABLE IF NOT EXISTS library
(
    id          int generated always as identity PRIMARY KEY,
    library_uid uuid UNIQUE  NOT NULL,
    name        VARCHAR(80)  NOT NULL,
    city        VARCHAR(255) NOT NULL,
    address     VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS books
(
    id        int generated always as identity PRIMARY KEY,
    book_uid  uuid UNIQUE  NOT NULL,
    name      VARCHAR(255) NOT NULL,
    author    VARCHAR(255),
    genre     VARCHAR(255),
    condition VARCHAR(20) DEFAULT 'EXCELLENT'
        CHECK (condition IN ('EXCELLENT', 'GOOD', 'BAD'))
);

CREATE TABLE IF NOT EXISTS library_books
(
    book_id         INT REFERENCES books (id),
    library_id      INT REFERENCES library (id),
    available_count INT NOT NULL CHECK (available_count >= 0)
);

INSERT INTO library (library_uid, name, city, address) VALUES
    ('83575e12-7ce0-48ee-9931-51919ff3c9ee', 'Библиотека имени 7 Непьющих', 'Москва', '2-я Бауманская ул., д.5, стр.1')
;

INSERT INTO books (book_uid, name, author, genre, condition) VALUES
    ('f7cdc58f-2caf-4b15-9727-f89dcc629b27', 'Краткий курс C++ в 7 томах', 'Бьерн Страуструп', 'Научная фантастика', 'EXCELLENT')
;

INSERT INTO library_books (book_id, library_id, available_count) VALUES
    (1, 1, 1)
;


-- +goose Down
DROP TABLE IF EXISTS library_books CASCADE;
DROP TABLE IF EXISTS library CASCADE;
DROP TABLE IF EXISTS books CASCADE;