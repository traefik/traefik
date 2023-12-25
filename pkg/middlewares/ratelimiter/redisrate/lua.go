package redisrate

import "github.com/redis/go-redis/v9"

//nolint:dupword
var allowTokenBucket = redis.NewScript(`
local key = KEYS[1]
local limit, burst, ttl, t, n, max_delay = tonumber(ARGV[1]), tonumber(ARGV[2]), tonumber(ARGV[3]), tonumber(ARGV[4]),
    tonumber(ARGV[5]), tonumber(ARGV[6])

if n > burst then
    return {tostring(false), tostring(0)}
end

local bucket = {
    limit = limit,
    burst = burst,
    tokens = 0,
    last = 0
}

local rl_source = redis.call('hgetall', key)

-- TODO: check case limit inf and equal to 0.
if table.maxn(rl_source) == 4 then
    -- Get bucket state from redis
    bucket.last = tonumber(rl_source[2])
    bucket.tokens = tonumber(rl_source[4])
end

-- TODO: edge cases for later.
local last = bucket.last
if t < last then
    last = t
end

local elapsed = t - last
local delta = bucket.limit * elapsed
local tokens = bucket.tokens + delta
tokens = math.min(tokens, bucket.burst)
tokens = tokens - n

local wait_duration = 0
if tokens < 0 then
    wait_duration = (tokens * -1) / bucket.limit
    if wait_duration > max_delay then
        local nowR = redis.call("TIME")
        local now = (tonumber(nowR[1]) * 1000000) + tonumber(nowR[2])
        elapsed = now - t
        delta = bucket.limit * elapsed
        tokens = tokens + n + delta
        tokens = math.min(tokens, burst)
        wait_duration = wait_duration + elapsed
    end
end

redis.call('hset', key, 'last', t)
redis.call('hset', key, 'tokens', tokens)
redis.call('expire', key, ttl)

return {tostring(true), tostring(wait_duration)}`)
