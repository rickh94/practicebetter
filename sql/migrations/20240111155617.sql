-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Drop "practice_piece" table
DROP TABLE `practice_piece`;
-- Drop "practice_sessions" table
DROP TABLE `practice_sessions`;
-- Create "new_practice_plans" table
CREATE TABLE `new_practice_plans` (
  `id` text NOT NULL,
  `user_id` text NOT NULL,
  `intensity` text NOT NULL,
  `date` integer NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (intensity IN ('light', 'medium', 'heavy'))
);
-- Copy rows from old table "practice_plans" to new temporary table "new_practice_plans"
INSERT INTO `new_practice_plans` (`id`, `user_id`, `intensity`, `date`, `completed`) SELECT `id`, `user_id`, `intensity`, `date`, `completed` FROM `practice_plans`;
-- Drop "practice_plans" table after copying rows
DROP TABLE `practice_plans`;
-- Rename temporary table "new_practice_plans" to "practice_plans"
ALTER TABLE `new_practice_plans` RENAME TO `practice_plans`;
-- Create index "practice_plans_user_id" to table: "practice_plans"
CREATE INDEX `practice_plans_user_id` ON `practice_plans` (`user_id`);
-- Drop "practice_spot" table
DROP TABLE `practice_spot`;
-- Create index "spots_piece_stage" to table: "spots"
CREATE INDEX `spots_piece_stage` ON `spots` (`piece_id`, `stage`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
