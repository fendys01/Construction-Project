create table notifications (
    id serial primary key,
    code varchar(20) not null,
    member_code varchar(128) NOT NULL,
    type integer NOT NULL,
    title varchar(128) NOT NULL,
    content varchar(512) NOT NULL,
    link varchar(256),
    is_read integer NOT NULL DEFAULT 0,
    created_date timestamptz(0) NOT NULL
)