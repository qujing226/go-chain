-- get_publicKey.lua
-- 说明：
-- KEYS[1]：用于存放公钥的 key

local value = redis.call('GET', KEYS[1])
if value then
    -- 获取成功后删除，保证只能取一次
    redis.call('DEL', KEYS[1])
end
return value