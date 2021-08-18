DROP TABLE IF EXISTS log_visit_app;

create table log_visit_app(
    id serial PRIMARY KEY,
    user_id int not null,
    role varchar(20) not null,
    total_visited int NOT NULL,
    last_active_date timestamptz(0)
)