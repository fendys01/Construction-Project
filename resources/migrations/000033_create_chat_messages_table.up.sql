CREATE TABLE chat_messages (
	id SERIAL PRIMARY KEY,
	chat_group_id INT REFERENCES chat_groups(id) NOT NULL,
    user_id INT NOT NULL,
	role VARCHAR(10) NOT NULL,
	messages TEXT NOT NULL,
	is_read bool NULL DEFAULT false,
	created_date TIMESTAMPTZ(0) NOT NULL
);