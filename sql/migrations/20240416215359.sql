-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_scale_modes" table
CREATE TABLE `new_scale_modes` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `basic` boolean NOT NULL,
  `cof` integer NOT NULL DEFAULT 0,
  CHECK (name IN ("Major (Ionian)", "Minor (Aeolian)", "Dorian", "Phrygian", "Lydian", "Mixolydian", "Locrian"))
);
-- Copy rows from old table "scale_modes" to new temporary table "new_scale_modes"
INSERT INTO `new_scale_modes` (`id`, `name`, `basic`) SELECT `id`, `name`, `basic` FROM `scale_modes`;
-- Drop "scale_modes" table after copying rows
DROP TABLE `scale_modes`;
-- Rename temporary table "new_scale_modes" to "scale_modes"
ALTER TABLE `new_scale_modes` RENAME TO `scale_modes`;
-- Create index "scale_modes_name" to table: "scale_modes"
CREATE UNIQUE INDEX `scale_modes_name` ON `scale_modes` (`name`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
