- 目前主要是为了实现 RPC 的其他功能，有后续的话，根据 Gossip 协议改成集群的
- ccs 订阅 redis 所有服务的 key
- 当 ccs 上某一个服务的某一个节点挂了，首先是当前 ccs 的服务节点能感知到
  - 更新redis






- 服务A如何获取自己订阅的所有节点的IP列表
  - 服务A将自己注册到 control center 中后，会从 config center 获取自己订阅的所有服务的IP地址，并建立连接
  - 第一次建立好了连接后，如果这些 IP 有变更，怎么办
    - 断了还好说，可以自己感知到
    - 如果是新增了服务节点呢？control center 肯定是最先感知到的，如何通知到订阅这个服务的节点呢？