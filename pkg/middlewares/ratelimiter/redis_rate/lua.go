package redis_rate

import "github.com/redis/go-redis/v9"

// Copyright (c) 2017 Pavel Pravosud
// https://github.com/rwz/redis-gcra/blob/master/vendor/perform_gcra_ratelimit.lua
// var allowN = redis.NewScript(`
// -- this script has side-effects, so it requires replicate commands mode
// redis.replicate_commands()

// local rate_limit_key = KEYS[1]
// local burst = ARGV[1]
// local rate = ARGV[2]
// local period = ARGV[3]
// local cost = tonumber(ARGV[4])

// local emission_interval = period / rate
// local increment = emission_interval * cost
// local burst_offset = emission_interval * burst

// -- redis returns time as an array containing two integers: seconds of the epoch
// -- time (10 digits) and microseconds (6 digits). for convenience we need to
// -- convert them to a floating point number. the resulting number is 16 digits,
// -- bordering on the limits of a 64-bit double-precision floating point number.
// -- adjust the epoch to be relative to Jan 1, 2017 00:00:00 GMT to avoid floating
// -- point problems. this approach is good until "now" is 2,483,228,799 (Wed, 09
// -- Sep 2048 01:46:39 GMT), when the adjusted value is 16 digits.
// local jan_1_2017 = 1483228800
// local now = redis.call("TIME")
// now = (now[1] - jan_1_2017) + (now[2] / 1000000)

// local tat = redis.call("GET", rate_limit_key)

// if not tat then
//   tat = now
// else
//   tat = tonumber(tat)
// end

// tat = math.max(tat, now)

// local new_tat = tat + increment
// local allow_at = new_tat - burst_offset

// local diff = now - allow_at
// local remaining = diff / emission_interval

// if remaining < 0 then
//   local reset_after = tat - now
//   local retry_after = diff * -1
//   return {
//     0, -- allowed
//     0, -- remaining
//     tostring(retry_after),
//     tostring(reset_after),
//   }
// end

// local reset_after = new_tat - now
// if reset_after > 0 then
//   redis.call("SET", rate_limit_key, new_tat, "EX", math.ceil(reset_after))
// end
// local retry_after = -1
// return {cost, remaining, tostring(retry_after), tostring(reset_after)}
// `)

// var allowAtMost = redis.NewScript(`
// -- this script has side-effects, so it requires replicate commands mode
// redis.replicate_commands()

// local rate_limit_key = KEYS[1]
// local burst = ARGV[1]
// local rate = ARGV[2]
// local period = ARGV[3]
// local cost = tonumber(ARGV[4])

// local emission_interval = period / rate
// local burst_offset = emission_interval * burst

// -- redis returns time as an array containing two integers: seconds of the epoch
// -- time (10 digits) and microseconds (6 digits). for convenience we need to
// -- convert them to a floating point number. the resulting number is 16 digits,
// -- bordering on the limits of a 64-bit double-precision floating point number.
// -- adjust the epoch to be relative to Jan 1, 2017 00:00:00 GMT to avoid floating
// -- point problems. this approach is good until "now" is 2,483,228,799 (Wed, 09
// -- Sep 2048 01:46:39 GMT), when the adjusted value is 16 digits.
// local jan_1_2017 = 1483228800
// local now = redis.call("TIME")
// now = (now[1] - jan_1_2017) + (now[2] / 1000000)

// local tat = redis.call("GET", rate_limit_key)

// if not tat then
//   tat = now
// else
//   tat = tonumber(tat)
// end

// tat = math.max(tat, now)

// local diff = now - (tat - burst_offset)
// local remaining = diff / emission_interval

// if remaining < 1 then
//   local reset_after = tat - now
//   local retry_after = emission_interval - diff
//   return {
//     0, -- allowed
//     0, -- remaining
//     tostring(retry_after),
//     tostring(reset_after),
//   }
// end

// if remaining < cost then
//   cost = remaining
//   remaining = 0
// else
//   remaining = remaining - cost
// end

// local increment = emission_interval * cost
// local new_tat = tat + increment

// local reset_after = new_tat - now
// if reset_after > 0 then
//   redis.call("SET", rate_limit_key, new_tat, "EX", math.ceil(reset_after))
// end

// return {
//   cost,
//   remaining,
//   tostring(-1),
//   tostring(reset_after),
// }
// `)

var allowTokenBucket = redis.NewScript(`
local function currentTime()
    local t = redis.call("TIME")
    return tonumber(t[1])*1000 + (tonumber(t[2]) / 1000)
end


local function advance(t, bucket)
    local last = bucket.last
    if t < last then
        last = t
    end
    local elapsed = t - last
    local delta = bucket.limit * elapsed
    local tokens = bucket.tokens + delta
    tokens = math.min(tokens, bucket.burst)
    return t, tokens
end

--[[
  A lua rate limiter script run in redis
  use token bucket algorithm.
  Algorithm explaination
  1. key, use this key to find the token bucket in redis
  2. there're several args should be passed in:
       intervalPerPermit, time interval in millis between two token permits; interval between 2 actions 1/limit (s/req)z
       refillTime, timestamp when running this lua script;
       limit, the capacity limit of the token bucket;
       interval, the time interval in millis of the token bucket;
]] --
local key = KEYS[1]
-- local burstTokens, intervalPerPermit, refillTime = tonumber(ARGV[1]), tonumber(ARGV[2]), tonumber(ARGV[3])
local limit, burst, ttl, refillTime, n, maxDelay = tonumber(ARGV[1]), tonumber(ARGV[2]), tonumber(ARGV[3]),
    tonumber(ARGV[4]), tonumber(ARGV[5]), tonumber(ARGV[6])
local bucket = {
    limit = limit,
    burst = burst,
    tokens = 0,
    last = 0,
    lastEvent = 0
}

local reservation = {
    tokens = 0,
    ok = false,
    limit = limit,
    lim = bucket,
    timeToAct = 0
}

local rlSource = redis.call('hgetall', key)

-- TODO: check case limit inf and equal to 0.
if table.maxn(rlSource) == 0 then
    -- first check if bucket not exists, if yes, create a new one with full capacity, then grant access
    reservation.tokens = 0
    bucket.tokens = 0
elseif table.maxn(rlSource) == 6 then
    -- ! main update
    -- local lastRefillTime, tokensRemaining = tonumber(rlSrouce[2]), tonumber(rlSrouce[4])
    -- local lastEvent = tonumber(rlSrouce[6])
    -- Get bucket state from redis
    bucket.last = tonumber(rlSource[2])
    bucket.tokens = tonumber(rlSource[4])
    bucket.lastEvent = tonumber(rlSource[6])
end

-- Reserve()
-- TODO: edge cases for later.
local t, tokens = advance(refillTime, bucket)
tokens = tokens - n
local waitDuration = 0
if tokens < 0 then
    waitDuration = (tokens * -1) / bucket.limit
end

local ok = false
if n <= bucket.burst then
    ok = true
end

reservation.ok = ok
reservation.limit = limit
reservation.lim = bucket
if ok then
    reservation.tokens = n
    reservation.timeToAct = t + waitDuration

    bucket.last = t
    bucket.tokens = tokens
    bucket.lastEvent = reservation.timeToAct

    reservation.lim = bucket
end
-- end Reserve()

local delay = 0
if ok then
    local now = currentTime()
    delay = reservation.timeToAct - now
    if delay < 0 then
        delay = 0
    end

    -- cancel 
    if delay > maxDelay and reservation.tokens ~= 0 and reservation.timeToAct >= t then
        local t = now
        local toRestore = reservation.lim.limit * (reservation.lim.lastEvent - reservation.timeToAct)
        local restoreTokens = reservation.tokens - toRestore
        if restoreTokens > 0 then
            local t, tokens = advance(t, reservation.lim)
            tokens = tokens + restoreTokens
            tokens = math.min(tokens, reservation.lim.burst)
            reservation.lim.last = t
            reservation.lim.tokens = tokens
            bucket = reservation.lim
            if reservation.timeToAct == reservation.lim.lastEvent then
                local now = currentTime()
                local preEvent = reservation.timeToAct + ((reservation.tokens * -1) / limit)
                if  preEvent >= t then
                    reservation.lim.lastEvent = preEvent
                    bucket = reservation.lim
                end
            end
        end
    end
    -- end cancel
end

-- ! cancel 

redis.call('hset', key, 'last', t)
redis.call('hset', key, 'tokens', reservation.lim.tokens)
redis.call('hset', key, 'lastEvent', reservation.lim.lastEvent)
redis.call('expire', key, ttl)

return {tostring(ok), tostring(tokens), tostring(waitDuration), tostring(delay)}
`)
