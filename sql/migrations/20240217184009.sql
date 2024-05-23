-- Create "scale_keys" table
CREATE TABLE `scale_keys` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL
);
-- Create index "scale_keys_name" to table: "scale_keys"
CREATE UNIQUE INDEX `scale_keys_name` ON `scale_keys` (`name`);
-- Create "scale_modes" table
CREATE TABLE `scale_modes` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `basic` boolean NOT NULL,
  CHECK (name IN ("Major (Ionian)", "Minor (Aeolian)", "Dorian", "Phrygian", "Lydian", "Mixolydian", "Locrian"))
);
-- Create index "scale_modes_name" to table: "scale_modes"
CREATE UNIQUE INDEX `scale_modes_name` ON `scale_modes` (`name`);
-- Create "scales" table
CREATE TABLE `scales` (
  `key_id` integer NOT NULL,
  `mode_id` integer NOT NULL,
  PRIMARY KEY (`key_id`, `mode_id`),
  CONSTRAINT `0` FOREIGN KEY (`mode_id`) REFERENCES `scale_modes` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`key_id`) REFERENCES `scale_keys` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);

INSERT INTO scale_keys (
    name
) VALUES
("G♯/A♭"),
("A"),
("A♯/B♭"),
("B"),
("C"),
("C♯/D♭"),
("D"),
("D♯/E♭"),
("E"),
("F"),
("F♯/G♭"),
("G");

INSERT INTO scale_modes (
    name,
    basic
) VALUES
("Major (Ionian)", TRUE),
("Dorian", FALSE),
("Phrygian", FALSE),
("Lydian", FALSE),
("Mixolydian", FALSE),
("Minor (Aeolian)", TRUE),
("Locrian", FALSE);

INSERT INTO scales (
    key_id,
    mode_id
) SELECT scale_keys.id, scale_modes.id
FROM scale_keys
CROSS JOIN scale_modes;
