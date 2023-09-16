package api

import (
	"fmt"

	"github.com/MetaDataLab/web3-console-backend/internal/db"
	web3errors "github.com/MetaDataLab/web3-console-backend/internal/errors"
	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	v2 "github.com/MetaDataLab/web3-console-backend/internal/resource/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

func init() {
	Register(DELETE_DEPLOYMENT_LABEL, NewDeleteDeploymentRequest, COOKIEAUTH)
}

var _ Task = (*DescribeTemplateTask)(nil)

type DeleteDeploymentArgs struct {
	DeploymentId uint  `validate:"required"`
	AccessKey    Token `mapstructure:",squash"`
}

type DeleteDeploymentRequest = AnyRequest[DeleteDeploymentArgs]

type DeleteDeploymentTask DeleteDeploymentRequest

type DeleteDeploymentResp = BaseResponse

func NewDeleteDeploymentRequest(data *map[string]interface{}) (Task, error) {
	req := DeleteDeploymentTask{}
	err := mapstructure.Decode(*data, &req)
	if err != nil {
		return nil, err
	}
	req.BaseRequest.RequestUUID = (*data)[REQUEST_UUID].(string)
	req.BaseRequest.Address = (*data)[ADDRESS].(string)
	validate := validator.New()
	err = validate.Struct(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (t *DeleteDeploymentTask) Run(c *gin.Context) (Response, error) {
	if isInvalidToken(&t.Body.AccessKey) {
		return nil, web3errors.ParamsError("invalid token")
	}
	id := t.Body.DeploymentId
	deps, err := db.ListDeployment([]uint{id}, t.Address)
	if err != nil {
		return nil, err
	}
	if len(deps) == 0 {
		return nil, fmt.Errorf("deployment %d not found", id)
	}
	token := t.Body.AccessKey
	dep := deps[0]
	if dep.Status != UNKNOWN && dep.Status != FAILED && dep.Status != RUNNING {
		return nil, fmt.Errorf("deployment deletion not allowed with status: %s", dep.Status)
	}
	dep.Status = DELETING
	err = db.SaveDeployment(&dep)
	if err != nil {
		return nil, err
	}
	addServerWaitGroup(c)
	go func() {
		doneServerWaitGroup(c)
		for _, r := range dep.Resources {
			switch r.ResourceType {
			case dao.VM:
				err = v2.DeleteUhostVm(r.ResourceId, dep.Region, dep.Az, "", token, t.RequestUUID)
			case dao.DB:
				err = v2.DeleteUhostDb(r.ResourceId, dep.Region, dep.Az, "", token, t.RequestUUID)
			}
			if err != nil {
				log.Errorw("delete resource failed, err", "resource id", r.ResourceId, "resource role", "deployment id", id, "err", err, "request id", t.RequestUUID)
			}
		}
		if err != nil {
			dep.Status = FAILED
			dep.Message = "delete resource failed"
			err = db.SaveDeployment(&dep)
			if err != nil {
				log.Errorw("save deployment failed", "deployment id", id, "status", FAILED, "err", err, "request id", t.RequestUUID)
			}
			return
		}
		err = db.DeleteDeployment(id)
		if err != nil {
			log.Errorw("delete deployment failed", "deployment id", id, "err", err, "request id", t.RequestUUID)
		}
	}()
	return &DeleteDeploymentResp{
		Action:      DELETE_DEPLOYMENT_LABEL + "Response",
		RequestUUID: t.RequestUUID,
	}, nil
}
