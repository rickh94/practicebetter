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
  `priority` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (stage IN ('repeat', 'extra_repeat', 'random', 'interleave', 'interleave_days', 'completed')),
  CHECK (LENGTH(name) > 0),
  CHECK (priority > -3),
  CHECK (priority < 3)
);
-- Copy most rows from old table "spots" to new temporary table "new_spots"
INSERT INTO `new_spots` (`id`, `piece_id`, `name`, `idx`, `stage`, `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced`, `priority`) SELECT `id`, `piece_id`, `name`, `idx`, `stage`, `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced`, `priority` FROM `spots` WHERE spots.stage != 'more_repeat';
-- Fixup some rows from old table "spots" to new temporary table "new_spots"
INSERT INTO `new_spots` (`id`, `piece_id`, `name`, `idx`, `stage`, `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced`, `priority`) SELECT `id`, `piece_id`, `name`, `idx`, 'extra_repeat', `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced`, `priority` FROM `spots` WHERE spots.stage == 'more_repeat';
-- Drop "spots" table after copying rows
DROP TABLE `spots`;
-- Rename temporary table "new_spots" to "spots"
ALTER TABLE `new_spots` RENAME TO `spots`;
-- Create index "spots_piece_id" to table: "spots"
CREATE INDEX `spots_piece_id` ON `spots` (`piece_id`);
-- Create "new_practice_plan_spots" table
CREATE TABLE `new_practice_plan_spots` (
  `practice_plan_id` text NOT NULL,
  `spot_id` text NOT NULL,
  `practice_type` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  PRIMARY KEY (`practice_plan_id`, `spot_id`),
  CONSTRAINT `0` FOREIGN KEY (`spot_id`) REFERENCES `spots` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (practice_type IN ('new', 'extra_repeat', 'interleave', 'interleave_days'))
);
-- Copy most rows from old table "practice_plan_spots" to new temporary table "new_practice_plan_spots"
INSERT INTO `new_practice_plan_spots` (`practice_plan_id`, `spot_id`, `practice_type`, `completed`) SELECT `practice_plan_id`, `spot_id`, `practice_type`, `completed` FROM `practice_plan_spots` WHERE practice_type != 'more_repeat';
-- Fixup rows from old table "practice_plan_spots" to new temporary table "new_practice_plan_spots"
INSERT INTO `new_practice_plan_spots` (`practice_plan_id`, `spot_id`, `practice_type`, `completed`) SELECT `practice_plan_id`, `spot_id`, 'extra_repeat', `completed` FROM `practice_plan_spots` WHERE practice_type == 'more_repeat';
-- Drop "practice_plan_spots" table after copying rows
DROP TABLE `practice_plan_spots`;
-- Rename temporary table "new_practice_plan_spots" to "practice_plan_spots"
ALTER TABLE `new_practice_plan_spots` RENAME TO `practice_plan_spots`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
