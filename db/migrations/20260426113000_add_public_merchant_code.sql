-- Human-friendly public merchant identifier.
-- Keep the UUID primary key for internal relations and add a shareable code for UI/API use.

CREATE OR REPLACE FUNCTION generate_merchant_code()
RETURNS TEXT AS $$
DECLARE
    chars TEXT := 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
    candidate TEXT;
    code_exists BOOLEAN;
BEGIN
    LOOP
        candidate := 'LEG-';

        FOR i IN 1..6 LOOP
            candidate := candidate || substr(chars, 1 + floor(random() * length(chars))::INT, 1);
        END LOOP;

        SELECT EXISTS (
            SELECT 1
            FROM merchants
            WHERE merchant_code = candidate
        ) INTO code_exists;

        IF NOT code_exists THEN
            RETURN candidate;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE merchants
ADD COLUMN IF NOT EXISTS merchant_code TEXT;

UPDATE merchants
SET merchant_code = generate_merchant_code()
WHERE merchant_code IS NULL;

ALTER TABLE merchants
ALTER COLUMN merchant_code SET DEFAULT generate_merchant_code();

ALTER TABLE merchants
ALTER COLUMN merchant_code SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_merchants_merchant_code
ON merchants(merchant_code);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'merchants_merchant_code_format_chk'
    ) THEN
        ALTER TABLE merchants
        ADD CONSTRAINT merchants_merchant_code_format_chk
        CHECK (merchant_code ~ '^LEG-[A-HJ-NP-Z2-9]{6}$');
    END IF;
END $$;
