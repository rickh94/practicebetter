-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_spots" table
CREATE TABLE `new_spots` (
  `id` text NOT NULL,
  `piece_id` text NOT NULL,
  `name` text NOT NULL,
  `idx` integer NOT NULL,
  `stage` text NOT NULL DEFAULT 'repeat',
  `audio_prompt_url` text NULL,
  `image_prompt_url` text NULL,
  `notes_prompt` text NULL,
  `text_prompt` text NULL,
  `current_tempo` integer NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (stage IN ('repeat', 'random', 'interleave', 'interleave_days', 'complete')),
  CHECK (current_tempo IS NULL OR current_tempo > 0),
  CHECK (idx > 0),
  CHECK (LENGTH(name) > 0)
);
-- Copy rows from old table "spots" to new temporary table "new_spots"
INSERT INTO `new_spots` (`id`, `piece_id`, `name`, `idx`, `stage`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`) SELECT `id`, `piece_id`, `name`, `idx`, `stage`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo` FROM `spots`;
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
  `composer` text NULL,
  `recording_link` text NULL,
  `measures` integer NULL,
  `beats_per_measure` integer NULL,
  `goal_tempo` integer NULL,
  `user_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (LENGTH(title) > 0),
  CHECK (measures is NULL OR measures > 0),
  CHECK (beats_per_measure is NULL OR beats_per_measure > 0),
  CHECK (goal_tempo is NULL OR goal_tempo > 0)
);
-- Copy rows from old table "pieces" to new temporary table "new_pieces"
INSERT INTO `new_pieces` (`id`, `title`, `description`, `composer`, `recording_link`, `measures`, `beats_per_measure`, `goal_tempo`, `user_id`) SELECT `id`, `title`, `description`, `composer`, `recording_link`, `measures`, `beats_per_measure`, `goal_tempo`, `user_id` FROM `pieces`;
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
