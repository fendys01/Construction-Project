create table device_list (
    id serial primary key,
    app_id varchar(250) not null,
    identifier varchar(250) NOT NULL,
    language varchar(20) NOT NULL,
    timezone integer NOT NULL DEFAULT 0,
    game_version varchar(100) not null, --app version (1.1)
    device_os varchar(100) not null,
    device_type  integer NOT NULL DEFAULT 0,
    device_model varchar(100),
    created_date timestamptz(0) NOT NULL
)