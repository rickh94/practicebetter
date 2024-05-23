-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_scale_keys" table
CREATE TABLE `new_scale_keys` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `cof` integer NOT NULL DEFAULT 1
);
-- Copy rows from old table "scale_keys" to new temporary table "new_scale_keys"
INSERT INTO `new_scale_keys` (`id`, `name`) SELECT `id`, `name` FROM `scale_keys`;
-- Drop "scale_keys" table after copying rows
DROP TABLE `scale_keys`;
-- Rename temporary table "new_scale_keys" to "scale_keys"
ALTER TABLE `new_scale_keys` RENAME TO `scale_keys`;
-- Create index "scale_keys_name" to table: "scale_keys"
CREATE UNIQUE INDEX `scale_keys_name` ON `scale_keys` (`name`);
-- Create "new_scale_modes" table
CREATE TABLE `new_scale_modes` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `basic` boolean NOT NULL,
  `cof` integer NOT NULL DEFAULT 1,
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
-- Create "new_scales" table
CREATE TABLE `new_scales` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `key_id` integer NOT NULL,
  `mode_id` integer NOT NULL,
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

UPDATE scale_keys SET cof = 0 WHERE name = "C";
UPDATE scale_keys SET cof = 1 WHERE name = "G";
UPDATE scale_keys SET cof = 2 WHERE name = "D";
UPDATE scale_keys SET cof = 3 WHERE name = "A";
UPDATE scale_keys SET cof = 4 WHERE name = "E";
UPDATE scale_keys SET cof = 5 WHERE name = "B";
UPDATE scale_keys SET cof = 6 WHERE name = "F♯/G♭";
UPDATE scale_keys SET cof = 7 WHERE name = "C♯/D♭";
UPDATE scale_keys SET cof = 8 WHERE name = "G♯/A♭";
UPDATE scale_keys SET cof = 9 WHERE name = "D♯/E♭";
UPDATE scale_keys SET cof = 10 WHERE name = "A♯/B♭";
UPDATE scale_keys SET cof = 11 WHERE name = "F";

UPDATE scale_modes SET cof = 0 WHERE name = "Major (Ionian)";
UPDATE scale_modes SET cof = 1 WHERE name = "Dorian";
UPDATE scale_modes SET cof = 2 WHERE name = "Phrygian";
UPDATE scale_modes SET cof = 3 WHERE name = "Lydian";
UPDATE scale_modes SET cof = 4 WHERE name = "Mixolydian";
UPDATE scale_modes SET cof = 5 WHERE name = "Minor (Aeolian)";
UPDATE scale_modes SET cof = 6 WHERE name = "Locrian";
