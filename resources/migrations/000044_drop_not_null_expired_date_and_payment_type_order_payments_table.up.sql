ALTER TABLE order_payments 
    ALTER COLUMN payment_type DROP NOT NULL,
    ALTER COLUMN expired_date DROP NOT NULL;