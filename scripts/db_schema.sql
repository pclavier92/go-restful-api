DROP DATABASE IF EXISTS Music;
CREATE DATABASE Music;
USE Music;

CREATE TABLE Music.Users (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(256) DEFAULT NULL,
  `password` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE Music.Artists (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE Music.Songs (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(256) DEFAULT NULL,
  `duration` varchar(256) DEFAULT NULL,
  `artist_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  FOREIGN KEY `FK_Songs_Artist` (`artist_id`) REFERENCES `Artists` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO Music.Users (`username`, `password`) VALUES ('admin','admin');