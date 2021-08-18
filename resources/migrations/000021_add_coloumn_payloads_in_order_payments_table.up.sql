ALTER TABLE order_payments 
  ADD COLUMN payment_url VARCHAR(100) NULL,
  ADD COLUMN payloads JSON NULL;
