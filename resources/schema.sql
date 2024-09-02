drop schema if exists DataOrders cascade;

create schema if not exists DataOrders;

create table DataOrders.delivery (
     id serial primary key,
     name varchar(255) not null,
     phone varchar(255) not null,
     zip varchar(255) not null,
     city varchar(255) not null,
     address varchar(255) not null,
     region varchar(255) not null,
     email varchar(255) not null

 --   constraint unique_delivery unique (name, phone, zip, city, address, region, email)
);

create table DataOrders.payment (
    transaction varchar(255) primary key,
    request_id varchar(255),
    currency varchar(255) not null,
    provider varchar(255) not null,
    amount float not null ,
    payment_dt integer not null ,
    bank varchar(255) not null ,
    delivery_cost float not null ,
    goods_total integer not null ,
    custom_fee float default 0

);

create table DataOrders.items (
    chrt_id bigserial primary key ,
    track_number varchar(255) not null ,
    price float not null ,
    rid varchar(255) not null ,
    name varchar(255) not null ,
    sale float not null ,
    size varchar(255) not null ,
    total_price float not null ,
    nm_id bigint not null ,
    brand varchar(255) not null ,
    status int not null
);

create table DataOrders.orders (
    order_uid varchar(255) not null primary key ,
    track_number varchar(255) not null,
    entry varchar(255) not null,
    delivery bigint,
    payment varchar(255) not null ,
    items bigserial,
    locale varchar(255) not null ,
    internal_signature varchar(255) not null ,
    customer_id varchar(255) not null ,
    delivery_service varchar(255) not null ,
    shardkey varchar(255) not null ,
    sm_id bigint not null ,
    date_created timestamp not null default current_date,
    oof_shard varchar(255) not null,

    constraint fk_orders_delivery foreign key (delivery) references DataOrders.delivery(id),
    constraint fk_orders_payment foreign key (payment) references DataOrders.payment(transaction),
    constraint fk_orders_items foreign key (items) references DataOrders.items(chrt_id)
);




