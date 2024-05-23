-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_spots" table
CREATE TABLE `new_spots` (
  `id` text NOT NULL,
  `piece_id` text NOT NULL,
  `name` text NOT NULL,
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
  `section_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`section_id`) REFERENCES `sections` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `1` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (stage IN ('repeat', 'extra_repeat', 'random', 'interleave', 'interleave_days', 'completed')),
  CHECK (LENGTH(name) > 0),
  CHECK (priority > -3),
  CHECK (priority < 3),
  CHECK (skip_days > 0)
);
-- Copy rows from old table "spots" to new temporary table "new_spots"
INSERT INTO `new_spots` (`id`, `piece_id`, `name`, `stage`, `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced`, `stage_started`, `skip_days`, `priority`) SELECT `id`, `piece_id`, `name`, `stage`, `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced`, `stage_started`, `skip_days`, `priority` FROM `spots`;
-- Drop "spots" table after copying rows
DROP TABLE `spots`;
-- Rename temporary table "new_spots" to "spots"
ALTER TABLE `new_spots` RENAME TO `spots`;
-- Create index "spots_piece_id" to table: "spots"
CREATE INDEX `spots_piece_id` ON `spots` (`piece_id`);
-- Create index "spots_piece_stage" to table: "spots"
CREATE INDEX `spots_piece_stage` ON `spots` (`piece_id`, `stage`);
-- Create index "spots_name" to table: "spots"
CREATE INDEX `spots_name` ON `spots` (`name`);
-- Create "new_scale_modes" table
CREATE TABLE `new_scale_modes` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `basic` boolean NOT NULL,
  CHECK (name IN ("Major (Ionian)", "Minor (Aeolian)", "Dorian", "Phrygian", "Lydian", "Mixolydian", "Locrian"))
);
-- Copy rows from old table "scale_modes" to new temporary table "new_scale_modes"
INSERT INTO `new_scale_modes` (`id`, `name`, `basic`) SELECT `id`, `name`, `basic` FROM `scale_modes`;
-- Drop "scale_modes" table after copying rows
DROP TABLE `scale_modes`;
-- Rename temporary table "new_scale_modes" to "scale_modes"
ALTER TABLE `new_scale_modes` RENAME TO `scale_modes`;
-- Create index "scale_modes_name" to table: "scale_modes"
CREATE UNIQUE INDEX `scale_modes_name` ON `scale_modes` (`name`);
-- Create "sections" table
CREATE TABLE `sections` (
  `id` text NOT NULL,
  `name` text NOT NULL,
  `description` text NOT NULL,
  `piece_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
