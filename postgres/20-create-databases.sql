-- file: 20-create-database.sql
CREATE DATABASE reservations;
GRANT ALL PRIVILEGES ON DATABASE reservations TO program;

CREATE DATABASE libraries;
GRANT ALL PRIVILEGES ON DATABASE libraries TO program;

CREATE DATABASE ratings;
GRANT ALL PRIVILEGES ON DATABASE ratings TO program;

CREATE DATABASE stats;
GRANT ALL PRIVILEGES ON DATABASE stats TO program;