ALTER TABLE chat_groups 
RENAME token TO chat_group_code;

ALTER TABLE chat_groups 
DROP CONSTRAINT chat_groups_member_itin_id_fkey;
    
ALTER TABLE chat_groups 
DROP CONSTRAINT chat_groups_tc_id_fkey;