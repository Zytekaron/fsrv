-- get if resource is publicly visible
SELECT flags
FROM Resources
WHERE resourceid = ?;