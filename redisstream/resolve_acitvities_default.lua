local ids=redis.call("ZREVRANGE",KEYS[1],0,ARGV[1])
if table.getn(ids)==0 then return {} end
return redis.call("MGET",unpack(ids))