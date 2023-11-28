-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_spots" table
CREATE TABLE `new_spots` (
  `id` text NOT NULL,
  `piece_id` text NOT NULL,
  `name` text NOT NULL,
  `idx` integer NULL DEFAULT 0,
  `stage` text NOT NULL DEFAULT 'repeat',
  `audio_prompt_url` text NOT NULL DEFAULT '',
  `image_prompt_url` text NOT NULL DEFAULT '',
  `notes_prompt` text NOT NULL DEFAULT '',
  `text_prompt` text NOT NULL DEFAULT '',
  `current_tempo` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (stage IN ('repeat', 'random', 'interleave', 'interleave_days', 'complete'))
);
-- Copy rows from old table "spots" to new temporary table "new_spots"
INSERT INTO `new_spots` (`id`, `piece_id`, `name`, `idx`, `stage`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`) SELECT `id`, `piece_id`, `name`, `idx`, `stage`, IFNULL(`audio_prompt_url`, '') AS `audio_prompt_url`, IFNULL(`image_prompt_url`, '') AS `image_prompt_url`, IFNULL(`notes_prompt`, '') AS `notes_prompt`, IFNULL(`text_prompt`, '') AS `text_prompt`, IFNULL(`current_tempo`, 0) AS `current_tempo` FROM `spots`;
-- Drop "spots" table after copying rows
DROP TABLE `spots`;
-- Rename temporary table "new_spots" to "spots"
ALTER TABLE `new_spots` RENAME TO `spots`;
-- Create index "spots_piece_id" to table: "spots"
CREATE INDEX `spots_piece_id` ON `spots` (`piece_id`);
-- Create "new_pieces" table
CREATE TABLE `new_pieces` (
  `id` text NOT NULL,
  `title` text NOT NULL,
  `description` text NULL,
  `composer` text NOT NULL DEFAULT '',
  `recording_link` text NOT NULL DEFAULT '',
  `measures` integer NOT NULL DEFAULT 0,
  `beats_per_measure` integer NOT NULL DEFAULT 0,
  `goal_tempo` integer NOT NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Copy rows from old table "pieces" to new temporary table "new_pieces"
INSERT INTO `new_pieces` (`id`, `title`, `description`, `composer`, `recording_link`, `measures`, `beats_per_measure`, `goal_tempo`, `user_id`) SELECT `id`, `title`, `description`, IFNULL(`composer`, '') AS `composer`, IFNULL(`recording_link`, '') AS `recording_link`, IFNULL(`measures`, 0) AS `measures`, IFNULL(`beats_per_measure`, 0) AS `beats_per_measure`, `goal_tempo`, `user_id` FROM `pieces`;
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
