ALTER TABLE `pharmacy`
  ADD COLUMN `ncpdp_id` VARCHAR(10) NOT NULL DEFAULT '' AFTER `phname`,
  ADD COLUMN `service_level` VARCHAR(20) NOT NULL DEFAULT '' AFTER `ncpdp_id`,
  ADD COLUMN `fax_number` VARCHAR(20) NOT NULL DEFAULT '' AFTER `service_level`,
  ADD COLUMN `email` VARCHAR(255) NOT NULL DEFAULT '' AFTER `fax_number`;
