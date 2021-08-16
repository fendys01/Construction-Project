create table settings (
    id serial primary key,
    set_group varchar(50) not null,
    set_key varchar(50) not null,
    set_label varchar(100) not null unique,
    content_type varchar(6) not null, -- string, json, bool, 
    content_value text not null,
    is_active boolean default false,
    created_date timestamptz(0) not null,
    updated_date timestamptz(0) null
);