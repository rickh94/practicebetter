-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_practice_sessions" table
CREATE TABLE `new_practice_sessions` (
  `id` text NOT NULL,
  `duration_minutes` integer NOT NULL,
  `date` integer NOT NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Copy rows from old table "practice_sessions" to new temporary table "new_practice_sessions"
INSERT INTO `new_practice_sessions` (`id`, `duration_minutes`, `date`) SELECT `id`, `duration_minutes`, `date` FROM `practice_sessions`;
-- Drop "practice_sessions" table after copying rows
DROP TABLE `practice_sessions`;
-- Rename temporary table "new_practice_sessions" to "practice_sessions"
ALTER TABLE `new_practice_sessions` RENAME TO `practice_sessions`;
-- Create index "practice_sessions_date" to table: "practice_sessions"
CREATE INDEX `practice_sessions_date` ON `practice_sessions` (`date`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
