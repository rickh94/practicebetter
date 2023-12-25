-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_practice_spot" table
CREATE TABLE `new_practice_spot` (
  `practice_session_id` text NOT NULL,
  `spot_id` text NOT NULL,
  `reps` integer NOT NULL DEFAULT 1,
  PRIMARY KEY (`practice_session_id`, `spot_id`),
  CONSTRAINT `0` FOREIGN KEY (`spot_id`) REFERENCES `spots` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_session_id`) REFERENCES `practice_sessions` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Copy rows from old table "practice_spot" to new temporary table "new_practice_spot"
INSERT INTO `new_practice_spot` (`practice_session_id`, `spot_id`) SELECT `practice_session_id`, `spot_id` FROM `practice_spot`;
-- Drop "practice_spot" table after copying rows
DROP TABLE `practice_spot`;
-- Rename temporary table "new_practice_spot" to "practice_spot"
ALTER TABLE `new_practice_spot` RENAME TO `practice_spot`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
