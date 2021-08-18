ALTER TABLE member_itin_relations 
  DROP CONSTRAINT member_itin_relations_member_itin_id_fkey;

ALTER TABLE member_itin_relations 
  ADD CONSTRAINT member_itin_relations_member_itin_id_fkey
  FOREIGN KEY (member_itin_id) 
  REFERENCES member_itins(id);