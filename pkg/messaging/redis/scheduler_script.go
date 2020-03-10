package redis

// nolint: lll
var schedulerScript = `
-- KEYS[1]: the scheduled set
-- KEYS[2]: the pending list

-- ARGV[1]: the current timestamp
-- ARGV[2]: the max number of messageIDs to transfer to the pending list

-- Returns: nil

-- Get the messageIDs out of the scheduled set
local messageIDs = redis.call("zrangebyscore", KEYS[1], 0, ARGV[1], "LIMIT", 0, ARGV[2])
local messageCount = table.getn(messageIDs)

if messageCount > 0 then
  -- Push them on to the pending list
  redis.call("lpush", KEYS[2], unpack(messageIDs))

  -- Remove them from the scheduled set
  return redis.call("zremrangebyrank", KEYS[1], 0, messageCount - 1)
end

return 0
`
