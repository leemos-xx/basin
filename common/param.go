package common

type RunParam struct {
	TTY               bool
	ContainerName     string
	Envs              []string
	Network           string
	PortMapping       []string
	Volume            string
	ImageName         string
	CgroupConfig      *CgroupParam
	ContainerCommands []string
}

type CgroupParam struct {
	CpuCfsQuota int
	CpuSet      string
	// TODO ?
	CpuShare    string
	MemoryLimit string
}
