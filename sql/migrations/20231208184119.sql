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
  CHECK (stage IN ('repeat', 'random', 'interleave', 'interleave_days', 'completed')),
  CHECK (LENGTH(name) > 0)
);
-- Copy rows from old table "spots" to new temporary table "new_spots"
INSERT INTO `new_spots` (`id`, `piece_id`, `name`, `idx`, `stage`, `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced`) SELECT `id`, `piece_id`, `name`, `idx`, `stage`, `measures`, `audio_prompt_url`, `image_prompt_url`, `notes_prompt`, `text_prompt`, `current_tempo`, `last_practiced` FROM `spots`;
-- Drop "spots" table after copying rows
DROP TABLE `spots`;
-- Rename temporary table "new_spots" to "spots"
ALTER TABLE `new_spots` RENAME TO `spots`;
-- Create index "spots_piece_id" to table: "spots"
CREATE INDEX `spots_piece_id` ON `spots` (`piece_id`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
