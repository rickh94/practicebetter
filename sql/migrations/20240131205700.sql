-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_practice_plan_spots" table
CREATE TABLE `new_practice_plan_spots` (
  `practice_plan_id` text NOT NULL,
  `spot_id` text NOT NULL,
  `practice_type` text NOT NULL,
  `evaluation` text NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  `idx` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`practice_plan_id`, `spot_id`),
  CONSTRAINT `0` FOREIGN KEY (`spot_id`) REFERENCES `spots` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (practice_type IN ('new', 'extra_repeat', 'interleave', 'interleave_days')),
  CHECK (evaluation = NULL OR evaluation IN ('poor', 'fine', 'excellent'))
);
-- Copy rows from old table "practice_plan_spots" to new temporary table "new_practice_plan_spots"
INSERT INTO `new_practice_plan_spots` (`practice_plan_id`, `spot_id`, `practice_type`, `completed`, `idx`) SELECT `practice_plan_id`, `spot_id`, `practice_type`, `completed`, `idx` FROM `practice_plan_spots`;
-- Drop "practice_plan_spots" table after copying rows
DROP TABLE `practice_plan_spots`;
-- Rename temporary table "new_practice_plan_spots" to "practice_plan_spots"
ALTER TABLE `new_practice_plan_spots` RENAME TO `practice_plan_spots`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
