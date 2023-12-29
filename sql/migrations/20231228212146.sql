-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_practice_plan_pieces" table
CREATE TABLE `new_practice_plan_pieces` (
  `practice_plan_id` text NOT NULL,
  `piece_id` text NOT NULL,
  `practice_type` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
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
