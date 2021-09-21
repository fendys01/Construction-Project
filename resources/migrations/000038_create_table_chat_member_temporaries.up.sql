CREATE TABLE chat_member_temporaries (
	id serial PRIMARY KEY,
	email varchar(100) NOT NULL,
    chat_group_id int references chat_groups(id) NOT NULL,
	created_date timestamptz(0) NOT NULL
);
