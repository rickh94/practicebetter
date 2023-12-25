UPDATE spots
SET stage_started = unixepoch('now')
WHERE spots.stage_started IS NULL;
