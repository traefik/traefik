-- Input parameters
local tokens_key = KEYS[1]..":tokens"           -- Key for the bucket's token counter
local last_access_key = KEYS[1]..":last_access" -- Key for the bucket's last access time

local capacity = tonumber(ARGV[1])  -- Maximum number of tokens in the bucket
local rate = tonumber(ARGV[2])      -- Rate of token generation (tokens/second)
local now = tonumber(ARGV[3])       -- Current timestamp in microseconds
local requested = tonumber(ARGV[4]) -- Number of tokens requested for the operation

-- Fetch the current token count
local last_tokens = tonumber(redis.call("get", tokens_key))
if last_tokens == nil then
    last_tokens = capacity
end

-- Fetch the last access time
local last_access = tonumber(redis.call("get", last_access_key))
if last_access == nil then
    last_access = 0
end

-- Calculate the number of tokens to be added due to the elapsed time since the
-- last access. We cap the number at the capacity of the bucket.
local elapsed = math.max(0, now - last_access)
local add_tokens = math.floor(elapsed * rate / 1000000)
local new_tokens = math.min(capacity, last_tokens + add_tokens)

-- Calculate the new last access time. We don't want to use the current time as
-- the new last access time, because that would result in a rounding error.
local new_access_time = last_access + math.ceil(add_tokens * 1000000 / rate)

-- Check if enough tokens have been accumulated
local allowed = new_tokens >= requested
if allowed then
    new_tokens = new_tokens - requested
end

-- Update state
redis.call("setex", tokens_key, 60, new_tokens)
redis.call("setex", last_access_key, 60, new_access_time)

-- Return 1 if the operation is allowed, 0 otherwise.
return allowed and 1 or 0
