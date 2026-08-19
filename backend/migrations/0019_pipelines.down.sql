DROP INDEX IF EXISTS idx_contacts_deal_stage_id;
ALTER TABLE contacts DROP COLUMN IF EXISTS deal_stage_id;
DROP INDEX IF EXISTS idx_deal_stages_pipeline;
ALTER TABLE deal_stages DROP COLUMN IF EXISTS pipeline_id;
DROP TABLE IF EXISTS pipelines;
