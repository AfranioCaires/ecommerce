ALTER TABLE customers ADD COLUMN name TEXT;
UPDATE customers SET name = split_part(email, '@', 1);
ALTER TABLE customers ALTER COLUMN name SET NOT NULL;
