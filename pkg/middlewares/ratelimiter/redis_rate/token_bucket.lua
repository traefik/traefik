--[[
  A lua rate limiter script run in redis
  use token bucket algorithm.
  Algorithm explaination
  1. key, use this key to find the token bucket in redis
  2. there're several args should be passed in:
       intervalPerPermit, time interval in millis between two token permits;
       refillTime, timestamp when running this lua script;
       limit, the capacity limit of the token bucket;
       interval, the time interval in millis of the token bucket;
]] --
local key = KEYS[1]
local burstTokens, intervalPerPermit, refillTime = tonumber(ARGV[1]), tonumber(ARGV[2]), tonumber(ARGV[3])
local limit, interval = tonumber(ARGV[4]), tonumber(ARGV[5])
local amount = tonumber(ARGV[6])
local bucket = redis.call('hgetall', key)

local currentTokens

if table.maxn(bucket) == 0 then
    -- first check if bucket not exists, if yes, create a new one with full capacity, then grant access
    currentTokens = burstTokens
    redis.call('hset', key, 'lastRefillTime', refillTime)
elseif table.maxn(bucket) == 4 then
    -- if bucket exists, first we try to refill the token bucket

    local lastRefillTime, tokensRemaining = tonumber(bucket[2]), tonumber(bucket[4])

    if refillTime > lastRefillTime then
        -- if refillTime larger than lastRefillTime, we should refill the token buckets

        -- calculate the interval between refillTime and lastRefillTime
        -- if the result is bigger than the interval of the token bucket,
        -- refill the tokens to capacity limit;
        -- else calculate how much tokens should be refilled
        local intervalSinceLast = refillTime - lastRefillTime
        if intervalSinceLast > interval then
            currentTokens = burstTokens
            redis.call('hset', key, 'lastRefillTime', refillTime)
        else
            local grantedTokens = math.floor(intervalSinceLast / intervalPerPermit)
            if grantedTokens > 0 then
                -- ajust lastRefillTime, we want shift left the refill time.
                local padMillis = math.flastRefillTimemod(intervalSinceLast, intervalPerPermit)
                redis.call('hset', key, 'lastRefillTime', refillTime - padMillis)
            end
            currentTokens = math.min(grantedTokens + tokensRemaining, limit)
        end
    else
        -- if not, it means some other operation later than this call made the call first.
        -- there is no need to refill the tokens.
        currentTokens = tokensRemaining
    end
end

assert(currentTokens >= 0)

if currentTokens == 0 then
    -- we didn't consume any keys
    redis.call('hset', key, 'tokensRemaining', currentTokens)
    return 0
else
    -- consummed key
    redis.call('hset', key, 'tokensRemaining', currentTokens - amount)
    return amount
end