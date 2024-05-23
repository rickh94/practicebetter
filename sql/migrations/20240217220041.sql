-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Drop "scales" table
DROP TABLE `scales`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
