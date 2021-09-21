ALTER TABLE device_list 
    DROP COLUMN identifier,
    DROP COLUMN language,
    DROP COLUMN timezone,
    DROP COLUMN game_version,
    DROP COLUMN device_os,
    DROP COLUMN device_type,
    DROP COLUMN device_model;

ALTER TABLE notifications 
    DROP COLUMN member_code;

ALTER TABLE device_list 
    ADD COLUMN user_id INT NULL,
	ADD COLUMN role VARCHAR(10) NULL;

ALTER TABLE notifications 
    ADD COLUMN subject VARCHAR(128) NULL,
    ADD COLUMN user_id INT NULL,
	ADD COLUMN role VARCHAR(10) NULL;