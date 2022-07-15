SELECT P.permissionid, P.permTypeRW, P.permTypeDenyAllow
FROM Permissions P
JOIN RolePermIntersect RPI on P.permissionid = RPI.permissionid
JOIN Roles R on R.roleid = RPI.roleid
JOIN KeyRoleIntersect KRI on R.roleid = KRI.roleid
WHERE keyid = ?
  AND resourceid = ?
  AND roleTypeRK = ? -- roletype role = 0, roletype key = 1
  AND P.permTypeRW = ? -- read = 0, 1 - write
  AND P.permTypeDenyAllow = ?; -- deny = 0, allow = 1
