package api

import (
	"github.com/MetaDataLab/web3-console-backend/internal/db"
	"github.com/MetaDataLab/web3-console-backend/internal/errors"
	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

func init() {
	Register(LIST_DEPLOYMENTS_LABEL, NewListDeploymentsTask, COOKIEAUTH)
}

type ListDeploymentsRequest struct {
	BaseRequest
	Address       string `mapstructure:"Address"  validate:"required"`
	DeploymentIds []uint `mapstructure:"DeploymentIds"`
}

type Deployment struct {
	DeploymentId   uint
	DeploymentName string
	Address        string
	Chain          string
	NetworkType    string
	DA             string
	ProverType     string
	CloudVendor    string
	Zone           string
	Status         string
	CreateTime     int64
	ResourceSet    []ResourceInfo
	Message        string
}

type ResourceInfo struct {
	ResourceId   string
	ResourceType string
	Role         string
}

type ListDeploymentsResponse struct {
	BaseResponse
	Deployments []Deployment
}

type ListDeploymentsTask struct {
	Request  *ListDeploymentsRequest
	Response *ListDeploymentsResponse
}

func NewListDeploymentsRequest(data *map[string]interface{}) (*ListDeploymentsRequest, error) {
	req := &ListDeploymentsRequest{}
	err := mapstructure.Decode(*data, &req)
	if err != nil {
		return nil, err
	}
	req.BaseRequest.RequestUUID = (*data)["RequestUUID"].(string)

	return req, nil
}

func NewListDeploymentsResponse(sessionId string) *ListDeploymentsResponse {
	return &ListDeploymentsResponse{
		BaseResponse: BaseResponse{
			Action:      LIST_DEPLOYMENTS_LABEL + "Response",
			RequestUUID: sessionId,
		},
	}
}

func NewListDeploymentsTask(data *map[string]interface{}) (Task, error) {
	req, err := NewListDeploymentsRequest(data)
	if err != nil {
		return nil, err
	}
	task := &ListDeploymentsTask{
		Request:  req,
		Response: NewListDeploymentsResponse(req.BaseRequest.RequestUUID),
	}

	validate := validator.New()
	err = validate.Struct(task.Request)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (task *ListDeploymentsTask) Run(c *gin.Context) (Response, error) {

	deps, err := db.ListDeployment([]uint(task.Request.DeploymentIds), task.Request.Address)
	if err != nil {
		log.Errorf("%s, db.ListDeployment failed, err %s", task.Request.RequestUUID, err.Error())
		task.Response.BaseResponse.RetCode = int(errors.ListDeploymentsFailed().Code())
		task.Response.BaseResponse.Message = errors.ListDeploymentsFailed().String()
		return task.Response, nil
	}

	var deployments = []Deployment{}

	for _, dep := range deps {
		var rs = []ResourceInfo{}

		for _, r := range dep.Resources {
			rs = append(rs, ResourceInfo{
				ResourceId:   r.ResourceId,
				ResourceType: r.ResourceType.String(),
				Role:         dao.Role_name[r.Role],
			})
		}
		d := Deployment{
			DeploymentId:   dep.BaseTemplateModel.BaseModel.ID,
			DeploymentName: dep.BaseTemplateModel.Name,
			Address:        dep.Address,
			Chain:          dao.ChainType_name[dep.ChainType],
			NetworkType:    dao.NetworkType_name[dep.NetworkType],
			DA:             dao.DaType_name[dep.Da],
			ProverType:     dao.ProverType_name[dep.ProverType],
			CloudVendor:    dao.CloudVendor_name[dep.CloudVendor],
			Zone:           dep.Az,
			Status:         dep.Status,
			CreateTime:     dep.CreatedAt.Unix(),
			ResourceSet:    rs,
			Message:        dep.Message,
		}

		deployments = append(deployments, d)
	}

	task.Response.Deployments = deployments
	return task.Response, nil
}
