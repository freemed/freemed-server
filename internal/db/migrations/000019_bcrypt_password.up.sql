ALTER TABLE user ADD COLUMN userpassword_bcrypt VARCHAR(255) NOT NULL DEFAULT '' AFTER userpassword;
