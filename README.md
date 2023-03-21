# basin
basin是轻量级的linux容器，用于理解docker的基本概念。

## 1.准备
### 1.1 编译
要求linux系统以及golang 1.17+，编译完成后会得到basin的可执行文件。
```bash
$ git clone https://github.com/leemos-xx/basin.git
$ cd basin
$ go build
```

### 1.2 镜像
basin未支持镜像制作，测试时可以使用docker镜像，以busybox镜像为例，首先拉取镜像。
```bash
$ docker pull busybox
```

通过该镜像运行docker容器，运行成功后会打印出容器的ID。
```bash
$ docker run -d busybox top -b
8b2843e6c0fc62b2107a0078938593de215f324b9fb4a6f8ddd421663c3a7612 
```

从容器中直接导出镜像的压缩包，执行完成后可以看到当前目录下打包出`busybox.tar`。
```bash
$ docker export -o busybox.tar 8b2843e6c0fc62b2107a0078938593de215f324b9fb4a6f8ddd421663c3a7612
```

将`busybox.tar`拷贝到`/root`路径下，basin将从该路径查询镜像。
```bash
$ mv busybox.tar /root
```

## 2 常用命令
```bash
$ ./basin
NAME:
   basin - A new cli application

USAGE:
   basin [global options] command [command options] [arguments...]

COMMANDS:
   init     Init container process run user's process in container. Do not call it outside
   run      Run a command in a new lightweight container
   ps       list all the containers
   logs     print logs of a container
   stop     stop a container
   rm       remove unused containers
   network  container network commands
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help

```
### 2.1 启动容器
`basin run`可以启动一个容器，通过`-it`可以声明启用TTY。可以看到通过如下命令，可以启动容器并在容器中执行`/bin/sh`命令。

在进入容器后执行`ps`命令，可以看到`/bin/sh`命令的PID为1，与宿主机的命名空间是隔离的。
```bash
root@ubuntu $ ./basin run -it -name busybox-example busybox /bin/sh
$ ps
PID   USER     TIME  COMMAND
    1 root      0:00 /bin/sh
    7 root      0:00 ps
```

### 2.2 容器列表
首先通过`bash run -d`后台启动一个容器，然后通过`basin ps`可以查看当前的容器信息。
```bash
$ ./basin run -d -name busybox-example busybox top -b
$ ./basin ps
ID           NAME              PID         STATUS      COMMAND     CREATED
1083530930   busybox-example   37045       running     top -b      2023-02-10 20:31:38
```

### 2.3 停止容器
可以通过指定容器名来停止容器。
```bash
$ ./basin stop busybox-example
$ ./basin ps
ID           NAME              PID         STATUS      COMMAND     CREATED
1083530930   busybox-example               stopped     top -b      2023-02-10 20:31:38
```

### 2.4 删除容器
可以通过指定容器名来删除容器。
```bash
$ ./basin rm busybox-example
$ ./basin ps
ID           NAME              PID         STATUS      COMMAND     CREATED
```

### 2.5 创建&加入容器网络
`basin network`是网络相关命令，支持`create`、`ps`和`remove`操作。在创建完网络后，通过`basin run`的`-network`参数可以指定容器要加入的网络。

ps: 目前仅支持driver为bridge的模式。
```bash
$ ./basin network create --driver bridge --subnet 173.1.1.0/24 basin0
$ ./basin network ps
NAME        IpRange        Driver
basin0      173.1.1.1/24   bridge
$ ./basin run -d -name busybox-example -network basin0 busybox top -b
$ ./basin stop busybox-example
$ ./basin rm busybox-example
$ ./basin network rm basin0
```

### 2.6 资源限制
通过`-cpu`、`-cpuset`和`-mem`可以进行资源限制。
```bash
$ ./basin run -d -name busybox-example -cpu 10000 busybox top -b
```

## 3 主要流程
以如下容器为例进行分析，在执行完如下命令后，可以进入容器。
```bash
$ ./basin network create --driver bridge --subnet 173.1.1.0/24 basin0
$ ./basin run -d -name busybox-example -network basin0 busybox /bin/sh
```

### 3.1 命名空间
在宿主机上查看进程号，可以看到容器进程`/bin/sh`是`basin`的子进程。
```bash
root@ubuntu $ ps -ef | grep /bin/sh
root       44106    2829  0 21:13 pts/1    00:00:00 ./basin run -it -name busybox-example busybox /bin/sh
root       44114   44106  0 21:13 pts/1    00:00:00 /bin/sh 
```

在宿主机查看两个进程的namespace，可以看到是不同的。
```bash
root@ubuntu $ ls -la /proc/44106/ns/
总用量 0
dr-x--x--x 2 root root 0  2月 10 21:16 .
dr-xr-xr-x 9 root root 0  2月 10 21:13 ..
lrwxrwxrwx 1 root root 0  2月 10 21:16 cgroup -> 'cgroup:[4026531835]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 ipc -> 'ipc:[4026531839]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 mnt -> 'mnt:[4026531841]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 net -> 'net:[4026531840]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 pid -> 'pid:[4026531836]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 pid_for_children -> 'pid:[4026531836]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 time -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 time_for_children -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 user -> 'user:[4026531837]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 uts -> 'uts:[4026531838]'

root@ubuntu $ ls -la /proc/44114/ns/
总用量 0
dr-x--x--x 2 root root 0  2月 10 21:16 .
dr-xr-xr-x 9 root root 0  2月 10 21:13 ..
lrwxrwxrwx 1 root root 0  2月 10 21:16 cgroup -> 'cgroup:[4026531835]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 ipc -> 'ipc:[4026532269]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 mnt -> 'mnt:[4026532266]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 net -> 'net:[4026532271]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 pid -> 'pid:[4026532270]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 pid_for_children -> 'pid:[4026532270]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 time -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 time_for_children -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 user -> 'user:[4026531837]'
lrwxrwxrwx 1 root root 0  2月 10 21:16 uts -> 'uts:[4026532268]'
```

### 3.2 网络
首先在宿主机通过`ifconfig`查看设备。
```bash
root@ubuntu $ ifconfig
30115: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet6 fe80::38f2:48ff:fe20:3667  prefixlen 64  scopeid 0x20<link>
        ether 3a:f2:48:20:36:67  txqueuelen 1000  (以太网)
        RX packets 7  bytes 586 (586.0 B)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 20  bytes 2723 (2.7 KB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

basin0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet6 fe80::2cdf:60ff:fef9:30b5  prefixlen 64  scopeid 0x20<link>
        ether 2e:df:60:f9:30:b5  txqueuelen 1000  (以太网)
        RX packets 46  bytes 3000 (3.0 KB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 47  bytes 6164 (6.1 KB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
```

然后在容器中查看网络设备。
```bash
$ ifconfig
cif-30115 Link encap:Ethernet  HWaddr AA:CE:85:FD:56:85  
          inet addr:173.1.1.7  Bcast:173.1.1.255  Mask:255.255.255.0
          inet6 addr: fe80::a8ce:85ff:fefd:5685/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:16 errors:0 dropped:0 overruns:0 frame:0
          TX packets:6 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000 
          RX bytes:2106 (2.0 KiB)  TX bytes:516 (516.0 B)

lo        Link encap:Local Loopback  
          inet addr:127.0.0.1  Mask:255.0.0.0
          inet6 addr: ::1/128 Scope:Host
          UP LOOPBACK RUNNING  MTU:65536  Metric:1
          RX packets:0 errors:0 dropped:0 overruns:0 frame:0
          TX packets:0 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000 
          RX bytes:0 (0.0 B)  TX bytes:0 (0.0 B)
```

可以看到宿主机和basin容器间的网络是通过网桥打通的。