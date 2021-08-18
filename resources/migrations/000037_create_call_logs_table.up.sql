CREATE TABLE call_logs (
	id SERIAL PRIMARY KEY,
    trx_id VARCHAR(50) UNIQUE NOT NULL,
    provider VARCHAR(50) NOT NULL,
	call_type VARCHAR(10) NOT NULL,
    bill_price int8 NULL,
    payloads JSON NOT NULL,
	created_date TIMESTAMPTZ(0) NOT NULL
);