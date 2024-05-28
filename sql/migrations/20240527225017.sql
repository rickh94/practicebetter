-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_reading" table
CREATE TABLE `new_reading` (
  `id` text NOT NULL,
  `title` text NOT NULL,
  `info` text NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  `composer` text NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (LENGTH(title) > 0)
);
-- Copy rows from old table "reading" to new temporary table "new_reading"
INSERT INTO `new_reading` (`id`, `title`, `info`, `composer`, `user_id`) SELECT `id`, `title`, `info`, `composer`, `user_id` FROM `reading`;
-- Drop "reading" table after copying rows
DROP TABLE `reading`;
-- Rename temporary table "new_reading" to "reading"
ALTER TABLE `new_reading` RENAME TO `reading`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
