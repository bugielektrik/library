CREATE TABLE authors(
    id text PRIMARY KEY not null ,
    full_name text not null ,
    pseudonym text not null,
    specialty text not null);

CREATE TABLE books(
    id text PRIMARY KEY NOT NULL,
    name text NOT NULL,
    genre text NOT NULL,
    isbn text NOT NULL,
    author_id text NOT NULL);

CREATE TABLE members(
    id text PRIMARY KEY NOT NULL,
    full_name text NOT NULL);

CREATE TABLE members_and_books(
    book_id text NOT NULL,
    member_id text NOT NULL,
    UNIQUE (book_id, member_id));
