ALTER TABLE orders ADD COLUMN tc_id INT;

 ALTER TABLE orders 
   ADD CONSTRAINT fk_users_to_orders_table
   FOREIGN KEY (tc_id) 
   REFERENCES users(id);