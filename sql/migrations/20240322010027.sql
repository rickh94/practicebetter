-- Create "spots_sections" table
CREATE TABLE `spots_sections` (
  `spot_id` text NOT NULL,
  `section_id` text NOT NULL,
  PRIMARY KEY (`spot_id`, `section_id`),
  CONSTRAINT `0` FOREIGN KEY (`section_id`) REFERENCES `sections` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`spot_id`) REFERENCES `spots` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
