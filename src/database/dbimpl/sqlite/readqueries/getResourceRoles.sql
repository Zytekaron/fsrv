-- get role allowed/denied keys by precedence for a given resource and key
SELECT Roles.roleid, P.permTypeDenyAllow, P.permTypeRWMD
FROM Permissions P
         JOIN RolePermIntersect RPI ON P.permissionid = RPI.permissionid
         JOIN Roles ON RPI.roleid = Roles.roleid
WHERE P.resourceid = ? AND Roles.roleTypeRK = 0 -- note: roleType 0 = role
ORDER BY Roles.rolePrecedence, Roles.roleid;