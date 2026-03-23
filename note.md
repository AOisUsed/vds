## todos
- 测试 cancel context的 是否影响数据一致性和性能



## 应用层功能
- SendMessage(msg)
- CancelMessage ()
- SubscribeDeviceMessage(id )
- ActivateAndRegisterDevice(ctx, id, opts)
- TerminateAndDeregisterDevice(ctx, id)




## Connection 问题
- redisRepo 操作:
  - 获得 Connection类型
    - 得到 connection -> 转化(？谁来做这个工作) 为 可以存入redis的类型
  - 存入 Connection
    - 把 connection 转化为数据类型， 存入redis

  两个部分： 
  - connection -> 数据结构
  - connection 数据结构 -> 活的连接(完整的生命周期)


  


