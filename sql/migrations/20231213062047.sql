-- Create "practice_plan_sessions" table
CREATE TABLE `practice_plan_sessions` (
  `plan_id` text NOT NULL,
  `session_id` text NOT NULL,
  PRIMARY KEY (`plan_id`, `session_id`),
  CONSTRAINT `0` FOREIGN KEY (`session_id`) REFERENCES `practice_sessions` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
