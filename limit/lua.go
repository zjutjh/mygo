package limit

// scriptTokenBucket 令牌桶限流脚本
// KEYS[1]: 限流 Key
// ARGV[1]: 桶容量 (burst)
// ARGV[2]: 生成速率 (rate, tokens/sec)
// ARGV[3]: 当前时间戳 (秒)
// ARGV[4]: 本次消耗 (cost)
const scriptTokenBucket = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])

-- 获取当前状态: tokens, last_time
local info = redis.call("HMGET", key, "tokens", "last_time")
local last_tokens = tonumber(info[1]) or capacity
local last_time = tonumber(info[2]) or now

-- 计算时间差补充令牌
local delta = math.max(0, now - last_time)
local to_add = delta * rate
local new_tokens = math.min(capacity, last_tokens + to_add)

local allowed = 0
local retry_after = 0

if new_tokens >= cost then
    allowed = 1
    new_tokens = new_tokens - cost
    -- 更新状态
    redis.call("HMSET", key, "tokens", new_tokens, "last_time", now)
    redis.call("EXPIRE", key, 60) -- 续期防止死键
else
    -- 计算大概需要等待多久才能有一个令牌
    local needed = cost - new_tokens
    if rate > 0 then
        retry_after = math.ceil(needed / rate)
    end
end

return {allowed, retry_after}
`

// scriptGCRA 漏桶(GCRA)限流脚本
// KEYS[1]: 限流 Key
// ARGV[1]: 速率 (rate, requests/sec)
// ARGV[2]: 突发 (burst)
// ARGV[3]: 当前时间戳 (毫秒)
const scriptGCRA = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

-- 发射间隔 (毫秒/次)
local emission_interval = 1000 / rate
-- 突发容忍时间 (毫秒) = burst * emission_interval
local burst_offset = burst * emission_interval

-- 获取 TAT (Theoretical Arrival Time)
local tat = tonumber(redis.call("GET", key)) or now

-- 如果 TAT 在过去，重置为 now
tat = math.max(tat, now)

-- 计算本次请求后的新 TAT
local new_tat = tat + emission_interval

-- 允许的最早时间点 = new_tat - burst_offset
-- 如果 allow_at <= now，说明还在桶容量允许范围内
local allow_at = new_tat - burst_offset

local allowed = 0
local retry_after = 0

if allow_at <= now then
    allowed = 1
    redis.call("SET", key, new_tat, "EX", 60)
else
    -- 拒绝，计算等待时间 (毫秒)
    retry_after = math.ceil((allow_at - now) / 1000) -- 转回秒
end

return {allowed, retry_after}
`
