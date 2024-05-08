CREATE TABLE users
(
    id                 bigserial not null primary key,
    email              varchar   not null unique,
    encrypted_password varchar   not null
);


CREATE TABLE menu_items
(
    id bigserial not null primary key,
    name varchar not null,
    price int not null,
    description varchar
)
