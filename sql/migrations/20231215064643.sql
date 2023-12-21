-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_practice_plans" table
CREATE TABLE `new_practice_plans` (
  `id` text NOT NULL,
  `user_id` text NOT NULL,
  `intensity` text NOT NULL,
  `date` integer NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  `practice_session_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`practice_session_id`) REFERENCES `practice_sessions` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (intensity IN ('light', 'medium', 'heavy'))
);
-- Copy rows from old table "practice_plans" to new temporary table "new_practice_plans"
INSERT INTO `new_practice_plans` (`id`, `user_id`, `intensity`, `date`, `completed`) SELECT `id`, `user_id`, `intensity`, `date`, `completed` FROM `practice_plans`;
-- Drop "practice_plans" table after copying rows
DROP TABLE `practice_plans`;
-- Rename temporary table "new_practice_plans" to "practice_plans"
ALTER TABLE `new_practice_plans` RENAME TO `practice_plans`;
-- Create "new_practice_plan_spots" table
CREATE TABLE `new_practice_plan_spots` (
  `practice_plan_id` text NOT NULL,
  `spot_id` text NOT NULL,
  `practice_type` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  PRIMARY KEY (`practice_plan_id`, `spot_id`),
  CONSTRAINT `0` FOREIGN KEY (`spot_id`) REFERENCES `spots` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (practice_type IN ('new', 'more_repeat', 'interleave', 'interleave_days'))
);
-- Copy rows from old table "practice_plan_spots" to new temporary table "new_practice_plan_spots"
INSERT INTO `new_practice_plan_spots` (`practice_plan_id`, `spot_id`, `practice_type`, `completed`) SELECT `practice_plan_id`, `spot_id`, `practice_type`, `completed` FROM `practice_plan_spots`;
-- Drop "practice_plan_spots" table after copying rows
DROP TABLE `practice_plan_spots`;
-- Rename temporary table "new_practice_plan_spots" to "practice_plan_spots"
ALTER TABLE `new_practice_plan_spots` RENAME TO `practice_plan_spots`;
-- Drop "practice_plan_sessions" table
DROP TABLE `practice_plan_sessions`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
