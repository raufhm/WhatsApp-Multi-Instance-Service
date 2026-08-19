-- Reverse the composite constraint back to the original two-column form.
-- WARNING: If distinct group + personal rows exist for the same
-- (tenant_id, normalized_address), this will fail.  Deduplicate first.

BEGIN;

ALTER TABLE contacts DROP CONSTRAINT IF EXISTS contacts_tenant_id_normalized_address_is_group_key;
ALTER TABLE contacts ADD CONSTRAINT contacts_tenant_id_normalized_address_key
    UNIQUE (tenant_id, normalized_address);

COMMIT;
