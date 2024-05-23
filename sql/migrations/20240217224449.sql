-- Create "user_scales" table
CREATE TABLE `user_scales` (
  `id` text NOT NULL,
  `user_id` text NOT NULL,
  `scale_id` integer NOT NULL,
  `practice_notes` text NOT NULL DEFAULT '',
  `last_practiced` integer NULL,
  `reference` text NOT NULL DEFAULT '',
  `working` boolean NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`scale_id`) REFERENCES `scales` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
