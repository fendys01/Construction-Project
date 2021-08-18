create table orders (
    id serial primary key,
    member_itin_id int references member_itins(id),
    paid_by int references members(id),
    order_code varchar(20) not null unique,
    order_status varchar(8), -- booked, issued, canceled, expired
    total_price bigint not null,
    created_date timestamptz(0) not null
);