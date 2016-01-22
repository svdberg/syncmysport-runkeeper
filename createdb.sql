
CREATE TABLE `sync` (
    `uid` INT(10) NOT NULL AUTO_INCREMENT,
    `rk_key` VARCHAR(64) NULL DEFAULT NULL,
    `stv_key` VARCHAR(64) NULL DEFAULT NULL,
    `last_succesfull_retrieve` DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (`uid`)
  );
