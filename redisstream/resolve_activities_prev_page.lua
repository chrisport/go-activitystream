local ids=redis.call("ZRANGEBYSCORE",KEYS[1],ARGV[1],"+inf","LIMIT",1,ARGV[2])
if table.getn(ids)==0 then return {} end
return redis.call("MGET",unpack(ids))