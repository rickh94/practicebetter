-- Create "scales" table
CREATE TABLE `scales` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `key_id` integer NOT NULL,
  `mode_id` integer NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`mode_id`) REFERENCES `scale_modes` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`key_id`) REFERENCES `scale_keys` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "user_scales" table
CREATE TABLE `user_scales` (
  `user_id` text NOT NULL,
  `scale_id` integer NOT NULL,
  `practice_notes` text NOT NULL DEFAULT '',
  `last_practiced` integer NULL,
  `reference` text NOT NULL DEFAULT '',
  PRIMARY KEY (`user_id`, `scale_id`),
  CONSTRAINT `0` FOREIGN KEY (`scale_id`) REFERENCES `scales` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);

INSERT INTO scales (
    key_id,
    mode_id
) SELECT scale_keys.id, scale_modes.id
FROM scale_keys
CROSS JOIN scale_modes;
