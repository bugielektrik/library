CREATE TABLE IF NOT EXISTS authors
  (
     created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     id           SERIAL PRIMARY KEY,
     full_name    VARCHAR NOT NULL,
     pseudonym    VARCHAR NOT NULL,
     specialty    VARCHAR NOT NULL
  );

CREATE TABLE IF NOT EXISTS books
  (
     created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     id           SERIAL PRIMARY KEY,
     name         VARCHAR NOT NULL,
     genre        VARCHAR NOT NULL,
     isbn         VARCHAR NOT NULL,
     authors      VARCHAR ARRAY NOT NULL
  );

CREATE TABLE IF NOT EXISTS members
  (
     created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     id           SERIAL PRIMARY KEY,
     full_name    VARCHAR NOT NULL,
     books        VARCHAR ARRAY NOT NULL
  );