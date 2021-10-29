CREATE TABLE stuff (
	id serial PRIMARY KEY,
    code_stuff varchar(30) unique NOT NULL,
	name_stuff varchar(100) not null,
	image varchar(200) null,
	description varchar(100) not null,
	price varchar(50) not null,
	type integer NOT NULL DEFAULT 0,
	is_active boolean default true,
	created_date timestamptz(0) NOT null,
	updated_date timestamptz(0) null,
	deleted_date timestamptz(0) null
);
