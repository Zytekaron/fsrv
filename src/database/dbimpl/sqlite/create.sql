CREATE TABLE Keys
(
    keyid   TEXT PRIMARY KEY,
    note    TEXT,
    expires INTEGER NOT NULL, -- unix millis
    created INTEGER NOT NULL  -- unix millis
);



CREATE TABLE Ratelimits
(
    keyid    TEXT PRIMARY KEY,
    requests INTEGER NOT NULL, -- number of requests in a given period
    reset    INTEGER NOT NULL, -- timestamp for reset (unix millis)

    FOREIGN KEY (keyid) REFERENCES Keys (keyid)
);



CREATE TABLE Resources
(
    resourceid TEXT PRIMARY KEY,
    flags   INTEGER(1) NOT NULL
);



CREATE TABLE Roles
(
    roleid         INTEGER PRIMARY KEY,
    roleName       TEXT       NOT NULL,
    rolePrecedence INTEGER    NOT NULL DEFAULT 0, -- note, roleid is tiebreaker
    roleTypeRK     INTEGER(1) NOT NULL            -- 0=role 1=keyrole
);
CREATE INDEX RolesByID ON Roles (roleid); -- for lookups
CREATE INDEX RolesByPrecedence ON Roles (rolePrecedence); -- for ordering
CREATE INDEX RolesByRoleType ON Roles (roleTypeRK); -- for lookups


CREATE TABLE Permissions
(
    permissionid      INTEGER,             -- file / dir permission id
    resourceid        TEXT       NOT NULL, -- type of permission granted
    permTypeRW        INTEGER(1) NOT NULL, -- 0=read 1=write
    permTypeDenyAllow INTEGER(1) NOT NULL, -- -0=deny 1=allow

    FOREIGN KEY (resourceid) REFERENCES Resources (resourceid)
);
CREATE INDEX PermissionsPerResource ON Permissions (resourceid);



CREATE TABLE KeyRoleIntersect
(
    keyid  INTEGER,
    roleid INTEGER,

    FOREIGN KEY (keyid) REFERENCES Keys (keyid),
    FOREIGN KEY (roleid) REFERENCES Roles (roleid)
);
CREATE INDEX RolesPerKey ON KeyRoleIntersect (keyid); -- create index for most commonly accessed direction


CREATE TABLE RolePermIntersect
(
    roleid       INTEGER,
    permissionid INTEGER,

    FOREIGN KEY (roleid) REFERENCES Roles (roleid),
    FOREIGN KEY (permissionid) REFERENCES Permissions (permissionid)
);
CREATE INDEX PermissionsPerRole ON RolePermIntersect (roleid);