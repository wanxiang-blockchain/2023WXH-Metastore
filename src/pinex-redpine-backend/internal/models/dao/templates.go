package dao

//go:generate stringer -type=Role,ChainType,DaType,NetworkType,CloudVendor,ProverType,ResourceType -output=./model_stringers.go
type ChainType uint16

const (
	_ ChainType = iota
	OP_STACK
	STARK_NET
	POLYGON_ZKEVM
)

var ChainType_name = map[ChainType]string{
	OP_STACK:      "OpStack",
	STARK_NET:     "StarkNet",
	POLYGON_ZKEVM: "PolygonzkEVM",
}

var ChainType_value = map[string]ChainType{
	"OpStack":      OP_STACK,
	"StarkNet":     STARK_NET,
	"PolygonzkEVM": POLYGON_ZKEVM,
}

type DaType uint16

const (
	_ DaType = iota
	ETH
	META_STORE
)

var DaType_name = map[DaType]string{
	ETH:        "ETH",
	META_STORE: "MetaStore",
}

var DaType_value = map[string]DaType{
	"ETH":       ETH,
	"MetaStore": META_STORE,
}

type NetworkType uint16

const (
	_ NetworkType = iota
	DEV_NET
	TEST_NET
	MAIN_NET
)

var NetworkType_name = map[NetworkType]string{
	DEV_NET:  "Devnet",
	TEST_NET: "Testnet",
	MAIN_NET: "Mainnet",
}

var NetworkType_value = map[string]NetworkType{
	"Devnet":  DEV_NET,
	"Testnet": TEST_NET,
	"Mainnet": MAIN_NET,
}

type Role uint16

const (
	_ Role = iota
	POLYGON_ZKEVM_PROVER
	POLYGON_ZKEVM_NODE
	OP_STACK_NODE
	POLYGON_ZKEVM_DB
)

var Role_name = map[Role]string{
	POLYGON_ZKEVM_PROVER: "Prover",
	POLYGON_ZKEVM_NODE:   "Node",
	OP_STACK_NODE:        "Node",
	POLYGON_ZKEVM_DB:     "DB",
}

type CloudVendor uint8

const (
	_ CloudVendor = iota
	SURFER_CLOUD
	AWS
)

var CloudVendor_name = map[CloudVendor]string{
	SURFER_CLOUD: "SurferCloud",
	AWS:          "AWS",
}

var CloudVendor_value = map[string]CloudVendor{
	"SurferCloud": SURFER_CLOUD,
	"AWS":         AWS,
}

type ProverType uint8

const (
	NONE_PROVER ProverType = iota
	EXCLUSIVE
	SHARED
)

var ProverType_name = map[ProverType]string{
	EXCLUSIVE: "Exclusive",
	SHARED:    "Shared",
}

var ProverType_value = map[string]ProverType{
	"Exclusive": EXCLUSIVE,
	"Shared":    SHARED,
}

type DeploymentTemplate struct {
	BaseTemplateModel
	ChainType      ChainType
	CreationOrders []CreationOrder
	Dbs            []DbTemplate `gorm:"many2many:deployment_db;"`
	Nodes          []NodeTemplate
}

type DbTemplate struct {
	BaseTemplateModel
	CloudVendor
	Role
	DbType      string `gorm:"size:256"`
	MachineType string `gorm:"size:256"`
	Port        int
	Mode        string

	DiskTemplateID uint
	DiskTemplate   DiskTemplate
}

type NodeTemplate struct {
	BaseTemplateModel
	Role

	DeploymentTemplateID uint

	DiskTemplateID uint
	DiskTemplate   DiskTemplate

	MachineTemplateID uint
	MachineTemplate   MachineTemplate

	ImageTemplateID uint
	ImageTemplate   ImageTemplate

	Eip_TemplateID *uint
	EipTemplate    *EipTemplate
}

type MachineTemplate struct {
	BaseTemplateModel
	CloudVendor
	Platform    string `gorm:"size:256"`
	CoreNum     int
	Memory      int
	MachineType string `gorm:"size:256"`
}

type ImageTemplate struct {
	BaseTemplateModel
	CloudVendor
	ImageId string `gorm:"size:256"`
	Region  string `gorm:"size:256"`
	AZ      string `gorm:"size:256"`

	DiskTemplateID uint
	DiskTemplate   DiskTemplate

	Userdatas []Userdata
}

type DiskTemplate struct {
	BaseTemplateModel
	CloudVendor
	Type string `gorm:"size:256"`
	Size uint64
}

type Userdata struct {
	BaseTemplateModel
	ImageTemplateID uint
	DaType
	NetworkType
	ProverType
	Content string `gorm:"size:4096"`
	Args    []UserDataArg
}

type EipTemplate struct {
	BaseTemplateModel
	BindWidth int
}

type UserDataArg struct {
	BaseTemplateModel
	UserdataID uint64
	Position   uint8
}

type CreationOrder struct {
	BaseModel
	Role
	Sequence             int
	DeploymentTemplateID uint
	ShouldBlock          bool
}
