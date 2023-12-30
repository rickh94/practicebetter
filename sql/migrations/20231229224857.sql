-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_spots" table
CREATE TABLE `new_spots` (
  `id` text NOT NULL,
  `piece_id` text NOT NULL,
  `name` text NOT NULL,
  `idx` integer NOT NULL,
  `stage` text NOT NULL DEFAULT 'repeat',
  `measures` text NULL,
  `audio_prompt_url` text NOT NULL DEFAULT '',
  `image_prompt_url` text NOT NULL DEFAULT '',
  `notes_prompt` text NOT NULL DEFAULT '',
  `text_prompt` text NOT NULL DEFAULT '',
  `current_tempo` integer NULL,
  `last_practiced` integer NULL,
  `stage_started` integer NULL,
  `skip_days` integer NOT NULL DEFAULT 1,
  `priority` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (stage IN ('repeat', 'extra_repeat', 'random', 'interleave', 'interleave_days', 'completed')),
  CHECK (LENGTH(name) > 0),
  CHECK (priority > -3),
  CHECK (priority < 3)
);
-- Copy rows from old table "spots" to new temporary table "new_spots"
INSERT INTO `new_spots` (`id`, `piece_id`, `name`, `idx`, `stage`, `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced`, `stage_started`, `skip_days`, `priority`) SELECT `id`, `piece_id`, `name`, `idx`, `stage`, `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced`, `stage_started`, IFNULL(`skip_days`, 1) AS `skip_days`, `priority` FROM `spots`;
-- Drop "spots" table after copying rows
DROP TABLE `spots`;
-- Rename temporary table "new_spots" to "spots"
ALTER TABLE `new_spots` RENAME TO `spots`;
-- Create index "spots_piece_id" to table: "spots"
CREATE INDEX `spots_piece_id` ON `spots` (`piece_id`);
-- Create "new_practice_plan_pieces" table
CREATE TABLE `new_practice_plan_pieces` (
  `practice_plan_id` text NOT NULL,
  `piece_id` text NOT NULL,
  `practice_type` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  `sessions` integer NOT NULL DEFAULT 1,
  `sessionse_completed` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`practice_plan_id`, `piece_id`, `practice_type`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (practice_type IN ('random_spots', 'starting_point'))
);
-- Copy rows from old table "practice_plan_pieces" to new temporary table "new_practice_plan_pieces"
INSERT INTO `new_practice_plan_pieces` (`practice_plan_id`, `piece_id`, `practice_type`, `completed`) SELECT `practice_plan_id`, `piece_id`, `practice_type`, `completed` FROM `practice_plan_pieces`;
-- Drop "practice_plan_pieces" table after copying rows
DROP TABLE `practice_plan_pieces`;
-- Rename temporary table "new_practice_plan_pieces" to "practice_plan_pieces"
ALTER TABLE `new_practice_plan_pieces` RENAME TO `practice_plan_pieces`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
