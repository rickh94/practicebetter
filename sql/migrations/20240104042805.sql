-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_practice_plan_spots" table
CREATE TABLE `new_practice_plan_spots` (
  `practice_plan_id` text NOT NULL,
  `spot_id` text NOT NULL,
  `practice_type` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  `idx` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`practice_plan_id`, `spot_id`),
  CONSTRAINT `0` FOREIGN KEY (`spot_id`) REFERENCES `spots` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (practice_type IN ('new', 'extra_repeat', 'interleave', 'interleave_days'))
);
-- Copy rows from old table "practice_plan_spots" to new temporary table "new_practice_plan_spots"
INSERT INTO `new_practice_plan_spots` (`practice_plan_id`, `spot_id`, `practice_type`, `completed`) SELECT `practice_plan_id`, `spot_id`, `practice_type`, `completed` FROM `practice_plan_spots`;
-- Drop "practice_plan_spots" table after copying rows
DROP TABLE `practice_plan_spots`;
-- Rename temporary table "new_practice_plan_spots" to "practice_plan_spots"
ALTER TABLE `new_practice_plan_spots` RENAME TO `practice_plan_spots`;
-- Create "new_practice_plan_pieces" table
CREATE TABLE `new_practice_plan_pieces` (
  `practice_plan_id` text NOT NULL,
  `piece_id` text NOT NULL,
  `practice_type` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  `sessions` integer NOT NULL DEFAULT 1,
  `idx` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`practice_plan_id`, `piece_id`, `practice_type`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (practice_type IN ('random_spots', 'starting_point'))
);
-- Copy rows from old table "practice_plan_pieces" to new temporary table "new_practice_plan_pieces"
INSERT INTO `new_practice_plan_pieces` (`practice_plan_id`, `piece_id`, `practice_type`, `completed`, `sessions`) SELECT `practice_plan_id`, `piece_id`, `practice_type`, `completed`, `sessions` FROM `practice_plan_pieces`;
-- Drop "practice_plan_pieces" table after copying rows
DROP TABLE `practice_plan_pieces`;
-- Rename temporary table "new_practice_plan_pieces" to "practice_plan_pieces"
ALTER TABLE `new_practice_plan_pieces` RENAME TO `practice_plan_pieces`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
