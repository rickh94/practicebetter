-- Create "practice_plan_scales" table
CREATE TABLE `practice_plan_scales` (
  `practice_plan_id` text NOT NULL,
  `user_scale_id` text NOT NULL,
  `completed` boolean NOT NULL DEFAULT 0,
  `idx` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`practice_plan_id`, `user_scale_id`),
  CONSTRAINT `0` FOREIGN KEY (`user_scale_id`) REFERENCES `user_scales` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`practice_plan_id`) REFERENCES `practice_plans` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
