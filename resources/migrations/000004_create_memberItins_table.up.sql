CREATE TABLE member_itins (
	id serial PRIMARY KEY,
	itin_code varchar(20) NOT NULL unique,
	title varchar(250) NOT NULL,
    created_by  int REFERENCES members(id),
    est_price BIGINT NULL CHECK (est_price> 0),
    start_date timestamp,
    end_date timestamp,
    details json NOT NULL,
    created_date timestamptz(0) NOT NULL,
	updated_date timestamptz(0) NULL,
	deleted_date timestamptz(0) NULL
);