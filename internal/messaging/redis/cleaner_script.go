package redis

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
  -- Note: Active queue depth never exceeds the number of handlers, which is
  -- capped at 255. It's safe to RPUSH 255 UUIDs in one shot, but if the max
  -- queue depth ever changes, this script may need to do these moves from queue
  -- to queue in a succession of smaller chunks.
  if count > 0 then
    redis.call("RPUSH", KEYS[2], unpack(messageIDs))
  end

  -- Delete dead consumers' active list
  redis.call("DEL", deadConsumerActiveList)

  redis.call("ZREM", KEYS[1], deadConsumerActiveList)
end

return 0
`
