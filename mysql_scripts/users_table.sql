-- Please note that the mysql_scripts folder and its contents are not part of the reference implementation. The sql script files contained in the mysql_scripts folder are means for creating the snippetbox DB (some scripts are still missing, but I intend to add them later once the project is finished). This also explains why this folder and the *.sql files within are not idiomatic Go.

-- Skip the USE command below if already logged into the snippetbox database.
USE snippetbox;

CREATE TABLE users(
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hashed_password CHAR(60) NOT NULL,
    created DATETIME NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);