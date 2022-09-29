SELECT roleName, P.permTypeDenyAllow, P.permTypeRW, Roles.roleTypeRK
FROM Permissions P
         JOIN RolePermIntersect RPI ON P.permissionid = RPI.permissionid
         JOIN Roles ON RPI.roleid = Roles.roleid
WHERE P.resourceid = ?
ORDER BY Roles.roleTypeRK DESC, Roles.rolePrecedence DESC, Roles.roleid;