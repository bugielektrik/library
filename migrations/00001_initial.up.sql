DO $$
  DECLARE author_id UUID;

  BEGIN
    -- EXTENSIONS --
    CREATE EXTENSION IF NOT EXISTS pgcrypto;

    -- TABLES --
    CREATE TABLE IF NOT EXISTS authors (
        created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        id          UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
        full_name   VARCHAR NOT NULL,
        pseudonym   VARCHAR NOT NULL,
        specialty   VARCHAR NOT NULL
    );

    CREATE TABLE IF NOT EXISTS books (
        created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        id          UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
        author_ids  UUID[] NOT NULL,
        name        VARCHAR NOT NULL,
        genre       VARCHAR NOT NULL,
        isbn        VARCHAR NOT NULL UNIQUE,
        rating      NUMERIC NOT NULL DEFAULT 0,
        is_archived BOOLEAN NOT NULL DEFAULT FALSE,
        description JSONB NOT NULL,
        FOREIGN KEY (author_id) REFERENCES authors (id)
    );

    CREATE TABLE IF NOT EXISTS members (
        created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        id          UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
        book_ids    UUID[] NOT NULL,
        full_name   VARCHAR NOT NULL
    );

    -- DATA --
    INSERT INTO authors (full_name, pseudonym, specialty)
    VALUES ('Масару Ибука', 'Гений-изобретатель', 'Раннее развитие ребенка')
    RETURNING id INTO author_id;

    INSERT INTO books (author_id, name, genre, isbn, description)
    VALUES (author_id, 'После трёх уже поздно', 'Книги воспитание детей', '9785916710724', '{"title": "Книга После трех уже поздно", "author": " Ибука Масару"}')
    RETURNING id INTO author_id;

  COMMIT;
END $$;