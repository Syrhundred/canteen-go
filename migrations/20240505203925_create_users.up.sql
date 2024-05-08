CREATE TABLE users
(
    id                 bigserial not null primary key,
    email              varchar   not null unique,
    encrypted_password varchar   not null,
    role varchar   not null
);


CREATE TABLE menu_items
(
    id bigserial not null primary key,
    name varchar not null,
    price int not null,
    description varchar
);

CREATE TABLE orders
(
    id          bigserial not null primary key,
    user_id     int       not null,
    createdAt   date      not null,
    totalAmount int       not null,
    foreign key (user_id) references users (id)
);

CREATE TABLE orderItem
(
    id           serial not null primary key,
    order_id     int    not null,
    menu_item_id int    not null,
    quantity     int    not null,
    foreign key (menu_item_id) references menu_items (id),
    foreign key (order_id) references orders (id)
);