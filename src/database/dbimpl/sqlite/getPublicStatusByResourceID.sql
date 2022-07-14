-- get if resource is publicly visible
SELECT ispublic
FROM Resources
WHERE resourceid = ?;