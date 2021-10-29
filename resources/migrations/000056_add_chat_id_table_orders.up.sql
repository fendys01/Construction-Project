ALTER TABLE orders ADD COLUMN chat_id INT;

 ALTER TABLE orders 
   ADD CONSTRAINT fk_chat_to_orders_table
   FOREIGN KEY (chat_id) 
   REFERENCES chat_groups(id);