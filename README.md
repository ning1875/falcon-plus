# 架构图
![image](https://github.com/ning1875/falcon-plus/blob/master/images/new.png)


### 我重写了聚合器,重写聚合器目的 poly_metric VS aggregator

- 解决endpoint多的聚合断点问题
- 解决聚合器单点问题，使得横向扩展得以实现
- 解耦聚合器各个单元，可以方便的增加新的聚合入口和聚合策略	



### falcon agent自升级
```
	过程说明：
	http-req --->hbs --->开启升级开关--->检查agent心跳信息中版本号，并检查当前hbs升级队列--->发送升级指令给agent ---> agent通过 升级命令中的url地址和目标版本号下载新的二进制（会有备份和回滚逻辑）--->agent check没有问题后获取自身的pid向自己发送kill信号 --->agent退出然后会被systemd拉起打到升级的目的--->新的心跳信息中版本checkok不会继续升级
	升级举例说明:
	1. falcon-agent新加了采集指标，测试OK后在代码中打上新的版本号比如7.1（现有是7.0）
	2. 然后将7.1版 放到下载服务器的路径下 目前是 127.0.0.1/data00/tgz/open-falcon/bin_7.1，并确保下载ok ,wget http://127.0.0.1/file/open-falcon/bin_7.1
	3. 然后向hbs 发送升级的http请求（这里有个保护机制:只能在hbs本机发起）
	4. 然后通过hbs 的http接口查询当前心跳上来的agent的版本查看升级进度 ，curl -s http://localhost:6031/agentversions |python -m "json.tool"
	5. 同时需要连接的redis集群观察 agent_upgrade_set 这个set的值，redis-cli -h ip -p port -c smembers agent_upgrade_set & scard agent_upgrade_set
	6. 目前看并发2000可以把一台下载的nginx万兆网卡流量打满。1.24GB/s

	## falcon-agent 自升级
	curl -X POST   http://127.0.0.1:6031/agent/upgrade -d '{"wgeturl":"http://127.0.0.1/file/open-falcon","version":"6.0.1","binfile_md5":"35ac8534c0b31237e844ef8ee2bb9b9e"}'

	curl -X GET  http://127.0.0.1:6031/agent/upgrade/nowargs
	{"msg":"success","data":{"type":0,"wgeturl":"http://127.0.0.1/file/open-falcon","version":"6.0.1","binfile_md5":"35ac8534c0b31237e844ef8ee2bb9b9e","cfgfile_md5":""}}

	curl http://127.0.0.1:6031/agentversions
	{"msg":"success","data":{"n3-021-225":"6.0.1"}}

	curl -X DELETE  http://127.0.0.1:6031/agent/upgrade
	{"msg":"success","data":"取消升级成功"}

	  uri:
		url: http://127.0.0.1:6031/agent/upgrade
		method: POST
		body: {"wgeturl":"http://127.0.0.1/file/open-falcon","version":"6.0.2","binfile_md5":"f5c597f15e379a77d1e2ceeec7bd99a8"}
		status_code: 200
		body_format: json
```
### 报警优化：
    1.alarm添加电话报警,IM报警	
	2.报警发送优化pop速度,避免报警堆积在redis中
	3.dashboard:配置页面增加说明,alarm页面只展示自己组的报警,新增报警记录搜索功能
	4.改造发送IM逻辑， 主备两个机器人发送失败就停止发送变为:第一轮主备两个机器人尝试发送4次,如果本轮失败的话推入失败队列 10分钟后再发送,依然是两个机器人尝试发送4次如果最终失败的话推入失败队列,后面会对最终失败队列做监控
### 报警屏蔽: 
	1.im通道能把报警接收人,要屏蔽的endpoint+counter，屏蔽接口地址等信息发送给接收人。
	2.接收人通过点击屏蔽1小时等按钮将屏蔽信息发往alarm的api接口
	3.alarm收到屏蔽信息，根据user+counter+rediskey前缀作为redis-key ，SETEX到redis中，value和ttl都是屏蔽的时长，并sadd到redis一个记录所有屏蔽记录的set中，然后更新到到自身的内存map中
	4.启动一个定时任务，定时同步redis中的屏蔽信息到alarm的内存中：smemebers redis屏蔽的set 然后get 这个key如果不存在说明屏蔽已失效，从alarm的map中删除，如果存在更新到内存中(解决alarm重启后重建屏蔽信息)
	5.alarm消费报警事件时分高低优先级 check 屏蔽map如果cache hit则取消发送报警：比如报警事件是同样的，只是接受人是一个列表，这里屏蔽的体现就是从发送人列表中将屏蔽人剔除
	6.取消屏蔽也需要发送一个请求到alarm，alarm更新redis和自己map即可
### 修复聚合器机器数量1k+聚合失败断点问题,聚合器加cache减轻db压力
   1.聚合器在获取机器数量多的组时调用的api原版的问题是 获取了host的其他信息，而agg需要的是根据group_name获取机器列表
   也就是获取一个host后还要关联查询很多信息导致1k+机器光调用这个接口就花费20s，新增了值获取机器列表的接口
   
### 解决graph和hbs代码中duplicate更新db引发的db插入失败问题:
	1.表有不止一个唯一键时 insert on duplit key update  可能会有问题 
	2.When the table has more than one unique or primary key, this statement is sensitive to the order in which the storage engines checks the keys. Depending on this order, the storage engine may determine different rows to mysql, and hence mysql can update different rows. 
	3.问题链接 https://bugs.mysql.com/bug.php?id=58637
	4.现象就是在mysql slow_querylog中能看到一条insert lock了100多秒 最后的row_affected是0就是插入失败了。导致db整体响应慢
	5.修复过程：变insert on duplicate 为先插入报错后再更新
### 所有组件中需要http调用其他接口失败加重试，加cache减少调用次数 使用这个包"github.com/patrickmn/go-cache"
### 双击房改造：
    1.服务组件双机房部署，agent采集流量按机房分开：transfer和hbs的域名做dns view，
	2.引入新组件falcon-apiproxy替换原有api，使用户查询时无需关心数据机房属性。通过nginx 控制请求的path既：所有需要请求双机房数据的查询都pass到apiproxy，apiproxy拿到请求的数据，解析成分机房的endpoint列表
	。判断endpoint属于哪个机房是apiproxy通过配置文件中配置的各个机房hbs地址然后定时遍历请求hbs rpc 获取每个hbs连接过来的endpoint信息存储到apiproxy中以 机房名称为key的map中，每次查询请求过来时在cache中查询这个endpoint应该属于哪个机房
    3.获取到endpoint列表对应机房的信息然后发起请求到各个api查询数据后拼接在一起回复给用户
	4.python falcon-dashboard双写db模式，使用户变更配置时操作一次即可
### 国内海外容量规划和agent采集间隔变化:
    1.agent默认采集间隔由30秒变为按采集项目的重要程度分为10，30，60，120四个档位。
	2.原有的默认采集cpu单核监控变为可配置
### HBS多实例全表查询mysql导致mysql压力大
    1.利用redis实现分布式锁
	2.抢到锁的hbs实例负责查询mysql更新到redis
	3.抢锁失败的hbs到redis中查询
    4.具体说明：setnx 锁的key  value是 unix_time.now + timeout ，setnx成功说明抢到锁，并把锁的ttl设置为timeout
	setnx失败检查锁的值是否已经超时(防止有expire命令没执行成功也就是通过value判断锁是否超时)，抢锁成功的实例查询mysql并更新的redis中，并且把这次更新的key 放到一个名字不会变的key，让所有hbs查询这个key 类似服务发现
### falcon-agent 新增监控项
	1.TcpExt.TCPFastRetransRate- #用来表示重传率这是一个放大1w倍的值 一般网络上认为百分之一的重传数为临界值,所以TcpExt.TCPFastRetransRate 超过100认为异常
	2.mem.shmem #用来表示共享内存的值,如果机器上有用到/dev/shm做读写的话,算内存使用率的时候  /dev/shm 的用量 +  普通的 mem used = 真实内存用量
	3.mem.memavailable（单位字节） & mem.memavailable.percent（百分比)
	4.percore.busy/core=core01每个逻辑cpu的使用量 ，监控每个逻辑cpu新增开关配置
	5.sys.ntp.offset
	代表本机和ntpserver之间的ntp偏移 ，单位是毫秒
	正的值代表是超前，负值代表是滞后
	6.sys.uptime
	代表机器启动时间 单位是秒

### python-dashboard
	1.双写db模式，使用户变更配置时操作一次即可
	2.dashboard机器组添加tag: 新增首页服务树,group页面组拉取自己的tag组
	3.dashboard: 重写screen搜索逻辑,优化速度,优化cas认证完跳转问题,给部分api提供跳过机制

	
### falcon-alarm 支持群组报警
### 聚合器gauge 和counter 问题
原始聚合器在处理counter类型的聚合时做法是 例如query查出来的数据转化成gauge类型的再聚合计算。这样算sum和avg就可以摆脱每次聚合数量不一致导致的
混入特别大的原始点 ，带来数据的混乱
具体比如 网卡速率在30k左右，但是/proc/net/dev 中的数据远大于这个值

### 新增组件 poly_metric  我重写了聚合器,重写聚合器目的 poly_metric VS aggregator
- 解决endpoint多的聚合断点问题
- 解决聚合器单点问题，使得横向扩展得以实现
- 解耦聚合器各个单元，可以方便的增加新的聚合入口和聚合策略	

###新增proxy模块支持api查询多机房
