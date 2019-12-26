USE msgbox;

CREATE TABLE users (
  name VARCHAR(32) NOT NULL,
  PRIMARY KEY (name)
);

CREATE TABLE usergroups (
  name VARCHAR(32) NOT NULL,
  PRIMARY KEY (name)
);

CREATE TABLE groupusers (
  id INT NOT NULL AUTO_INCREMENT,
  groupname VARCHAR(32) NOT NULL,
  username VARCHAR(32) NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (groupname) REFERENCES usergroups(name),
  FOREIGN KEY (username) REFERENCES users(name)
);

CREATE INDEX groupusers_idx ON groupusers(groupname);
