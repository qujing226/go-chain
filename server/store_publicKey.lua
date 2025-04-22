-- store_publicKey.lua
-- 说明：
-- KEYS[1]：存储公钥的 key（例如 DID 字符串）
-- ARGV[1]：公钥的序列化数据（例如 JSON 字符串）

return redis.call('SET', KEYS[1], ARGV[1], 'EX', 300)