package resource

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/upgsql"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type createDeploymentReq struct {
	CreateDeploymentRequest
	dbPassword     string
	dbParamGroupId string

	projectId string
	vpcId     string
	subnetId  string
}

func (req *createDeploymentReq) create() ([]Resource, string, error) {
	if req.Type == OP_STACK {
		return req.createOpStack()
	} else if req.Type == POLYGON_ZKEVM {
		return req.createPolygonZKEvm()
	}
	return nil, "", errors.New("unknown template")
}

func (req *createDeploymentReq) createOpStack() ([]Resource, string, error) {
	vm := opNodeVm()
	req.populateCommonParams(vm.commonParams)
	vm.name = vm.nameBase + "_node"
	vm.keyPairId = req.KeyPairId
	err := vm.create()
	if err != nil {
		return nil, "", err
	}
	return []Resource{
		{
			Id:   vm.vmId,
			Type: OP_NODE,
		},
	}, "", nil
}

func (req *createDeploymentReq) createPolygonZKEvm() ([]Resource, string, error) {
	params := &commonParams{}
	req.populateCommonParams(params)
	vpcId, subnetId, ips, err := getRandomIps(params, 2)
	if err != nil {
		return nil, "", err
	}
	proverIp := ips[0]
	nodeIp := ips[1]
	log.Debugf("using network resource, vpc: %s, subnet: %s, proverIp: %s, nodeIp: %s", vpcId, subnetId, proverIp, nodeIp)
	params.vpcId = vpcId
	params.subnetId = subnetId
	// always use default user
	dbUser := "root"

	db, err := req.createDb(params)
	if err != nil {
		return nil, "", err
	}
	vmNode, err := req.createZkevmNode(params, proverIp, nodeIp, dbUser, db.password, db.ip, fmt.Sprint(db.port))
	if err != nil {
		return nil, "", err
	}

	err = waitForUhostReady(vmNode.vmId, params)
	if err != nil {
		return nil, "", err
	}
	var pendingTime = 20
	log.Debugf("pending for %d seconds, and then create zkevm node", pendingTime)
	// these vms cannot be created simultaneously
	// there's a bizarre dependence between prover and node
	time.Sleep(time.Duration(pendingTime) * time.Second)
	vmProver, err := req.createZkevmProver(params, proverIp, nodeIp, dbUser, db.password, db.ip)
	if err != nil {
		return nil, "", err
	}

	return []Resource{
		{
			Id:   db.dbId,
			Type: POLYGON_ZKEVM_DB,
		},
		{
			Id:   vmNode.vmId,
			Type: POLYGON_ZKEVM_NODE,
		},
		{
			Id:   vmProver.vmId,
			Type: POLYGON_ZKEVM_PROVER,
		},
	}, db.password, nil
}

func (req *createDeploymentReq) createDb(params *commonParams) (*dbResource, error) {
	req.dbPassword = randPassword(16)
	log.Debugf("db password: %s", req.dbPassword)
	db := pgDb()
	db.commonParams = params
	db.password = req.dbPassword

	err := db.create()
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (req *createDeploymentReq) createZkevmProver(params *commonParams, proverIp, nodeIp, dbUser, dbPassword, dbIp string) (*vmResource, error) {
	vmProver := zkEvmProverVm()
	vmProver.commonParams = params
	vmProver.name = vmProver.nameBase + "_prover"
	vmProver.keyPairId = req.KeyPairId
	vmProver.ip = proverIp
	vmProver.ud.args = append(vmProver.ud.args, nodeIp, dbIp, dbUser, dbPassword)
	log.Debugf("creating zkevm prover, prover instance name: %s, userdata args: {node instance ip: %s, db ip: %s, db user: %s, db password: %s}",
		vmProver.nameBase, nodeIp, dbIp, dbUser, dbPassword)
	err := vmProver.create()
	if err != nil {
		return nil, err
	}
	return &vmProver, nil
}

func (req *createDeploymentReq) createZkevmNode(params *commonParams, proverIp, nodeIp, dbUser, dbPassword, dbIp, dbPort string) (*vmResource, error) {
	vmNode := zkEvmNodeVm()
	vmNode.commonParams = params
	vmNode.name = vmNode.nameBase + "_node"
	vmNode.keyPairId = req.KeyPairId
	vmNode.ip = nodeIp
	log.Debugf("creating zkevm node, node instance name: %s, userdata args: {prover instance ip: %s, db user: %s, db password: %s, db ip: %s, db port: %s}",
		vmNode.nameBase, proverIp, dbUser, dbPassword, dbIp, dbPort)
	vmNode.ud.args = append(vmNode.ud.args, proverIp, dbUser, dbPassword, dbIp, dbPort)
	err := vmNode.create()
	if err != nil {
		return nil, err
	}
	return &vmNode, nil
}

func (req *createDeploymentReq) populateCommonParams(params *commonParams) {
	params.nameBase = req.Name
	params.region = req.Region
	params.zone = req.Az
	params.token.AccessKeyId = req.AccessKeyId
	params.token.SecretAccessKey = req.SecretAccessKey
}

type userdata struct {
	template string
	args     []interface{}
}

func (ud *userdata) String() string {
	return fmt.Sprintf(ud.template, ud.args...)
}

type commonParams struct {
	nameBase  string
	region    string
	zone      string
	projectId string
	vpcId     string
	subnetId  string
	token     Token
}

type vmDisk struct {
	size     int
	diskType string
}

type vmResource struct {
	*commonParams
	keyPairId string
	name      string
	// always using default values
	ud          *userdata
	cpu         int
	cpuPlatform string
	memory      int
	imgId       string
	rootDisk    vmDisk
	dataDisk    vmDisk
	ip          string

	// assigned after creating
	vmId string
}

func (r *vmResource) create() error {
	cfg := ucloud.NewConfig()
	cfg.UserAgent = config.USER_AGENT
	if config.GConf.ApiBaseUrl != "" {
		log.Debugf("using api base url: %s", config.GConf.ApiBaseUrl)
		cfg.BaseUrl = config.GConf.ApiBaseUrl
	}

	credential := auth.NewCredential()
	credential.PublicKey = r.token.AccessKeyId
	credential.PrivateKey = r.token.SecretAccessKey

	uhostClient := uhost.NewClient(&cfg, &credential)

	req := uhostClient.NewCreateUHostInstanceRequest()
	req.Name = ucloud.String(r.name)
	req.Zone = ucloud.String(r.zone)
	req.Region = ucloud.String(r.region)
	if r.vpcId != "" {
		req.VPCId = ucloud.String(r.vpcId)
	}
	if r.subnetId != "" {
		req.SubnetId = ucloud.String(r.subnetId)
	}
	if r.projectId != "" {
		req.ProjectId = ucloud.String(r.projectId)
	}
	req.PrivateIp = append(req.PrivateIp, r.ip)
	req.Disks = make([]uhost.UHostDisk, 2)
	req.MachineType = ucloud.String("O")
	req.Disks[0] = uhost.UHostDisk{
		Size:   ucloud.Int(r.rootDisk.size),
		IsBoot: ucloud.String("True"),
		Type:   ucloud.String(r.rootDisk.diskType),
	}
	req.Disks[1] = uhost.UHostDisk{
		Size:   ucloud.Int(r.dataDisk.size),
		IsBoot: ucloud.String("False"),
		Type:   ucloud.String(r.dataDisk.diskType),
	}
	req.ImageId = ucloud.String(r.imgId)
	req.LoginMode = ucloud.String("KeyPair")
	req.KeyPairId = &r.keyPairId
	req.MinimalCpuPlatform = ucloud.String(r.cpuPlatform)
	req.CPU = ucloud.Int(r.cpu)
	req.Memory = ucloud.Int(r.memory)
	if r.ud != nil {
		userdataString := r.ud.String()
		userDataBase64 := base64.StdEncoding.EncodeToString([]byte(userdataString))
		req.UserData = ucloud.String(userDataBase64)
	}

	log.Debugw("creating uhost instance",
		"image id", r.imgId,
		"ip", r.ip,
		"KeyPairId", req.KeyPairId,
		"machine type", req.MachineType,
		"cpu", req.CPU,
		"memory", req.Memory,
		"vpc", req.VPCId,
		"subnet", req.SubnetId,
		"disks", req.Disks,
	)
	resp, err := uhostClient.CreateUHostInstance(req)
	if err != nil {
		return err
	}
	if resp.GetRetCode() != 0 {
		return fmt.Errorf("failed creating uhost: %s", resp.GetMessage())
	}
	if len(resp.UHostIds) != 1 {
		return errors.New("get vm id failed")
	}
	r.vmId = resp.UHostIds[0]
	log.Debugf("vm creation done, vm id: %s", r.vmId)
	return nil
}

type dbResource struct {
	*commonParams
	password    string
	version     string
	machineType string
	port        int
	dbMode      string
	diskSize    int

	paramGroupId int

	// populated after Create is called
	ip   string
	dbId string
}

// blocking until the database is ready
func (r *dbResource) create() error {
	cfg := ucloud.NewConfig()
	cfg.UserAgent = config.USER_AGENT
	if config.GConf.ApiBaseUrl != "" {
		cfg.BaseUrl = config.GConf.ApiBaseUrl
		log.Debugf("using api base url: %s", config.GConf.ApiBaseUrl)
	}

	credential := auth.NewCredential()
	credential.PublicKey = r.token.AccessKeyId
	credential.PrivateKey = r.token.SecretAccessKey

	client := upgsql.NewClient(&cfg, &credential)

	listParamsReq := client.NewGenericRequest()
	listParamsReq.SetAction("ListUPgSQLParamTemplate")
	listParamsReq.SetZone(r.zone)
	listParamsReq.SetRegion(r.region)
	listParamsReq.SetProjectId(r.projectId)

	log.Debugf("listing parameter groups")
	listResp, err := client.GenericInvoke(listParamsReq)
	if err != nil {
		return err
	}
	if listResp.GetRetCode() != 0 {
		return fmt.Errorf("failed list param template: %s", listResp.GetMessage())
	}
	iDataSet := listResp.GetPayload()["Data"]
	dataSet, ok := iDataSet.([]interface{})
	if !ok {
		return errors.New("unmarshal param template response dataset failed")
	}

	log.Debugf("get list response, finding default parameter group id for db version: %s", r.version)
	for _, data := range dataSet {
		if attrs, ok := data.(map[string]interface{}); ok {
			modifiable, ok := attrs["Modifiable"].(bool)
			if !ok {
				return errors.New("unmarshal param template response dataset failed")
			}
			if modifiable {
				continue
			}
			dbVersion, ok := attrs["DBVersion"].(string)
			if !ok {
				return errors.New("unmarshal param template response dataset failed")
			}
			if dbVersion != r.version {
				continue
			}
			var groupId int
			groupIdF, ok := attrs["GroupID"].(float64)
			if ok {
				groupId = int(groupIdF)
			} else {
				groupId, ok = attrs["GroupID"].(int)
				if !ok {
					return errors.New("unmarshal param template response dataset failed")
				}
			}
			r.paramGroupId = groupId
			log.Debugf("parameter group found: %d", groupId)
		}
	}
	if r.paramGroupId == 0 {
		return errors.New("get parameter group id failed")
	}

	req := client.NewCreateUPgSQLInstanceRequest()
	req.Name = ucloud.String(r.nameBase)
	req.Zone = ucloud.String(r.zone)
	req.Region = ucloud.String(r.region)
	if r.vpcId != "" {
		req.VPCID = ucloud.String(r.vpcId)
	}
	if r.subnetId != "" {
		req.SubnetID = ucloud.String(r.subnetId)
	}
	if r.projectId != "" {
		req.ProjectId = ucloud.String(r.projectId)
	}
	req.ParamGroupID = ucloud.Int(r.paramGroupId)
	req.DBVersion = ucloud.String(r.version)
	req.AdminPassword = ucloud.String(r.password)

	req.MachineType = ucloud.String(r.machineType)
	req.Port = ucloud.Int(r.port)
	req.DiskSpace = ucloud.String(fmt.Sprint(r.diskSize))
	req.InstanceMode = ucloud.String(r.dbMode)

	log.Debugw("creating pgsql instance",
		"disk space", req.DiskSpace,
		"port", req.Port,
		"instance mode", req.InstanceMode,
		"machine type", req.MachineType,
		"password", req.AdminPassword,
		"parameter group", req.ParamGroupID,
		"vpc", req.VPCID,
		"subnet", req.SubnetID,
	)

	resp, err := client.CreateUPgSQLInstance(req)
	if err != nil {
		return err
	}
	if resp.GetRetCode() != 0 {
		return fmt.Errorf("failed creating pgsql: %s", resp.GetMessage())
	}
	log.Debug("pgsql creation done, pending pgsql initialization")

	dbId := resp.InstanceID
	listReq := client.NewListUPgSQLInstanceRequest()
	listReq.Region = ucloud.String(r.region)
	listReq.Zone = ucloud.String(r.zone)
	if r.projectId != "" {
		listReq.ProjectId = ucloud.String(r.projectId)
	}

	for {
		listResp, err := client.ListUPgSQLInstance(listReq)
		if err != nil {
			return err
		}
		if listResp.GetRetCode() != 0 {
			return fmt.Errorf("failed list pgsql instance: %s", listResp.GetMessage())
		}
		for _, meta := range listResp.DataSet {
			if meta.InstanceID == dbId {
				if meta.State == "Fail" {
					return errors.New("created db failed")
				}
				if meta.State == "Running" {
					r.ip = meta.IP
					r.dbId = dbId
					log.Debugf("db initialization done, db ip: %s, db id: %s", r.ip, r.dbId)
					return nil
				}
				log.Debugf("found target pgsql from response list, current instance status: %s", meta.State)
			}
		}
		time.Sleep(20 * time.Second)
	}
}

func waitForUhostReady(id string, params *commonParams) error {
	cfg := ucloud.NewConfig()
	cfg.UserAgent = config.USER_AGENT
	if config.GConf.ApiBaseUrl != "" {
		cfg.BaseUrl = config.GConf.ApiBaseUrl
	}

	credential := auth.NewCredential()
	credential.PublicKey = params.token.AccessKeyId
	credential.PrivateKey = params.token.SecretAccessKey

	log.Debug("pending uhost instance initiation")
	client := uhost.NewClient(&cfg, &credential)
	describeReq := client.NewDescribeUHostInstanceRequest()
	describeReq.Region = ucloud.String(params.region)
	describeReq.Zone = ucloud.String(params.zone)
	describeReq.UHostIds = []string{id}
	describeReq.ProjectId = ucloud.String(params.projectId)
	for {
		describeResp, err := client.DescribeUHostInstance(describeReq)
		if err != nil {
			return err
		}
		if describeResp.GetRetCode() != 0 {
			return fmt.Errorf("failed describe uhost instance: %s", describeResp.GetMessage())
		}
		if len(describeResp.UHostSet) == 1 {
			instanceMeta := describeResp.UHostSet[0]
			log.Debugf("found target uhost from response list, current instance status: %s", instanceMeta.State)
			if instanceMeta.State == "Running" {
				log.Debugf("uhost initialization done, id: %s", instanceMeta.UHostId)
				return nil
			}
		}
		time.Sleep(time.Second * 10)
	}
}
