CREATE TABLE member_itin_relations (
	id serial PRIMARY KEY,
    member_itin_id  int REFERENCES members(id),
    member_id  int REFERENCES members(id),
    created_date timestamptz(0) NOT NULL,
	deleted_date timestamptz(0) NULL
);