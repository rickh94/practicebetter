-- Create "users" table
CREATE TABLE `users` (
  `id` text NOT NULL,
  `fullname` text NOT NULL DEFAULT '',
  `email` text NOT NULL,
  `email_verified` boolean NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  CHECK (email_verified IN (0, 1))
);
-- Create index "users_email" to table: "users"
CREATE UNIQUE INDEX `users_email` ON `users` (`email`);
-- Create "credentials" table
CREATE TABLE `credentials` (
  `credential_id` blob NOT NULL,
  `public_key` blob NOT NULL,
  `transport` blob NOT NULL,
  `attestation_type` text NOT NULL,
  `flags` blob NOT NULL,
  `authenticator` blob NOT NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`credential_id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "credentials_credential_id" to table: "credentials"
CREATE UNIQUE INDEX `credentials_credential_id` ON `credentials` (`credential_id`);
