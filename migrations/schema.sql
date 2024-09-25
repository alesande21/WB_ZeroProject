CREATE TABLE IF NOT EXISTS orders (
                                      order_uid VARCHAR(100) PRIMARY KEY ,
                                      track_number VARCHAR(100) NOT NULL ,
                                      entry VARCHAR(50) NOT NULL ,
                                      locale VARCHAR(50) NOT NULL  ,
                                      internal_signature VARCHAR(50) NOT NULL  ,
                                      customer_id VARCHAR(100) NOT NULL  ,
                                      delivery_service VARCHAR(100) NOT NULL  ,
                                      shardkey VARCHAR(10) NOT NULL  ,
                                      sm_id BIGINT NOT NULL  ,
                                      date_created TIMESTAMPTZ NOT NULL,
                                      oof_shard VARCHAR(5) NOT NULL

    --   constraint fk_orders_delivery foreign key (order_uid) references delivery(id),
    --   constraint fk_orders_payment foreign key (order_uid) references payment(transaction),
    --   constraint fk_orders_items foreign key (order_uid) references items(chrt_id)
);

CREATE TABLE IF NOT EXISTS delivery (
     id SERIAL PRIMARY KEY,
     order_id VARCHAR(100) REFERENCES orders(order_uid) ON DELETE CASCADE,
     name VARCHAR(100) NOT NULL ,
     phone VARCHAR(50) NOT NULL ,
     zip VARCHAR(20) NOT NULL ,
     city VARCHAR(30) NOT NULL ,
     address VARCHAR(100) NOT NULL ,
     region VARCHAR(50) NOT NULL ,
     email VARCHAR(50) NOT NULL

    --   constraint unique_delivery unique (name, phone, zip, city, address, region, email)
);

CREATE TABLE IF NOT EXISTS payment (
    transaction_id VARCHAR(100) REFERENCES orders(order_uid) ON DELETE CASCADE,
    request_id VARCHAR(100) NOT NULL ,
    currency VARCHAR(50) NOT NULL ,
    provider VARCHAR(100) NOT NULL ,
    amount FLOAT NOT NULL  ,
    payment_dt INTEGER NOT NULL  ,
    bank VARCHAR(100) NOT NULL  ,
    delivery_cost FLOAT NOT NULL  ,
    goods_total INTEGER NOT NULL  ,
    custom_fee FLOAT NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
    chrt_id BIGINT PRIMARY KEY ,
    order_id VARCHAR(100) REFERENCES orders(order_uid) ON DELETE CASCADE,
    track_number VARCHAR(100) NOT NULL  ,
    price FLOAT NOT NULL  ,
    rid VARCHAR(100) NOT NULL  ,
    name VARCHAR(100) NOT NULL  ,
    sale FLOAT NOT NULL  ,
    size VARCHAR(50) NOT NULL  ,
    total_price FLOAT NOT NULL  ,
    nm_id BIGINT NOT NULL  ,
    brand VARCHAR(50) NOT NULL  ,
    status INT NOT NULL
);



