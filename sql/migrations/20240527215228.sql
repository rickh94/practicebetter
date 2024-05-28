-- Create "reading" table
CREATE TABLE `reading` (
  `id` text NOT NULL,
  `title` text NOT NULL,
  `info` text NULL,
  `composer` text NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CHECK (LENGTH(title) > 0)
);
