CREATE TABLE member_temporaries (
	id serial PRIMARY KEY,
	email varchar(100) UNIQUE NOT NULL,
    member_itin_id int references member_itins(id) NOT NULL,
	created_date timestamptz(0) NOT NULL
);
