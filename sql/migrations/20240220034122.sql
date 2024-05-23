-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_user_scales" table
CREATE TABLE `new_user_scales` (
  `id` text NOT NULL,
  `user_id` text NOT NULL,
  `scale_id` integer NOT NULL,
  `practice_notes` text NOT NULL,
  `last_practiced` integer NULL,
  `reference` text NOT NULL,
  `working` boolean NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`scale_id`) REFERENCES `scales` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Copy rows from old table "user_scales" to new temporary table "new_user_scales"
INSERT INTO `new_user_scales` (`id`, `user_id`, `scale_id`, `practice_notes`, `last_practiced`, `reference`, `working`) SELECT `id`, `user_id`, `scale_id`, `practice_notes`, `last_practiced`, `reference`, `working` FROM `user_scales`;
-- Drop "user_scales" table after copying rows
DROP TABLE `user_scales`;
-- Rename temporary table "new_user_scales" to "user_scales"
ALTER TABLE `new_user_scales` RENAME TO `user_scales`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
