-- Create "user_config" table
CREATE TABLE `user_config` (
  `user_id` text NOT NULL,
  `time_between_breaks` int NOT NULL DEFAULT 33,
  `default_plan_intensity` text NOT NULL DEFAULT 'medium',
  PRIMARY KEY (`user_id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (default_plan_intensity IN ('light', 'medium', 'heavy'))
);

INSERT INTO user_config (user_id) SELECT id FROM users;
