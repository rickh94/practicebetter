-- Create "practice_plans" table
CREATE TABLE `practice_plans` (
  `id` text NOT NULL,
  `user_id` text NOT NULL,
  `intensity` text NOT NULL,
  `date` integer NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (intensity IN ('light', 'medium', 'heavy'))
);
-- Create "practice_plan_spots" table
CREATE TABLE `practice_plan_spots` (
  `plan_id` text NOT NULL,
  `spot_id` text NOT NULL,
  `practice_type` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  PRIMARY KEY (`plan_id`, `spot_id`),
  CONSTRAINT `0` FOREIGN KEY (`spot_id`) REFERENCES `spots` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (practice_type IN ('new', 'extra_repeat', 'interleave', 'interleave_days'))
);
-- Create "practice_plan_pieces" table
CREATE TABLE `practice_plan_pieces` (
  `plan_id` text NOT NULL,
  `piece_id` text NOT NULL,
  `practice_type` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  PRIMARY KEY (`plan_id`, `piece_id`),
  CONSTRAINT `0` FOREIGN KEY (`piece_id`) REFERENCES `pieces` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (practice_type IN ('random_spots', 'starting_point'))
);
