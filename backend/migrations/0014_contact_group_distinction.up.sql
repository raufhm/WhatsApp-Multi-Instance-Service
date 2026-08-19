-- Replace the contacts uniqueness constraint to include is_group so that a
-- group and a personal contact sharing the same phone prefix (e.g. 12345678)
-- are stored as distinct rows.  The old constraint
--   UNIQUE (tenant_id, normalized_address)
-- would collapse them into one row, losing the type distinction and
-- overwriting is_group or display_name on upsert.
--
-- The address normalizer already appends @g.us for groups and @s.whatsapp.net
-- for personal contacts, so most real data will not collide.  The composite
-- index is a defense-in-depth guard for edge cases and for the public
-- UpsertContact path where the caller supplies IsGroup but the raw provider
-- address may not contain the server suffix.

BEGIN;

-- Drop the old single-column uniqueness constraint.
ALTER TABLE contacts DROP CONSTRAINT IF EXISTS contacts_tenant_id_normalized_address_key;

-- Create the new composite uniqueness constraint that includes is_group.
ALTER TABLE contacts ADD CONSTRAINT contacts_tenant_id_normalized_address_is_group_key
    UNIQUE (tenant_id, normalized_address, is_group);

COMMIT;
