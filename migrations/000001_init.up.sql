create table if not exists users(
    id serial primary key,
    login text unique not null,
    password text not null,
    refresh_token text
);