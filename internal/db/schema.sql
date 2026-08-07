-- Auto-generated from GORM model structs
-- 51 tables

CREATE TABLE `annotations` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `atimestamp` DATETIME NOT NULL,
  `apatient` BIGINT NOT NULL DEFAULT 0,
  `amodule` VARCHAR(255) NOT NULL DEFAULT '',
  `atable` VARCHAR(255) NOT NULL DEFAULT '',
  `aid` BIGINT NOT NULL DEFAULT 0,
  `auser` BIGINT NOT NULL DEFAULT 0,
  `annotation` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `appttemplate` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `atname` VARCHAR(255) NOT NULL DEFAULT '',
  `atduration` BIGINT NOT NULL DEFAULT 0,
  `atequipment` LONGBLOB,
  `atcolor` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `authorizations` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `authdtadd` DATETIME NOT NULL,
  `authdtmod` DATETIME NOT NULL,
  `authpatient` BIGINT NOT NULL DEFAULT 0,
  `authdtbegin` DATETIME NOT NULL,
  `authdtend` DATETIME NOT NULL,
  `authnum` VARCHAR(255) NOT NULL DEFAULT '',
  `authtype` BIGINT NOT NULL DEFAULT 0,
  `authprov` BIGINT NOT NULL DEFAULT 0,
  `authprovid` VARCHAR(255) NOT NULL DEFAULT '',
  `authinsco` BIGINT NOT NULL DEFAULT 0,
  `authvisits` BIGINT NOT NULL DEFAULT 0,
  `authvisitsused` BIGINT NOT NULL DEFAULT 0,
  `authvisitsremain` BIGINT NOT NULL DEFAULT 0,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `bccdc` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `agent_code` VARCHAR(255) NOT NULL DEFAULT '',
  `description` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `bcontact` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `bcfname` VARCHAR(255) NOT NULL DEFAULT '',
  `bcmname` VARCHAR(255) NOT NULL DEFAULT '',
  `bclname` VARCHAR(255) NOT NULL DEFAULT '',
  `bcaddr` VARCHAR(255) NOT NULL DEFAULT '',
  `bccity` VARCHAR(255) NOT NULL DEFAULT '',
  `bcstate` VARCHAR(255) NOT NULL DEFAULT '',
  `bczip` VARCHAR(255) NOT NULL DEFAULT '',
  `bcphone` VARCHAR(255) NOT NULL DEFAULT '',
  `stamp` DATETIME NOT NULL,
  `user` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `billkey` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `billkeydate` DATETIME NOT NULL,
  `billkey` LONGBLOB,
  `bkprocs` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `bodysite` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `abbrev` VARCHAR(255) NOT NULL DEFAULT '',
  `display_value` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `bservice` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `bsname` VARCHAR(255) NOT NULL DEFAULT '',
  `bsaddr` VARCHAR(255) NOT NULL DEFAULT '',
  `bscity` VARCHAR(255) NOT NULL DEFAULT '',
  `bsstate` VARCHAR(255) NOT NULL DEFAULT '',
  `bszip` VARCHAR(255) NOT NULL DEFAULT '',
  `bsphone` VARCHAR(255) NOT NULL DEFAULT '',
  `bsetin` VARCHAR(255) NOT NULL DEFAULT '',
  `bstin` VARCHAR(255) NOT NULL DEFAULT '',
  `stamp` DATETIME NOT NULL,
  `user` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `calgroup` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `groupname` VARCHAR(255) NOT NULL DEFAULT '',
  `groupfacility` BIGINT NOT NULL DEFAULT 0,
  `groupfrequency` BIGINT NOT NULL DEFAULT 0,
  `grouplength` BIGINT NOT NULL DEFAULT 0,
  `groupmembers` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `calgroupattend` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `calgroupid` BIGINT NOT NULL DEFAULT 0,
  `calid` BIGINT NOT NULL DEFAULT 0,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `calstatus` VARCHAR(255) NOT NULL DEFAULT '',
  `stamp` DATETIME NOT NULL
);

CREATE TABLE `claimlog` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `cltimestamp` DATETIME NOT NULL,
  `cluser` BIGINT NOT NULL DEFAULT 0,
  `clprocedure` BIGINT NOT NULL DEFAULT 0,
  `clpayrec` BIGINT NOT NULL DEFAULT 0,
  `claction` VARCHAR(255) NOT NULL DEFAULT '',
  `clcomment` VARCHAR(255) NOT NULL DEFAULT '',
  `clformat` VARCHAR(255) NOT NULL DEFAULT '',
  `cltarget` VARCHAR(255) NOT NULL DEFAULT '',
  `cltargetopt` VARCHAR(255) NOT NULL DEFAULT '',
  `clbillkey` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `claimtype` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `clmtpname` VARCHAR(255) NOT NULL DEFAULT '',
  `clmtpdescrip` VARCHAR(255) NOT NULL DEFAULT '',
  `clmtpadd` DATETIME NOT NULL,
  `clmtpmod` DATETIME NOT NULL
);

CREATE TABLE `clearinghouse` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `chname` VARCHAR(255) NOT NULL DEFAULT '',
  `chaddr` VARCHAR(255) NOT NULL DEFAULT '',
  `chcity` VARCHAR(255) NOT NULL DEFAULT '',
  `chstate` VARCHAR(255) NOT NULL DEFAULT '',
  `chzip` VARCHAR(255) NOT NULL DEFAULT '',
  `chphone` VARCHAR(255) NOT NULL DEFAULT '',
  `chetin` VARCHAR(255) NOT NULL DEFAULT '',
  `chx12gssender` VARCHAR(255) NOT NULL DEFAULT '',
  `chx12gsreceiver` VARCHAR(255) NOT NULL DEFAULT '',
  `stamp` DATETIME NOT NULL,
  `user` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `config` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `c_option` VARCHAR(255) NOT NULL DEFAULT '',
  `c_value` VARCHAR(255),
  `c_title` VARCHAR(255),
  `c_section` VARCHAR(255),
  `c_type` VARCHAR(255) NOT NULL DEFAULT '',
  `c_options` VARCHAR(255)
);

CREATE TABLE `covtypes` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `covtpname` VARCHAR(255) NOT NULL DEFAULT '',
  `covtpdescrip` VARCHAR(255) NOT NULL DEFAULT '',
  `covtpdtadd` DATETIME NOT NULL,
  `covtpdtmod` DATETIME NOT NULL
);

CREATE TABLE `cpt` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `abbrev` VARCHAR(255) NOT NULL DEFAULT '',
  `cptnameint` VARCHAR(255),
  `cptnameext` VARCHAR(255),
  `cptgender` VARCHAR(255) NOT NULL DEFAULT '',
  `cpttaxed` VARCHAR(255) NOT NULL DEFAULT '',
  `cpttype` BIGINT NOT NULL DEFAULT 0,
  `cptreqcpt` VARCHAR(255),
  `cptexccpt` VARCHAR(255),
  `cptreqicd` VARCHAR(255),
  `cptrexcicd` VARCHAR(255),
  `cptrelval` DOUBLE NOT NULL DEFAULT 0,
  `cptdeftos` BIGINT NOT NULL DEFAULT 0,
  `cptdefstdfee` DOUBLE NOT NULL DEFAULT 0,
  `cptstdfee` VARCHAR(255) NOT NULL DEFAULT '',
  `cpttos` VARCHAR(255) NOT NULL DEFAULT '',
  `cpttosprfx` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `cptmod` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `cptmod` VARCHAR(255) NOT NULL DEFAULT '',
  `cptmoddescrip` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `documents_tc` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `type` VARCHAR(255) NOT NULL DEFAULT '',
  `category` VARCHAR(255) NOT NULL DEFAULT '',
  `description` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `drugforms` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `code` VARCHAR(255) NOT NULL DEFAULT '',
  `description` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `drugquantityqual` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `code` VARCHAR(255) NOT NULL DEFAULT '',
  `description` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `drugsampleinv` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `drugcode` VARCHAR(255) NOT NULL DEFAULT '',
  `drugndc` VARCHAR(255) NOT NULL DEFAULT '',
  `drugclass` VARCHAR(255) NOT NULL DEFAULT '',
  `packagecount` BIGINT NOT NULL DEFAULT 0,
  `location` VARCHAR(255) NOT NULL DEFAULT '',
  `drugco` VARCHAR(255) NOT NULL DEFAULT '',
  `drugrep` VARCHAR(255) NOT NULL DEFAULT '',
  `invoice` VARCHAR(255) NOT NULL DEFAULT '',
  `samplecount` BIGINT NOT NULL DEFAULT 0,
  `samplecountremain` BIGINT NOT NULL DEFAULT 0,
  `lot` VARCHAR(255) NOT NULL DEFAULT '',
  `expiration` DATETIME,
  `received` DATETIME,
  `assignedto` VARCHAR(255) NOT NULL DEFAULT '',
  `loguser` BIGINT NOT NULL DEFAULT 0,
  `logdate` DATETIME NOT NULL,
  `disposeby` VARCHAR(255) NOT NULL DEFAULT '',
  `disposedate` DATETIME,
  `disposemethod` VARCHAR(255) NOT NULL DEFAULT '',
  `disposereason` VARCHAR(255) NOT NULL DEFAULT '',
  `witness` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `enctype` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `enclosure` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `facility` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `psrname` VARCHAR(255) NOT NULL DEFAULT '',
  `psraddr1` VARCHAR(255),
  `psraddr2` VARCHAR(255),
  `psrcity` VARCHAR(255),
  `psrstate` VARCHAR(255),
  `psrzip` VARCHAR(255),
  `psrcountry` VARCHAR(255),
  `psrnote` VARCHAR(255),
  `psrphone` VARCHAR(255),
  `psrfax` VARCHAR(255),
  `psremail` VARCHAR(255),
  `psrein` VARCHAR(255),
  `psrnpi` VARCHAR(255),
  `psrtaxonomy` VARCHAR(255),
  `psrintext` VARCHAR(255),
  `psrpos` BIGINT NOT NULL DEFAULT 0,
  `psrx12id` VARCHAR(255),
  `psrx12idtype` VARCHAR(255)
);

CREATE TABLE `i18nlanguages` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `abbrev` VARCHAR(255) NOT NULL DEFAULT '',
  `language` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `icd9` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `abbrev` VARCHAR(255) NOT NULL DEFAULT '',
  `language` VARCHAR(255) NOT NULL DEFAULT '',
  `icd9code` VARCHAR(255) NOT NULL DEFAULT '',
  `icd10code` VARCHAR(255),
  `icd9descrip` VARCHAR(255) NOT NULL DEFAULT '',
  `icd10descrip` VARCHAR(255),
  `icdmetadesc` VARCHAR(255),
  `icdng` DATETIME,
  `icddrg` VARCHAR(255) NOT NULL DEFAULT '',
  `icdnum` BIGINT NOT NULL DEFAULT 0,
  `icdamt` TEXT,
  `icdcoll` DOUBLE NOT NULL DEFAULT 0
);

CREATE TABLE `immunization` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `dateof` DATETIME NOT NULL,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `provider` BIGINT NOT NULL DEFAULT 0,
  `admin_provider` BIGINT NOT NULL DEFAULT 0,
  `eoc` BIGINT,
  `immunization` BIGINT NOT NULL DEFAULT 0,
  `route` BIGINT NOT NULL DEFAULT 0,
  `body_site` BIGINT NOT NULL DEFAULT 0,
  `manufacturer` VARCHAR(255),
  `lot_number` VARCHAR(255),
  `previous_doses` BIGINT NOT NULL DEFAULT 0,
  `recovered` TINYINT(1) NOT NULL DEFAULT 0,
  `notes` VARCHAR(255),
  `orderid` BIGINT NOT NULL DEFAULT 0,
  `locked` BIGINT NOT NULL DEFAULT 0,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `insco` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `inscodtadd` DATETIME NOT NULL,
  `inscodtmod` DATETIME NOT NULL,
  `insconame` VARCHAR(255) NOT NULL DEFAULT '',
  `inscoalias` VARCHAR(255) NOT NULL DEFAULT '',
  `inscoaddr1` VARCHAR(255) NOT NULL DEFAULT '',
  `inscoaddr2` VARCHAR(255) NOT NULL DEFAULT '',
  `inscocity` VARCHAR(255) NOT NULL DEFAULT '',
  `inscostate` VARCHAR(255) NOT NULL DEFAULT '',
  `inscozip` VARCHAR(255) NOT NULL DEFAULT '',
  `inscophone` VARCHAR(255) NOT NULL DEFAULT '',
  `inscofax` VARCHAR(255) NOT NULL DEFAULT '',
  `inscogroup` BIGINT NOT NULL DEFAULT 0,
  `inscotype` BIGINT NOT NULL DEFAULT 0,
  `inscoassign` BIGINT NOT NULL DEFAULT 0,
  `inscomod` VARCHAR(255) NOT NULL DEFAULT '',
  `inscoidmap` VARCHAR(255) NOT NULL DEFAULT '',
  `inscox12id` VARCHAR(255) NOT NULL DEFAULT '',
  `inscodefformat` VARCHAR(255) NOT NULL DEFAULT '',
  `inscodeftarget` VARCHAR(255) NOT NULL DEFAULT '',
  `inscodeftargetopt` VARCHAR(255) NOT NULL DEFAULT '',
  `inscodefformate` VARCHAR(255) NOT NULL DEFAULT '',
  `inscodeftargete` VARCHAR(255) NOT NULL DEFAULT '',
  `inscodeftargetopte` VARCHAR(255) NOT NULL DEFAULT '',
  `inscoarchive` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `inscogroup` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `inscogroup` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `insmod` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `insmod` VARCHAR(255) NOT NULL DEFAULT '',
  `insmoddesc` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `intservtype` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `intservtype` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `loinc` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `loinc_num` VARCHAR(255) NOT NULL DEFAULT '',
  `component` VARCHAR(255) NOT NULL DEFAULT '',
  `property` VARCHAR(255) NOT NULL DEFAULT '',
  `type_aspct` VARCHAR(255) NOT NULL DEFAULT '',
  `system` VARCHAR(255) NOT NULL DEFAULT '',
  `scale_typ` VARCHAR(255) NOT NULL DEFAULT '',
  `method_typ` VARCHAR(255) NOT NULL DEFAULT '',
  `answerlist` VARCHAR(255) NOT NULL DEFAULT '',
  `status` VARCHAR(255) NOT NULL DEFAULT '',
  `shortname` VARCHAR(255) NOT NULL DEFAULT '',
  `external_copyright_notice` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `messages` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `msgby` BIGINT NOT NULL DEFAULT 0,
  `sender` VARCHAR(255) NOT NULL DEFAULT '',
  `msgtime` DATETIME NOT NULL,
  `msgfor` BIGINT NOT NULL DEFAULT 0,
  `msgrecip` VARCHAR(255) NOT NULL DEFAULT '',
  `msgpatient` BIGINT NOT NULL DEFAULT 0,
  `msgperson` VARCHAR(255) NOT NULL DEFAULT '',
  `msgurgency` BIGINT NOT NULL DEFAULT 0,
  `msgsubject` VARCHAR(255) NOT NULL DEFAULT '',
  `msgtext` VARCHAR(255) NOT NULL DEFAULT '',
  `msgread` BIGINT NOT NULL DEFAULT 0,
  `msgunique` VARCHAR(255),
  `msgtag` VARCHAR(255),
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `patient` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `ptdtadd` DATETIME NOT NULL,
  `ptdtmod` DATETIME,
  `ptsalut` VARCHAR(255) NOT NULL DEFAULT '',
  `ptlname` VARCHAR(255) NOT NULL DEFAULT '',
  `ptmaidenname` VARCHAR(255),
  `ptfname` VARCHAR(255) NOT NULL DEFAULT '',
  `ptmname` VARCHAR(255),
  `ptsuffix` VARCHAR(255) NOT NULL DEFAULT '',
  `ptsex` VARCHAR(255) NOT NULL DEFAULT '',
  `ptid` VARCHAR(255) NOT NULL DEFAULT '',
  `ptdiag1` BIGINT,
  `ptdiag2` BIGINT,
  `ptdiag3` BIGINT,
  `ptdiag4` BIGINT,
  `ptdiagset` VARCHAR(255) NOT NULL DEFAULT '',
  `ptarchive` BIGINT NOT NULL DEFAULT 0,
  `iso` VARCHAR(255) NOT NULL DEFAULT '',
  `ptblood` VARCHAR(255) NOT NULL DEFAULT '',
  `ptdead` BIGINT NOT NULL DEFAULT 0,
  `ptdeaddt` DATETIME,
  `ptbudg` DOUBLE NOT NULL DEFAULT 0,
  `ptbilltype` VARCHAR(255) NOT NULL DEFAULT '',
  `ptprimaryfacility` BIGINT NOT NULL DEFAULT 0,
  `ptprimarylanguage` VARCHAR(255) NOT NULL DEFAULT '',
  `ptdob` DATETIME,
  `ptpcp` BIGINT NOT NULL DEFAULT 0,
  `ptpharmacy` BIGINT NOT NULL DEFAULT 0,
  `ssn` VARCHAR(255),
  `pemail` VARCHAR(255),
  `dmv` VARCHAR(255),
  `patient` BIGINT NOT NULL DEFAULT 0,
  `module` VARCHAR(255) NOT NULL DEFAULT '',
  `oid` BIGINT NOT NULL DEFAULT 0,
  `stamp` DATETIME NOT NULL,
  `summary` VARCHAR(255) NOT NULL DEFAULT '',
  `locked` TINYINT(1) NOT NULL DEFAULT 0,
  `annotation` VARCHAR(255),
  `user` BIGINT NOT NULL DEFAULT 0,
  `provider` BIGINT NOT NULL DEFAULT 0,
  `status` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `patient_emr` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `module` VARCHAR(255) NOT NULL DEFAULT '',
  `oid` BIGINT NOT NULL DEFAULT 0,
  `stamp` DATETIME NOT NULL,
  `summary` VARCHAR(255) NOT NULL DEFAULT '',
  `locked` TINYINT(1) NOT NULL DEFAULT 0,
  `annotation` VARCHAR(255),
  `user` BIGINT NOT NULL DEFAULT 0,
  `provider` BIGINT NOT NULL DEFAULT 0,
  `language` VARCHAR(255) NOT NULL DEFAULT '',
  `status` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `patient_ids` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `foreign_id` VARCHAR(255) NOT NULL DEFAULT '',
  `facility` BIGINT NOT NULL DEFAULT 0,
  `practice` BIGINT NOT NULL DEFAULT 0,
  `user` BIGINT NOT NULL DEFAULT 0,
  `stamp` DATETIME NOT NULL,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `patienttag` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `tag` VARCHAR(255) NOT NULL DEFAULT '',
  `patient` BIGINT NOT NULL DEFAULT 0,
  `user` BIGINT NOT NULL DEFAULT 0,
  `datecreate` DATETIME NOT NULL,
  `dateexpire` DATETIME
);

CREATE TABLE `payrec` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `payrecdtadd` DATETIME,
  `payrecdtmod` DATETIME,
  `payrecpatient` BIGINT NOT NULL DEFAULT 0,
  `payreccat` BIGINT NOT NULL DEFAULT 0,
  `payrecproc` BIGINT NOT NULL DEFAULT 0,
  `payrecsource` BIGINT NOT NULL DEFAULT 0,
  `payreclink` BIGINT NOT NULL DEFAULT 0,
  `payrectype` BIGINT NOT NULL DEFAULT 0,
  `payrecnum` VARCHAR(255),
  `payrecamt` DOUBLE NOT NULL DEFAULT 0,
  `payreclock` VARCHAR(255),
  `payrecdescrip` VARCHAR(255) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `pds` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `module` VARCHAR(255) NOT NULL DEFAULT '',
  `contents` LONGBLOB
);

CREATE TABLE `physician` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `phylname` VARCHAR(255) NOT NULL DEFAULT '',
  `phyfname` VARCHAR(255) NOT NULL DEFAULT '',
  `phymname` VARCHAR(255) NOT NULL DEFAULT '',
  `phytitle` VARCHAR(255) NOT NULL DEFAULT '',
  `phypractice` BIGINT NOT NULL DEFAULT 0,
  `phypracein` VARCHAR(255) NOT NULL DEFAULT '',
  `phyaddr1a` VARCHAR(255) NOT NULL DEFAULT '',
  `phyaddr2a` VARCHAR(255) NOT NULL DEFAULT '',
  `phycitya` VARCHAR(255) NOT NULL DEFAULT '',
  `phystatea` VARCHAR(255) NOT NULL DEFAULT '',
  `phyzipa` VARCHAR(255) NOT NULL DEFAULT '',
  `phyphonea` VARCHAR(255) NOT NULL DEFAULT '',
  `phyfaxa` VARCHAR(255) NOT NULL DEFAULT '',
  `phyaddr1b` VARCHAR(255) NOT NULL DEFAULT '',
  `phyaddr2b` VARCHAR(255) NOT NULL DEFAULT '',
  `phycityb` VARCHAR(255) NOT NULL DEFAULT '',
  `phystateb` VARCHAR(255) NOT NULL DEFAULT '',
  `phyzipb` VARCHAR(255) NOT NULL DEFAULT '',
  `phyphoneb` VARCHAR(255) NOT NULL DEFAULT '',
  `phyfaxb` VARCHAR(255) NOT NULL DEFAULT '',
  `phyemail` VARCHAR(255) NOT NULL DEFAULT '',
  `phypager` VARCHAR(255) NOT NULL DEFAULT '',
  `phyupin` VARCHAR(255) NOT NULL DEFAULT '',
  `physsn` VARCHAR(255) NOT NULL DEFAULT '',
  `phydegrees` VARCHAR(255) NOT NULL DEFAULT '',
  `physpecialties` VARCHAR(255) NOT NULL DEFAULT '',
  `phyid1` VARCHAR(255) NOT NULL DEFAULT '',
  `phystatus` BIGINT NOT NULL DEFAULT 0,
  `phyref` VARCHAR(255) NOT NULL DEFAULT '',
  `phyrefcount` BIGINT NOT NULL DEFAULT 0,
  `phyrefamt` DOUBLE NOT NULL DEFAULT 0,
  `phyrefcoll` DOUBLE NOT NULL DEFAULT 0,
  `phychargemap` VARCHAR(255) NOT NULL DEFAULT '',
  `phyidmap` VARCHAR(255) NOT NULL DEFAULT '',
  `phygrpprac` BIGINT,
  `phyanesth` BIGINT NOT NULL DEFAULT 0,
  `phyhl7id` VARCHAR(255) NOT NULL DEFAULT '',
  `phydea` VARCHAR(255) NOT NULL DEFAULT '',
  `phyclia` VARCHAR(255) NOT NULL DEFAULT '',
  `phynpi` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `pnotes` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `pnotesdt` TEXT,
  `pnotesdtadd` TEXT,
  `pnotesdtmod` TEXT,
  `pnotespat` BIGINT NOT NULL DEFAULT 0,
  `pnotesdescrip` VARCHAR(255) NOT NULL DEFAULT '',
  `pnotesdoc` BIGINT NOT NULL DEFAULT 0,
  `pnoteseoc` TEXT,
  `pnotes_S` VARCHAR(255) NOT NULL DEFAULT '',
  `pnotes_O` VARCHAR(255) NOT NULL DEFAULT '',
  `pnotes_A` VARCHAR(255) NOT NULL DEFAULT '',
  `pnotes_P` VARCHAR(255) NOT NULL DEFAULT '',
  `pnotes_I` VARCHAR(255) NOT NULL DEFAULT '',
  `pnotes_E` VARCHAR(255) NOT NULL DEFAULT '',
  `pnotes_R` VARCHAR(255) NOT NULL DEFAULT '',
  `pnotessbp` TEXT,
  `pnotesdbp` TEXT,
  `pnotestemp` TEXT,
  `pnotesheartrate` TEXT,
  `pnotesresprate` TEXT,
  `weight` TEXT,
  `height` TEXT,
  `bmi` TEXT,
  `iso` TEXT,
  `locked` BIGINT NOT NULL DEFAULT 0,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `pos` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `posname` VARCHAR(255) NOT NULL DEFAULT '',
  `posdescrip` VARCHAR(255) NOT NULL DEFAULT '',
  `posdtadd` DATETIME,
  `posdtmod` DATETIME
);

CREATE TABLE `practice` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `pracname` VARCHAR(255) NOT NULL DEFAULT '',
  `pracein` VARCHAR(255),
  `addr1a` VARCHAR(255),
  `addr2a` VARCHAR(255),
  `citya` VARCHAR(255),
  `statea` VARCHAR(255),
  `zipa` VARCHAR(255),
  `phonea` VARCHAR(255),
  `faxa` VARCHAR(255),
  `addr1b` VARCHAR(255),
  `addr2b` VARCHAR(255),
  `cityb` VARCHAR(255),
  `stateb` VARCHAR(255),
  `zipb` VARCHAR(255),
  `phoneb` VARCHAR(255),
  `faxb` VARCHAR(255),
  `email` VARCHAR(255),
  `cellular` VARCHAR(255),
  `pracnpi` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `procrec` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `procpatient` BIGINT NOT NULL DEFAULT 0,
  `proceoc` VARCHAR(255),
  `proccpt` BIGINT NOT NULL DEFAULT 0,
  `proccptmod` BIGINT NOT NULL DEFAULT 0,
  `proccptmod2` BIGINT NOT NULL DEFAULT 0,
  `proccptmod3` BIGINT NOT NULL DEFAULT 0,
  `procdiag1` BIGINT NOT NULL DEFAULT 0,
  `procdiag2` BIGINT NOT NULL DEFAULT 0,
  `procdiag3` BIGINT NOT NULL DEFAULT 0,
  `procdiag4` BIGINT NOT NULL DEFAULT 0,
  `procdiagset` VARCHAR(255) NOT NULL DEFAULT '',
  `proccharges` DOUBLE NOT NULL DEFAULT 0,
  `procunits` DOUBLE NOT NULL DEFAULT 0,
  `procvoucher` VARCHAR(255),
  `procphysician` BIGINT NOT NULL DEFAULT 0,
  `procdt` DATETIME NOT NULL,
  `procdtend` DATETIME,
  `procpos` BIGINT NOT NULL DEFAULT 0,
  `proccomment` VARCHAR(255),
  `procbalorig` DOUBLE NOT NULL DEFAULT 0,
  `procbalcurrent` DOUBLE NOT NULL DEFAULT 0,
  `procamtpain` DOUBLE NOT NULL DEFAULT 0,
  `procbilled` BIGINT NOT NULL DEFAULT 0,
  `procbillable` BIGINT NOT NULL DEFAULT 0,
  `procauth` BIGINT NOT NULL DEFAULT 0,
  `procrefdoc` BIGINT NOT NULL DEFAULT 0,
  `procrefdt` DATETIME,
  `procamtallowed` TEXT,
  `procdtbilled` DATETIME,
  `proccurcovid` BIGINT NOT NULL DEFAULT 0,
  `proccurcovtp` BIGINT NOT NULL DEFAULT 0,
  `proccov1` BIGINT NOT NULL DEFAULT 0,
  `proccov2` BIGINT NOT NULL DEFAULT 0,
  `proccov3` BIGINT NOT NULL DEFAULT 0,
  `proccov4` BIGINT NOT NULL DEFAULT 0,
  `procmedicaidref` VARCHAR(255),
  `procmedicaidresub` VARCHAR(255),
  `proclabcharges` DOUBLE NOT NULL DEFAULT 0,
  `procstatus` VARCHAR(255),
  `procslidingscale` VARCHAR(255),
  `proctosoverride` BIGINT NOT NULL DEFAULT 0,
  `orderid` BIGINT NOT NULL DEFAULT 0,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `scheduler` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `caldateof` DATETIME NOT NULL,
  `calcreated` DATETIME NOT NULL,
  `calmodified` DATETIME,
  `caltype` VARCHAR(255) NOT NULL DEFAULT '',
  `calhour` BIGINT NOT NULL DEFAULT 0,
  `calminute` BIGINT NOT NULL DEFAULT 0,
  `calduration` BIGINT NOT NULL DEFAULT 0,
  `calfacility` BIGINT,
  `calroom` BIGINT,
  `calphysician` BIGINT NOT NULL DEFAULT 0,
  `calpatient` BIGINT NOT NULL DEFAULT 0,
  `calcptcode` BIGINT,
  `calstatus` VARCHAR(255) NOT NULL DEFAULT '',
  `calprenote` VARCHAR(255) NOT NULL DEFAULT '',
  `calpostnote` VARCHAR(255),
  `calmark` BIGINT NOT NULL DEFAULT 0,
  `calgroupid` BIGINT NOT NULL DEFAULT 0,
  `calgroupmembers` VARCHAR(255),
  `calrecurnote` VARCHAR(255),
  `calrecurid` BIGINT NOT NULL DEFAULT 0,
  `calappttemplate` BIGINT NOT NULL DEFAULT 0,
  `calattendees` VARCHAR(255),
  `user` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `scheduler_status` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `csstamp` DATETIME NOT NULL,
  `cspatient` BIGINT NOT NULL DEFAULT 0,
  `csappt` BIGINT NOT NULL DEFAULT 0,
  `csstatus` VARCHAR(255) NOT NULL DEFAULT '',
  `csenote` VARCHAR(255) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `schedulerstatustype` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `sname` VARCHAR(255) NOT NULL DEFAULT '',
  `sdescrip` VARCHAR(255) NOT NULL DEFAULT '',
  `scolor` VARCHAR(255) NOT NULL DEFAULT '',
  `sage` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `systemnotification` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `stamp` DATETIME NOT NULL,
  `nuser` BIGINT NOT NULL DEFAULT 0,
  `ntext` VARCHAR(255) NOT NULL DEFAULT '',
  `naction` VARCHAR(255) NOT NULL DEFAULT '',
  `nmodule` VARCHAR(255) NOT NULL DEFAULT '',
  `npatient` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `user` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `username` VARCHAR(255) NOT NULL DEFAULT '',
  `userpassword` VARCHAR(255) NOT NULL DEFAULT '',
  `usertype` VARCHAR(255),
  `userrealphy` BIGINT NOT NULL DEFAULT 0,
  `userfname` VARCHAR(255),
  `usermname` VARCHAR(255),
  `userlname` VARCHAR(255),
  `userdescrip` VARCHAR(255),
  `userlevel` LONGBLOB,
  `userfac` LONGBLOB,
  `userphy` LONGBLOB,
  `userphygrp` LONGBLOB,
  `usermanageopt` LONGBLOB,
  `useremail` VARCHAR(255),
  `usersms` BIGINT,
  `usersmsprovider` BIGINT,
  `usertitle` VARCHAR(255)
);

CREATE TABLE `workflow_status` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `stamp` DATETIME NOT NULL,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `user` BIGINT NOT NULL DEFAULT 0,
  `status_type` BIGINT NOT NULL DEFAULT 0,
  `status_completed` TINYINT(1) NOT NULL DEFAULT 0
);

CREATE TABLE `workflow_status_type` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `status_name` VARCHAR(255) NOT NULL DEFAULT '',
  `status_order` VARCHAR(255) NOT NULL DEFAULT '',
  `status_module` VARCHAR(255) NOT NULL DEFAULT '',
  `active` TINYINT(1) NOT NULL DEFAULT 0
);

CREATE TABLE `allergies` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `modules` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `module_name` VARCHAR(255) NOT NULL DEFAULT '',
  `module_class` VARCHAR(255) NOT NULL DEFAULT '',
  `module_table` VARCHAR(255) NOT NULL DEFAULT '',
  `module_hidden` TINYINT(1) NOT NULL DEFAULT 0
);

CREATE TABLE `patient_address` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `line1` VARCHAR(255),
  `line2` VARCHAR(255),
  `city` VARCHAR(255),
  `stpr` VARCHAR(255),
  `postal` VARCHAR(255),
  `zip` VARCHAR(255),
  `active` TINYINT(1) NOT NULL DEFAULT 0
);

CREATE TABLE `pharmacy` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `phname` VARCHAR(255) NOT NULL DEFAULT '',
  `phcity` VARCHAR(255),
  `phstate` VARCHAR(255)
);

CREATE TABLE `zipcodes` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `zip` VARCHAR(255) NOT NULL DEFAULT '',
  `city` VARCHAR(255) NOT NULL DEFAULT '',
  `state` VARCHAR(255) NOT NULL DEFAULT '',
  `latitude` DOUBLE NOT NULL DEFAULT 0,
  `longitude` DOUBLE NOT NULL DEFAULT 0,
  `timezone` BIGINT NOT NULL DEFAULT 0,
  `dst` BIGINT NOT NULL DEFAULT 0,
  `country` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `vitals` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `patient` BIGINT NOT NULL,
  `date_taken` DATETIME NOT NULL,
  `systolic` INTEGER,
  `diastolic` INTEGER,
  `heart_rate` INTEGER,
  `respiratory_rate` INTEGER,
  `temperature` DECIMAL(4,1),
  `oxygen_saturation` INTEGER,
  `height_cm` DECIMAL(5,1),
  `weight_kg` DECIMAL(5,1),
  `bmi` DECIMAL(4,1),
  `notes` TEXT,
  `user` BIGINT NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  FOREIGN KEY (`patient`) REFERENCES `patient`(`id`)
);

CREATE TABLE `medications` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `drug_name` VARCHAR(255) NOT NULL DEFAULT '',
  `dosage` VARCHAR(255) NOT NULL DEFAULT '',
  `frequency` VARCHAR(255) NOT NULL DEFAULT '',
  `start_date` DATETIME,
  `end_date` DATETIME,
  `prescribing_provider` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `prescriptions` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `drug_name` VARCHAR(255) NOT NULL DEFAULT '',
  `dosage` VARCHAR(255) NOT NULL DEFAULT '',
  `frequency` VARCHAR(255) NOT NULL DEFAULT '',
  `quantity` VARCHAR(255) NOT NULL DEFAULT '',
  `refills` BIGINT NOT NULL DEFAULT 0,
  `date_written` DATETIME NOT NULL,
  `prescribing_provider` BIGINT NOT NULL DEFAULT 0,
  `pharmacy` VARCHAR(255) NOT NULL DEFAULT '',
  `status` VARCHAR(255) NOT NULL DEFAULT 'active',
  `notes` TEXT,
  `user` BIGINT NOT NULL DEFAULT 0,
  FOREIGN KEY (`patient`) REFERENCES `patient`(`id`)
);

-- Legacy tables referenced in queries (no Go model files; column definitions
-- inferred from SQL queries in api/*.go and existing FreeMED 0.9.x schema)

CREATE TABLE `callin` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `cilname` VARCHAR(255) NOT NULL DEFAULT '',
  `cifname` VARCHAR(255) NOT NULL DEFAULT '',
  `cicomplaint` VARCHAR(255) NOT NULL DEFAULT '',
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME
);

CREATE TABLE `patient_coverage` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `insurance_company` BIGINT NOT NULL DEFAULT 0,
  `coverage_type` BIGINT NOT NULL DEFAULT 0,
  `policy_number` VARCHAR(255) NOT NULL DEFAULT '',
  `group_number` VARCHAR(255) NOT NULL DEFAULT '',
  `effective_date` DATETIME,
  `termination_date` DATETIME,
  `primary_coverage` TINYINT(1) NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `schedulerblockslots` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `sbsdate` DATETIME NOT NULL,
  `sbshour` BIGINT NOT NULL DEFAULT 0,
  `sbsminute` BIGINT NOT NULL DEFAULT 0,
  `sbsduration` BIGINT NOT NULL DEFAULT 0,
  `sbsprovider` BIGINT NOT NULL DEFAULT 0,
  `sbsreason` VARCHAR(255) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `referrals` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `referring_provider` BIGINT NOT NULL DEFAULT 0,
  `referral_to` BIGINT NOT NULL DEFAULT 0,
  `referral_type` VARCHAR(255) NOT NULL DEFAULT '',
  `reason` VARCHAR(255) NOT NULL DEFAULT '',
  `status` VARCHAR(255) NOT NULL DEFAULT '',
  `date_referred` DATETIME NOT NULL,
  `date_completed` DATETIME,
  `notes` VARCHAR(255) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `scanned_docs` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `filename` VARCHAR(255) NOT NULL DEFAULT '',
  `description` VARCHAR(255) NOT NULL DEFAULT '',
  `page_count` BIGINT NOT NULL DEFAULT 0,
  `document_date` DATETIME,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT 'active'
);

CREATE TABLE `letters` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `letter_type` VARCHAR(255) NOT NULL DEFAULT '',
  `recipient` VARCHAR(255) NOT NULL DEFAULT '',
  `subject` VARCHAR(255) NOT NULL DEFAULT '',
  `body` TEXT,
  `date_sent` DATETIME,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT '',
  FOREIGN KEY (`patient`) REFERENCES `patient`(`id`)
);

CREATE TABLE `patient_correspondence` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `correspondence_type` VARCHAR(255) NOT NULL DEFAULT '',
  `direction` VARCHAR(255) NOT NULL DEFAULT '',
  `contact_name` VARCHAR(255) NOT NULL DEFAULT '',
  `contact_method` VARCHAR(255) NOT NULL DEFAULT '',
  `summary` TEXT,
  `date` DATETIME,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT '',
  FOREIGN KEY (`patient`) REFERENCES `patient`(`id`)
);

CREATE TABLE `clinical_orders` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `order_type` VARCHAR(255) NOT NULL DEFAULT '',
  `description` VARCHAR(255) NOT NULL DEFAULT '',
  `status` VARCHAR(255) NOT NULL DEFAULT '',
  `date_ordered` DATETIME,
  `ordering_provider` BIGINT NOT NULL DEFAULT 0,
  `notes` TEXT,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT '',
  FOREIGN KEY (`patient`) REFERENCES `patient`(`id`)
);

CREATE TABLE `episode_of_care` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `start_date` DATETIME NOT NULL,
  `end_date` DATETIME,
  `description` VARCHAR(255) NOT NULL DEFAULT '',
  `status` VARCHAR(255) NOT NULL DEFAULT '',
  `provider` BIGINT NOT NULL DEFAULT 0,
  `notes` VARCHAR(255) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

