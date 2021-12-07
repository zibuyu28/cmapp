cmapp
---

* chain and machine application

### RoadMap

- [x] 完善`websocket`库，包括客户端和服务端
- [x] 实现`agentfw`，实现使用`ws`协议来和本地的`grpc`服务端交互
- [x] 实现`agentfw`和`core`进行交互
- [x] `core`中实现调用`machine`的逻辑，并且开放`machien api`接口给外部调用
- [x] 实现k8s主机驱动
- [x] 尝试通过k8s主机驱动创建`example-app`
- [x] 开始实现`fabric`链驱动
----

### 2021.08.23
- [x] 调研`virtualbox`远程`sdk`
- [x] 实现`virtualbox`的本地驱动
- [x] 实现`virtualbox`远程驱动 
- [x] 增加`core`中创建主机的`http`接口
- [x] 增加`core`中查询主机、更新主机、删除主机信息接口
- [x] 测试创建主机
- [x] 增加`fabric`链驱动（TODO: 有很多工作要做）
- [x] 首先还是尝试创建`vb`中的`fabric`链

### 2021.12.05
- [x] 链入参动态化
- [x] `vb`主机启动进程需要异步子进程化，不能让`package`里面有异步命令
- [ ] 端口检测
- [ ] `kubernetes`主机测试
- [ ] `kubernetes`上部署`fabric`

-----
#### 创建链流程
* 创建一台`virtualbox`主机
* 创建完成之后，使用链驱动创建链，此时会选择节点部署的主机
* 驱动先解析链的参数，然后创建链的信息
* 调用`core`提供的一些主机的接口，`newapp,setenv`之类的
* 此时`core`中会有一个`app`的记录（或者记录到表中），结构和`worker0`中的很相似；
  所有链驱动调用主机相关的接口都会经过这个模块，并且构建出`app`记录，然后`core`会调用
  主机驱动中的接口来达到真正地创建节点并启动
* 能达到`NxN`就成功了

