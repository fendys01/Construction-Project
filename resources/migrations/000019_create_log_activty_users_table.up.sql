create table log_activity_users(
    id serial PRIMARY KEY,
    user_id int not null,
    role varchar(10) not null,
    title varchar(20) not null, 
    activity varchar(100) NOT NULL,
    event_type varchar(20) NOT NULL,
    created_date timestamptz(0) NOT NULL
)