-- Create "pieces" table
CREATE TABLE `pieces` (
  `id` text NOT NULL,
  `title` text NOT NULL,
  `description` text NULL,
  `composer` text NULL,
  `recording_link` text NULL,
  `measures` integer NULL,
  `beats_per_measure` integer NULL,
  `goalTempo` integer NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "pieces_user_id" to table: "pieces"
CREATE INDEX `pieces_user_id` ON `pieces` (`user_id`);
-- Create index "pieces_user_id_title" to table: "pieces"
CREATE INDEX `pieces_user_id_title` ON `pieces` (`user_id`, `title`);
-- Create "spots" table
CREATE TABLE `spots` (
  `id` text NOT NULL,
  `piece_id` text NOT NULL,
  `name` text NOT NULL,
  `idx` integer NULL,
  `stage` text NOT NULL DEFAULT 'repeat',
  `audio_prompt_url` text NULL,
  `image_prompt_url` text NULL,
  `notes_prompt` text NULL,
  `text_prompt` text NULL,
  `current_tempo` integer NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (stage IN ('repeat', 'random', 'interleave', 'interleave_days', 'complete'))
);
-- Create index "spots_piece_id" to table: "spots"
CREATE INDEX `spots_piece_id` ON `spots` (`piece_id`);
