package redis

// TODO: This script might not hold up well for large numbers of orphaned
// messages. Need to chunk up the work.

// nolint: lll
var cleanerScript = `
-- KEYS[1]: the consumers set
-- KEYS[2]: the pending list

-- ARGV[1]: the timestamp before which an active list is considered as beloging to a dead consumer

-- Returns: nil

-- Find dead consumers' active lists
local deadConsumerActiveLists = redis.call("ZRANGEBYSCORE", KEYS[1], 0, ARGV[1])

for _, deadConsumerActiveList in ipairs(deadConsumerActiveLists) do
  local messageIDs = redis.call("LRANGE", deadConsumerActiveList, 0, -1)
  local count = table.getn(messageIDs)

  -- Push any orphaned message IDs onto the pending list
  if count > 0 then
    redis.call("RPUSH", KEYS[2], unpack(messageIDs))
  end

  -- Delete dead consumers' active list
  redis.call("DEL", deadConsumerActiveList)

  redis.call("ZREM", KEYS[1], deadConsumerActiveList)
end

return 0
`
