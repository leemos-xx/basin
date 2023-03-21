package common

const (
	// Basin 项目名
	Basin = "basin"

	// CgroupName Cgroup名
	CgroupName = Basin + "-cgroup"

	// Running 容器状态为运行中
	Running = "running"
	// Stop 容器状态为结束
	Stop = "stopped"
	// Exit 容器状态为退出
	Exit = "exited"

	// IdLength 容器ID的默认长度
	IdLength = 10

	// ContainerDataUrl 容器数据主路径
	ContainerDataUrl = "/var/run/" + Basin + "/"
	// ContainerDataUrlFormat 用于根据容器名拼装数据路径
	ContainerDataUrlFormat = ContainerDataUrl + "%s/"
	// ConfigFileName 配置文件名
	ConfigFileName = "config.json"
	// LogFileName 日志文件名
	LogFileName = "container.log"

	// RootUrl 根路径
	RootUrl = "/root/"
	// LowerDirFormat lower层路径
	LowerDirFormat = RootUrl + "%s/lower"
	// UpperDirFormat upper层路径
	UpperDirFormat = RootUrl + "%s/upper"
	// WorkDirFormat work层路径
	WorkDirFormat = RootUrl + "%s/work"
	// MergedDirFormat merged层路径
	MergedDirFormat = RootUrl + "%s/merged"
	// OverlayFsFormat 拼接命令格式
	OverlayFsFormat = "lowerdir=%s,upperdir=%s,workdir=%s"

	// Perm0777 用户、组用户和其它用户都有读/写/执行权限
	Perm0777 = 0777
	// Perm0755 用户具有读/写/执行权限，组用户和其它用户具有读/写权限；
	Perm0755 = 0755
	// Perm0644 用户具有读/写权限，组用户和其它用户具只读权限；
	Perm0644 = 0644
	// Perm0622 用户具有读/写权限，组用户和其它用户具只写权限；
	Perm0622 = 0622

	MountPointIndex = 4
)
