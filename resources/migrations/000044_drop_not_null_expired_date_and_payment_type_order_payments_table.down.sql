ALTER TABLE order_payments 
    ALTER COLUMN payment_type SET NOT NULL,
    ALTER COLUMN expired_date SET NOT NULL;