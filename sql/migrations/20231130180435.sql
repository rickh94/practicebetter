-- Create "practice_sessions" table
CREATE TABLE `practice_sessions` (
  `id` text NOT NULL,
  `duration_minutes` integer NOT NULL,
  `date` integer NOT NULL,
  PRIMARY KEY (`id`)
);
-- Create index "practice_sessions_date" to table: "practice_sessions"
CREATE INDEX `practice_sessions_date` ON `practice_sessions` (`date`);
-- Create "practice_piece" table
CREATE TABLE `practice_piece` (
  `practice_session_id` text NOT NULL,
  `piece_id` text NOT NULL,
  `measures` text NOT NULL,
  PRIMARY KEY (`practice_session_id`, `piece_id`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_session_id`) REFERENCES `practice_sessions` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "practice_spot" table
CREATE TABLE `practice_spot` (
  `practice_session_id` text NOT NULL,
  `spot_id` text NOT NULL,
  PRIMARY KEY (`practice_session_id`, `spot_id`),
  CONSTRAINT `0` FOREIGN KEY (`spot_id`) REFERENCES `spots` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_session_id`) REFERENCES `practice_sessions` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
