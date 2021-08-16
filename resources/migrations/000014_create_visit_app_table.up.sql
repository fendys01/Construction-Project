create table log_visit_app(
    id serial PRIMARY KEY,
    member_id int references members(id) not null,
    total_visited int NOT NULL,
    last_active_date timestamptz(0) NOT NULL
)