CREATE TABLE chat_groups (
	id SERIAL PRIMARY KEY,
	created_by INT REFERENCES members(id) NOT NULL,
	member_itin_id INT REFERENCES member_itins(id) NULL,
    tc_id INT REFERENCES users(id) NULL,
	token VARCHAR(250) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
	created_date TIMESTAMPTZ(0) NOT NULL
);