ping # 测试连通性

keys *

dbsize # 查看整个数据库的大小

exists place # (integer) 1
exists tea # (integer) 1 => golang中就是redis.Nil

del place # 成功返回: (integer) 1; 失败返回: (integer) 0

type place # 返回缓存的类型

move place 1 # redis是有多个数据库的, 默认的数据库是0, 如果移动的数据库找不到会直接新建一个
select 1 # select用于切换数据库, 必须在同个数据库下才能查询对应的缓存信息

ttl place # 过期或者key不存在就是-2, 永不过期是-1, 其他时间返回的是秒
expire place 3600
persist place # 设置为永不过期
rename place place2

set num 666 # 默认用String进行存储, 但应该是会做atoi的相关操作
get num
incr num # 对非整数类型就会报错: (error) ERR value is not an integer or out of range

set floatnum 3.14
incrbyfloat floatnum 2.0 # 浮点数的加法用incrbyfloat

# ==== str ====
strlen place
append place hi
getrange place 0 4 # [left, right]区间, 会多输出一个值

# ==== set ====
sadd langs java php c++ golang
smembers langs # 注意: 只能这么输出所有的内容, 使用get会报错
srandmembers langs 3
sismember langs golang # 判断是否在集合内
srem langs java
scard langs
spop langs 2 # 先进先出 是一个队列的结构

# ==== zset ====
zadd foot 16011 tid 20082 huny 2873 nosy
zrange foot 0 -1 # (0 -1)用于获取所有的内容, 但是不会带上分数, -1的含义就是最后一个index, redis里面都是闭区间!
zrange foot 0 -1 withscores # 输出的格式不是很好
zrangebyscore foot 3000 30000 withscores # 把所有的score在[3000, 30000]都输出
zrangebyscore foot 3000 30000 withscores limit 0 1 # limit后面第一个参数是offset, 第二个是count
zincrby foot 2000 tid # zset的加法操作
zcard foot
zcount foot 2000 20000
zrank foot tid # 获取元素的按从小到大顺序的索引, 最小值是0


# ==== list ====
lpush names lily sandy # 从左边push进去, 1) sandy 2) lily
rpush names bob amy # 从右边push进去 3) bob 4) amy
lpop names
rpop names
llen names # 用于获取长度
lpush userids 111 222 111 222 222 222 333 222 222
lrem userids 0 111 # 0表示删除所有的111
lrem userids 1 222 # +1表示从左边删除1个222
lrem userids -3 222 # -3表示从右边删除3个222
lindex userids 2 # 根据id获取值
ltrim userids 200 300 # 只保留范围值

# ==== hash(不是集合 而是一个个体) ====
hset umap name beijing
hset umap name shanghai age 20 address china
hsetnx umap tail 180 # 为hash中不存在的字段赋值, 可以新增一个字段
hget umap age
hmget umap name age tail
hgetall umap
hkeys umap
hvals umap
hexists umap name
hdel umap age