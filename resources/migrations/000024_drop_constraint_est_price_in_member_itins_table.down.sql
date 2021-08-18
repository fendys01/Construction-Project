ALTER TABLE member_itins 
    ADD CONSTRAINT member_itins_est_price_check CHECK ((est_price > 0));