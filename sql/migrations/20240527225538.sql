-- Create "practice_plan_reading" table
CREATE TABLE `practice_plan_reading` (
  `practice_plan_id` text NOT NULL,
  `reading_id` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  `idx` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`practice_plan_id`, `reading_id`),
  CONSTRAINT `0` FOREIGN KEY (`reading_id`) REFERENCES `reading` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
