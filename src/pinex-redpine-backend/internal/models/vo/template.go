package vo

type HostDiskSet struct {
	IsBoot bool
	Size   int
}

type PgSQLInstance struct {
	DBVersion    string
	DiskSpace    int
	MachineType  string
	InstanceMode string
	ChargeType   string
	Price        float64
	Role         string
}

type HostInstance struct {
	CPU        int
	Memory     int
	DiskSet    []HostDiskSet
	ChargeType string
	Price      float64
	Role       string
}
type EIPInfo struct {
	Bandwidth  int
	ChargeType string
	Price      float64
}

type Template struct {
	HostInstances  []HostInstance
	PgSQLInstances []PgSQLInstance
	EIPSet         []EIPInfo
}
