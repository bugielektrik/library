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
        author_id   UUID NOT NULL,
        name        VARCHAR NOT NULL,
        genre       VARCHAR NOT NULL,
        isbn        INTEGER NOT NULL UNIQUE,
        rating      NUMERIC NOT NULL DEFAULT 0,
        is_archived BOOLEAN NOT NULL DEFAULT FALSE,
        description JSONB NOT NULL,
        FOREIGN KEY (author_id) REFERENCES authors (id)
    );

    CREATE TABLE IF NOT EXISTS members (
        created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        id          UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
        book_id     UUID NOT NULL,
        full_name   VARCHAR NOT NULL,
        FOREIGN KEY (book_id) REFERENCES books (id)
    );

    CREATE TABLE IF NOT EXISTS members_books (
      member_id     UUID NOT NULL,
      book_id       UUID NOT NULL,
      PRIMARY KEY (member_id, book_id),
      FOREIGN KEY (member_id) REFERENCES members(id) ON UPDATE CASCADE,
      FOREIGN KEY (book_id) REFERENCES books(id) ON UPDATE CASCADE
    );

    -- DATA --
    INSERT INTO authors (full_name, pseudonym, specialty)
    VALUES ('Масару Ибука', 'Гений-изобретатель', 'Раннее развитие ребенка')
    RETURNING id INTO author_id;

    INSERT INTO books (author_id, name, genre, isbn, description)
    VALUES (author_id, 'После трёх уже поздно', 'Книги воспитание детей', 9785916710724, '{"seo":{"title":"Книга \"После трех уже поздно\" Ибука Масару  – купить книгу ISBN 978-5-91671-789-1 с быстрой доставкой в интернет-магазине","link":[{"href":"https://book.kz/product/posle-treh-uzhe-pozdno-ibuka-masaru-252509129/","rel":"canonical"}],"meta":[{"content":"Книга \"После трех уже поздно\" Ибука Масару  – купить книгу ISBN 978-5-91671-789-1 с быстрой доставкой в интернет-магазине","name":"og:title","property":"og:title"},{"content":"В наличии Книга \"После трех уже поздно\" (Ибука Масару), Альпина нон-фикшн в интернет-магазине со скидкой! ✅ Отзывы и фото 🚚 Быстрая доставка","name":"og:description","property":"og:description"},{"content":"В наличии Книга \"После трех уже поздно\" (Ибука Масару), Альпина нон-фикшн в интернет-магазине со скидкой! ✅ Отзывы и фото 🚚 Быстрая доставка","name":"description"},{"content":"website","name":"og:type","property":"og:type"},{"content":"noindex,nofollow","name":"robots"},{"content":"1211635852237386","property":"fb:app_id"}]}}');

  COMMIT;
END $$;