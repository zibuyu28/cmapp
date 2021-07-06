###  mrobot 设计

* 主机驱动，有两部分功能
    1. 是一个 grpc client，在 create 这个命令中会启动对应驱动的 grpc server, 驱动可以由用户指定放在固定的目录中即可，
       参考 docker/machine 的驱动实现方式
       1. 目前准备内置 k8s 驱动。 驱动实现在 drivers 这个目录中
       2. 所以目前 mrobot 也是一个二进制的主机驱动，启动的时候不加任何 flag，设置对应的环境变量即可
    2. 也是一个 machine agent，在创建主机的时候会安装对应的 mrobot 到对应的远程主机（这是一个难点，具体实现需要由驱动提供
       ，包括k8s和vb的实现，注意是在 core 中直接调用驱动的 InstallMRobot 方法，然后驱动需要实现直接调用sdk获取其他方法来
       在对应的主机上安装并且启动 mrobot），mrobot 同是也是一个 agent 需要提供 ag 包 MachineAPI 相关接口的功能
    

* 思考：mrobot 同是也是一个 agent 需要提供 ag 包 MachineAPI 相关接口的功能
    1. 因为 ag 的功能也是驱动提供的
    2. mrobot 是提供的中间层，最后调用还是调用实现的主机驱动的 grpc。
    3. 另一种方案：直接把驱动二进制丢过去，然后支持 ./driver-xxx start 命令（所有需要的参数全部再环境变量中获取）