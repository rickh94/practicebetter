-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_users" table
CREATE TABLE `new_users` (
  `id` text NOT NULL,
  `fullname` text NOT NULL DEFAULT '',
  `email` text NOT NULL,
  `email_verified` boolean NULL DEFAULT 0,
  `active_practice_plan_id` text NULL,
  `active_practice_plan_started` integer NULL,
  `config_default_plan_intensity` text NOT NULL DEFAULT 'medium',
  `config_time_between_breaks` integer NOT NULL DEFAULT 33,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`active_practice_plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CHECK (config_default_plan_intensity IN ('light', 'medium', 'heavy')),
  CHECK (config_time_between_breaks > 5),
  CHECK (config_time_between_breaks < 100),
  CHECK (email_verified IN (0, 1))
);
-- Copy rows from old table "users" to new temporary table "new_users"
INSERT INTO `new_users` (`id`, `fullname`, `email`, `email_verified`, `active_practice_plan_id`, `active_practice_plan_started`) SELECT `id`, `fullname`, `email`, `email_verified`, `active_practice_plan_id`, `active_practice_plan_started` FROM `users`;
-- Drop "users" table after copying rows
DROP TABLE `users`;
-- Rename temporary table "new_users" to "users"
ALTER TABLE `new_users` RENAME TO `users`;
-- Create index "users_email" to table: "users"
CREATE UNIQUE INDEX `users_email` ON `users` (`email`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
