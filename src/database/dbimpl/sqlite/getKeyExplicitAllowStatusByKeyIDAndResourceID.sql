SELECT COUNT(keyid)
FROM Permissions P
JOIN RolePermIntersect RPI on P.permissionid = RPI.permissionid
JOIN Roles R on R.roleid = RPI.roleid
JOIN KeyRoleIntersect KRI on R.roleid = KRI.roleid
WHERE keyid = ?
  AND resourceid = ?
  AND roleTypeRK = 1 -- roletype = 1 = key
  AND P.permTypeRW = ?
  AND P.permTypeDenyAllow = 1;
