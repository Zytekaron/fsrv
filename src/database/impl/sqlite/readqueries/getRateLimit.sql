-- get ratelimit info for a given key
SELECT requests, reset
FROM Ratelimits
WHERE keyid = ?;