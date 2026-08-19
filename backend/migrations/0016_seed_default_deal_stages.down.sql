-- Reverse: delete all seeded deal stages (they will be re-created if needed)
DELETE FROM deal_stage_history;
DELETE FROM deal_stages;
