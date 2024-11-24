CREATE TABLE user (
    id serial primary key,
    datecreated text,
    datemodified text,
    username text,
    name text,
    email text unique,
    hash text,
    role text
);
