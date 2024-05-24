-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    id serial  primary key,
    username text NOT NULL unique,
    password text not null,
    email text NOT NULL unique,
    user_type text NOT NULL DEFAULT 'user' check ( user_type in ('user', 'admin') )
);

insert into users (username, password, email, user_type) values
('Test Max', 'admin', 'max@lol.ru',  'admin'),
('User1', 'user1', 'user@lol.ru',  'user');

-- +goose Down
DROP TABLE IF EXISTS users CASCADE;