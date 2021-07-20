cmapp
---

* chain and machine application

### RoadMap

- [x] 完善`websocket`库，包括客户端和服务端
- [x] 实现`agentfw`，实现使用`ws`协议来和本地的`grpc`服务端交互
- [ ] 实现`agentfw`和`core`进行交互
- [ ] `core`中实现调用`machine`的逻辑，并且开放`machien api`接口给外部调用
- [ ] 实现k8s主机驱动
- [ ] 尝试通过k8s主机驱动创建`example-app`
- [ ] 开始实现`hpc`链驱动