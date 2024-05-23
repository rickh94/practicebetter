-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_scales" table
CREATE TABLE `new_scales` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `key_id` integer NOT NULL,
  `mode_id` integer NOT NULL,
  `icon_name` text NOT NULL DEFAULT '',
  CONSTRAINT `0` FOREIGN KEY (`mode_id`) REFERENCES `scale_modes` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`key_id`) REFERENCES `scale_keys` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Copy rows from old table "scales" to new temporary table "new_scales"
INSERT INTO `new_scales` (`id`, `key_id`, `mode_id`) SELECT `id`, `key_id`, `mode_id` FROM `scales`;
-- Drop "scales" table after copying rows
DROP TABLE `scales`;
-- Rename temporary table "new_scales" to "scales"
ALTER TABLE `new_scales` RENAME TO `scales`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
