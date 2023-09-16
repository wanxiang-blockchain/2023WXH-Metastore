package dao

type ResourceType uint8

const (
	_ ResourceType = iota
	VM
	DB
)

type Resource struct {
	BaseModel
	DeploymentID uint

	ResourceId   string `gorm:"size:256"`
	ResourceType ResourceType
	Role         Role
}

type Deployment struct {
	BaseTemplateModel
	Address     string
	ChainType   ChainType
	NetworkType NetworkType
	Da          DaType
	ProverType  ProverType
	CloudVendor CloudVendor
	Region      string
	Az          string
	Status      string `gorm:"size:256"`
	Message     string `gorm:"size:256"`
	Resources   []Resource
}
