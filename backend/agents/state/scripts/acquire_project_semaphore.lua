-- KEYS[1] = setKey
-- ARGV[1] = maxConcurrent
-- ARGV[2] = trackerID
-- ARGV[3] = ttlSeconds
-- ARGV[4] = heartbeatKey

local setKey = KEYS[1]
local maxConcurrent = tonumber(ARGV[1])
local trackerID = ARGV[2]
local ttlSeconds = tonumber(ARGV[3])
local heartbeatKey = ARGV[4]

-- Prevent duplicate acquisition
if redis.call("EXISTS", heartbeatKey) == 1 then
  return 2
end

if redis.call("SCARD", setKey) >= maxConcurrent then
  return 0
end

redis.call("SADD", setKey, trackerID)
redis.call("SETEX", heartbeatKey, ttlSeconds, 1)
return 1
