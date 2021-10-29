ALTER TABLE orders 
    ADD COLUMN details json NOT NULL DEFAULT '[]'::JSON;