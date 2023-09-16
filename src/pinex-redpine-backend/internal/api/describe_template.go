package api

import (
	"fmt"
	"strings"

	web3errors "github.com/MetaDataLab/web3-console-backend/internal/errors"
	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/models/vo"
	v2 "github.com/MetaDataLab/web3-console-backend/internal/resource/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

func init() {
	Register(DESCRIBE_DEPLOYMENT_TEMPLATE_LABEL, NewDescribeTemplateTask, COOKIEAUTH)
}

type Token = v2.AccessToken

type DeploymentParamGroup struct {
	Chain       string `validate:"required"`
	NetworkType string `validate:"required"`
	ProverType  string
	DA          string `validate:"required"`
	CloudVendor string `validate:"required"`
	Region      string `validate:"required"`
	Zone        string `validate:"required"`
	AccessKey   Token  `mapstructure:",squash"`
}

type parsedParams struct {
	chain       dao.ChainType
	networkType dao.NetworkType
	proverType  dao.ProverType
	da          dao.DaType
	cloudVendor dao.CloudVendor
}

var _ Task = (*DescribeTemplateTask)(nil)

type DescribeTemplateRequest = DeploymentParamGroup

type DescribeTemplateTask AnyRequest[DescribeTemplateRequest]

type DescribeTemplateResp struct {
	BaseResponse
	vo.Template
}

func NewDescribeTemplateTask(data *map[string]interface{}) (Task, error) {
	req := DescribeTemplateTask{}
	err := mapstructure.Decode(*data, &req)
	if err != nil {
		return nil, err
	}
	req.BaseRequest.RequestUUID = (*data)[REQUEST_UUID].(string)
	validate := validator.New()
	err = validate.Struct(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (t *DescribeTemplateTask) Run(*gin.Context) (Response, error) {
	if isInvalidToken(&t.Body.AccessKey) {
		return nil, web3errors.ParamsError("invalid token")
	}
	tmpl, dep, _, err := doLoadTemplate(&t.Body, "", "", t.RequestUUID)
	if err != nil {
		return nil, err
	}
	priceMap, err := dep.LoadPrice()
	if err != nil {
		return nil, err
	}
	respTmpl := vo.Template{
		EIPSet:         []vo.EIPInfo{},
		HostInstances:  []vo.HostInstance{},
		PgSQLInstances: []vo.PgSQLInstance{},
	}
	for i := range tmpl.Nodes {
		node := tmpl.Nodes[i]
		if len(node.ImageTemplate.Userdatas) == 0 {

		}
		priceT := priceMap[node.Role]
		hostInfo := vo.HostInstance{
			CPU:        node.MachineTemplate.CoreNum,
			Memory:     node.MachineTemplate.Memory,
			ChargeType: "Month",
			Price:      priceT.InstancePrice,
			Role:       roleToVoString(node.Role),
			DiskSet: []vo.HostDiskSet{
				{IsBoot: true, Size: int(node.ImageTemplate.DiskTemplate.Size)},
				{IsBoot: false, Size: int(node.DiskTemplate.Size)},
			},
		}
		respTmpl.HostInstances = append(respTmpl.HostInstances, hostInfo)
		if node.EipTemplate != nil {
			eipInfo := vo.EIPInfo{
				Bandwidth:  node.EipTemplate.BindWidth,
				ChargeType: "Month",
				Price:      priceT.NetworkPrice,
			}
			respTmpl.EIPSet = append(respTmpl.EIPSet, eipInfo)
		}
	}
	for i := range tmpl.Dbs {
		db := tmpl.Dbs[i]
		priceT := priceMap[db.Role]
		dbInfo := vo.PgSQLInstance{
			DBVersion:    db.DbType,
			DiskSpace:    int(db.DiskTemplate.Size),
			MachineType:  db.MachineType,
			InstanceMode: db.Mode,
			ChargeType:   "Month",
			Price:        priceT.InstancePrice,
			Role:         roleToVoString(db.Role),
		}
		respTmpl.PgSQLInstances = append(respTmpl.PgSQLInstances, dbInfo)
	}
	resp := &DescribeTemplateResp{
		BaseResponse: BaseResponse{
			Action:      DESCRIBE_DEPLOYMENT_TEMPLATE_LABEL + "Response",
			RequestUUID: t.BaseRequest.RequestUUID,
		},
		Template: respTmpl,
	}
	return resp, nil
}

func doLoadTemplate(t *DeploymentParamGroup, name string, keypairId string, requestId string) (*dao.DeploymentTemplate, *v2.Deployment, *parsedParams, error) {
	chain, err := parseChainType(t.Chain)
	if err != nil {
		return nil, nil, nil, web3errors.ParamsError(err)
	}
	if t.DA == "" {
		t.DA = "Eth"
	}
	daType, err := parseDaType(t.DA)
	if err != nil {
		return nil, nil, nil, web3errors.ParamsError(err)
	}
	networkType, err := parseNetworkType(t.NetworkType)
	if err != nil {
		return nil, nil, nil, web3errors.ParamsError(err)
	}
	proverType, err := parseProverType(t.ProverType)
	if err != nil {
		return nil, nil, nil, web3errors.ParamsError(err)
	}
	cloudVendor, err := parseCloudVendor(t.CloudVendor)
	if err != nil {
		return nil, nil, nil, web3errors.ParamsError(err)
	}

	dep := v2.NewDeployment(chain, daType, networkType, proverType, cloudVendor,
		t.Region, t.Zone, t.AccessKey, requestId, v2.WithKeyPairId(keypairId), v2.WithName(name))
	tmpl, err := dep.LoadTemplate()
	if err != nil {
		return nil, nil, nil, err
	}
	// every template at least has a node
	if tmpl == nil || tmpl.Nodes == nil {
		return nil, nil, nil, web3errors.ParamsError(fmt.Errorf("deployment for parameter conbination [%s, %s, %s, %s, %s] not found", chain, daType, networkType, proverType, cloudVendor))
	}
	// and every node at least has a userdata
	for i := range tmpl.Nodes {
		node := tmpl.Nodes[i]
		if node.ImageTemplate.ImageId == "" {
			return nil, nil, nil, web3errors.ParamsError(fmt.Errorf("deployment not available for region: %s, zone: %s", t.Region, t.Zone))
		}
		if len(node.ImageTemplate.Userdatas) == 0 {
			return nil, nil, nil, web3errors.ParamsError(fmt.Errorf("deployment for parameter conbination [%s, %s, %s, %s, %s] not found", chain, daType, networkType, proverType, cloudVendor))
		}
	}
	params := &parsedParams{
		chain:       chain,
		networkType: networkType,
		proverType:  proverType,
		da:          daType,
		cloudVendor: cloudVendor,
	}
	return tmpl, dep, params, nil
}

func parseChainType(v string) (dao.ChainType, error) {
	switch strings.ToLower(v) {
	case formEnumString(dao.POLYGON_ZKEVM):
		return dao.POLYGON_ZKEVM, nil
	case formEnumString(dao.OP_STACK):
		return dao.OP_STACK, nil
	case formEnumString(dao.STARK_NET):
		return dao.OP_STACK, nil
	}
	return 0, fmt.Errorf("unknown chain type: %s", v)
}

func parseCloudVendor(v string) (dao.CloudVendor, error) {
	switch strings.ToLower(v) {
	case formEnumString(dao.SURFER_CLOUD):
		return dao.SURFER_CLOUD, nil
	case formEnumString(dao.AWS):
		return dao.AWS, nil
	}
	return 0, fmt.Errorf("unknown cloud vendor type: %s", v)
}

func parseProverType(v string) (dao.ProverType, error) {
	if v == "" {
		return dao.ProverType(dao.NONE_PROVER), nil
	}
	switch strings.ToLower(v) {
	case formEnumString(dao.EXCLUSIVE):
		return dao.EXCLUSIVE, nil
	case formEnumString(dao.SHARED):
		return dao.SHARED, nil
	}
	return 0, fmt.Errorf("unknown prover type: %s", v)
}

func parseNetworkType(v string) (dao.NetworkType, error) {
	switch strings.ToLower(v) {
	case formEnumString(dao.DEV_NET):
		return dao.DEV_NET, nil
	case formEnumString(dao.TEST_NET):
		return dao.TEST_NET, nil
	case formEnumString(dao.MAIN_NET):
		return dao.MAIN_NET, nil
	}
	return 0, fmt.Errorf("unknown network type: %s", v)
}

func parseDaType(v string) (dao.DaType, error) {
	switch strings.ToLower(v) {
	case formEnumString(dao.ETH):
		return dao.ETH, nil
	case formEnumString(dao.META_STORE):
		return dao.META_STORE, nil
	}
	return 0, fmt.Errorf("unknown da type: %s", v)
}

func roleToVoString(r dao.Role) string {
	return formEnumString[dao.Role](r)
}

func formEnumString[T fmt.Stringer](v T) string {
	return strings.ToLower(strings.ReplaceAll(v.String(), "_", ""))
}
