-- get role allowed/denied keys by precedence for a given resource and key
SELECT roleName, P.type
FROM Permissions P
         JOIN RolePermIntersect RPI ON P.permissionid = RPI.permissionid
         JOIN Roles ON RPI.roleid = Roles.roleid
WHERE P.resourceid = ? AND Roles.roleType = 0 -- note: roleType 0 = role
ORDER BY P.type, Roles.rolePrecedence, Roles.roleid;