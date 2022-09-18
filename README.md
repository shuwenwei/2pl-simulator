#### 附加思考
会出现死锁吗?如何避免?

会

预防的方式：wait-die或者wound-wait机制，存在较多的误杀的情况

wait-die: 当申请加锁的数据被其他事务占用，时间戳更小时等待，否则自己abort

wound-wait: 当申请加锁的数据被其他事务占用，时间戳大时等待，否则将当前事务等待的事务abort

检测方式：维护一个wait-for graph，当事务间依赖构成环时选择环的一个节点abort

锁超时方式：允许申请锁的事务等待一段时间，在期间内无法获得锁则超时abort



homework中使用的方式为wait-die