CREATE TABLE chat_group_relations (
	id SERIAL PRIMARY KEY,
	member_id INT REFERENCES members(id) NOT NULL,
	chat_group_id INT REFERENCES chat_groups(id) NULL,
	created_date TIMESTAMPTZ(0) NOT NULL,
    deleted_date TIMESTAMPTZ(0) NULL
);