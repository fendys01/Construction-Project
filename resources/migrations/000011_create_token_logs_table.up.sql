create table token_logs (
    id serial primary key,
    channel varchar(20) not null,
    used_for varchar(20) not null, 
    via varchar(5) not null,
    username varchar(50) not null,
    token varchar(6) not null,
    exp_date timestamptz(0) not null,
    created_date timestamptz(0) not null
);