# Webserver开发
这是项目的webserver开发的部分，整个项目的架构图如下所示：
![架构图](image/architecture.jpg)

这里简要介绍架构图中的几种技术
1. redis缓存：redis是一个键值对的内存数据库，是一种高速缓存方案，用于加速程序对数据层的操作，如果你想更加深入地了解redis的话（当然，对本次开发来说没啥必要），可以点击[链接](https://redis.io/topics/data-types-intro)。
2. redisqueue：这是一个消息队列，在高流量下，使用消息队列的原因请参考该[链接](https://zhuanlan.zhihu.com/p/53602080)。

该项目从开发流程上，分为以下两个部分
1. 读数据：利用redis高速缓存加速数据读。
    - 当redis中不存在数据时，从数据库中读取数据，并将读取到的数据写入redis中。
    - 当redis中有数据时，直接返回redis中的结果。
2. 写数据：利用消息队列，调整数据库的写操作的流量，本次仅只对优惠券剩余数目为例子。
    - 在写数据之前，需要从redis中，检查商家剩余优惠券数目。
    - 如果商家还有优惠券，则执行利用消息队列执行写操作。
    - 如果商家已经没有优惠券了，则直接返回一个错误码。

## Todo List
- 接口文档，参照[链接](https://www.eolinker.com/#/share/index?shareCode=1P4kre)
- 说明文档，参照[链接](https://shimo.im/docs/9vtcTDHJDYQr8xVp/read)

***PS：本次对数据库的写操作，只有<u>用户获取优惠券</u>这个接口需要调用消息队列，其余写操作，包括商家用户注册，商家发放优惠券这些接口，直接对数据库进行写即可***

这里主要实现接口文档中的接口，按照任务进行划分如下
- [ ] 任务1：用户注册（包括商家和普通用户），用户和商家登录，这两个接口的实现。——仅依赖于[数据库](https://github.com/2019-sysu-group-project/project-database)。
- [ ] 任务2：商家新建优惠券，获取优惠券信息，这两个接口的实现。——仅依赖于[数据库](https://github.com/2019-sysu-group-project/project-database)。
- [ ] 任务3：用户获取优惠券。——依赖于[数据库](https://github.com/2019-sysu-group-project/project-database)+[消息队列](https://github.com/2019-sysu-group-project/message-queue)

点击数据库和消息队列的链接，查看数据库和消息队列的部署方式(没有部署方式代表还未开发完毕)。

在开发之前，如果你不熟悉golang的web开发，点击[链接](https://github.com/astaxie/build-web-application-with-golang)开始一个简单的入门，包含golang的相关概念，以及web开发的一些技术。

Golang第三方包下载会非常慢，因此，使用七牛云来加速go get获取第三方库的过程，[参考链接](https://github.com/goproxy/goproxy.cn)。


### 各项任务细节
开发使用golang的[gbin框架](https://github.com/gin-gonic/gin)。

任务1，用户注册，用户登录：
1. 使用JWT作为用户认证的方式，如果你不清楚JWT，参考[链接](http://www.ruanyifeng.com/blog/2018/07/json_web_token-tutorial.html)
2. 需要额外编写JWT认证的函数。具体见代码部分。

关于任务3，用户获取优惠券，定义流程如下：
1. 首先从redis里，知道redis是否有该商家的
