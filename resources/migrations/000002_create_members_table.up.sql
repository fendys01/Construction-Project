-- member code example: m-20210617-avcd02
CREATE TABLE members (
	id serial PRIMARY KEY,
	member_code varchar(28) unique not null,
	name varchar(50) NOT NULL,
	username varchar(50) UNIQUE NOT NULL,
	email varchar(100) UNIQUE NOT NULL,
	phone varchar(15) UNIQUE NOT NULL,
	"password" varchar(100) NOT NULL,
	img varchar(200) NULL,
	is_valid_email boolean default false,
	is_valid_phone boolean default false,
	is_active boolean default false,
	created_date timestamptz(0) NOT NULL,
	updated_date timestamptz(0) NULL
);
