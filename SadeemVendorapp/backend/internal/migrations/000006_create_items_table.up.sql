CREATE TABLE items (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id     uuid NOT NULL,
    name          VARCHAR(255) NOT NULL,
    price         DECIMAL(10,2) NOT NULL,
    img           VARCHAR(255),
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_vendor_id
    FOREIGN KEY (vendor_id)
        REFERENCES vendors (id)
        ON DELETE CASCADE
);
ALTER TABLE items ADD COLUMN discount DECIMAL(5,2) DEFAULT 0;



ALTER TABLE items
ADD COLUMN discount_expires_at TIMESTAMP;

CREATE OR REPLACE FUNCTION update_discount_on_expiry()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.discount_expires_at IS NOT NULL AND OLD.discount_expires_at < NOW() THEN
        NEW.discount := 0;
        NEW.discount_expires_at := NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER discount_expiry_trigger
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION update_discount_on_expiry();






ALTER TABLE items ADD COLUMN quantity INT NOT NULL DEFAULT 0;