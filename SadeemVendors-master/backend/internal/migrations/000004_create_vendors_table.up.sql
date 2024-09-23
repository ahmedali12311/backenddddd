CREATE TABLE vendors (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name             VARCHAR(255) NOT NULL,
    img              VARCHAR(255),
    description      TEXT,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE vendors
ADD COLUMN subscription_end TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '30 days'
ADD COLUMN is_visible BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN subscription_days INTEGER NOT NULL DEFAULT 0;
CREATE OR REPLACE FUNCTION update_visibility() 
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.subscription_end < CURRENT_TIMESTAMP THEN
        NEW.is_visible := FALSE;
    ELSE
        NEW.is_visible := TRUE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER check_subscription_end
BEFORE INSERT OR UPDATE ON vendors
FOR EACH ROW
EXECUTE FUNCTION update_visibility();
