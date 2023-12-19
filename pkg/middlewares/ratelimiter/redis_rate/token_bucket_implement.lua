local key = KEYS[1]
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
if table.maxn(rlSource) == 6 then
    -- Get bucket state from redis
    bucket.last = tonumber(rlSource[2])
    bucket.tokens = tonumber(rlSource[4])
    bucket.lastEvent = tonumber(rlSource[6])
end

-- Reserve()
-- TODO: edge cases for later.
local t = refillTime
local last = bucket.last
if t < last then
    last = t
end
local elapsed = t - last
local delta = bucket.limit * elapsed
local tokens = bucket.tokens + delta
tokens = math.min(tokens, bucket.burst)
tokens = tokens - n

-- TODO: waitDuration and delay

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

-- end Reserve()

local delay = 0
if ok then
    reservation.tokens = n
    reservation.timeToAct = t + waitDuration
    bucket.last = t
    bucket.tokens = tokens
    bucket.lastEvent = reservation.timeToAct

    reservation.lim = bucket

    local nowR = redis.call("TIME")
    local now = tonumber(nowR[1]) * 1000 + (tonumber(nowR[2]) / 1000)

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
            local last = bucket.last
            if t < last then
                last = t
            end
            local elapsed = t - last
            local delta = bucket.limit * elapsed
            local tokens = bucket.tokens + delta
            tokens = math.min(tokens, bucket.burst)
            -- local t, tokens = advance(t, reservation.lim)
            tokens = tokens + restoreTokens
            tokens = math.min(tokens, reservation.lim.burst)
            bucket.last = t
            bucket.tokens = tokens
            reservation.lim = bucket
            if reservation.timeToAct == reservation.lim.lastEvent then
                local nowR = redis.call("TIME")
                local now = tonumber(nowR[1]) * 1000 + (tonumber(nowR[2]) / 1000)
                local preEvent = reservation.timeToAct + ((reservation.tokens * -1) / limit)
                if preEvent >= t then
                    reservation.lim.lastEvent = preEvent
                    bucket = reservation.lim
                end
            end
        end
    end
    -- end cancel
end

redis.call('hset', key, 'last', t)
redis.call('hset', key, 'tokens', bucket.tokens)
redis.call('hset', key, 'lastEvent', bucket.lastEvent)
redis.call('expire', key, ttl)

return {tostring(ok), tostring(tokens), tostring(waitDuration), tostring(delay)}
