CREATE TABLE itin_suggestions (
	id serial PRIMARY KEY,
	itin_code varchar(20) NOT NULL unique,
    created_by  int REFERENCES users(id),
	title varchar(250) NOT NULL,
    content TEXT NOT NULL,
    img varchar(250) NULL,
    price BIGINT NOT NULL CHECK (price> 0),
    start_date timestamp,
    end_date timestamp,
    details json NOT NULL,
    created_date timestamptz(0) NOT NULL,
	updated_date timestamptz(0) NULL,
	deleted_date timestamptz(0) NULL
);