ALTER TABLE device_list 
    ADD COLUMN device_type integer NOT NULL,
    ADD COLUMN device_model varchar(100);
