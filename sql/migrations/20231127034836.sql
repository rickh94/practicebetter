-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_pieces" table
CREATE TABLE `new_pieces` (
  `id` text NOT NULL,
  `title` text NOT NULL,
  `description` text NULL,
  `composer` text NULL,
  `recording_link` text NULL,
  `measures` integer NULL,
  `beats_per_measure` integer NULL,
  `goal_tempo` integer NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Copy rows from old table "pieces" to new temporary table "new_pieces"
INSERT INTO `new_pieces` (`id`, `title`, `description`, `composer`, `recording_link`, `measures`, `beats_per_measure`, `user_id`) SELECT `id`, `title`, `description`, `composer`, `recording_link`, `measures`, `beats_per_measure`, `user_id` FROM `pieces`;
-- Drop "pieces" table after copying rows
DROP TABLE `pieces`;
-- Rename temporary table "new_pieces" to "pieces"
ALTER TABLE `new_pieces` RENAME TO `pieces`;
-- Create index "pieces_user_id" to table: "pieces"
CREATE INDEX `pieces_user_id` ON `pieces` (`user_id`);
-- Create index "pieces_user_id_title" to table: "pieces"
CREATE INDEX `pieces_user_id_title` ON `pieces` (`user_id`, `title`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
