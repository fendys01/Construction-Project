create table order_payments (
    id serial primary key,
    order_id int references orders(id),
    payment_type varchar (20) not null,
    amount BIGINT NOT NULL CHECK (amount> 0),
    payment_status varchar(10) not null,
    expired_date timestamptz(0) not null,
    created_date timestamptz(0) not null
);