package api

import (
	"context"

	"github.com/MetaDataLab/web3-console-backend/internal/db"
	web3errors "github.com/MetaDataLab/web3-console-backend/internal/errors"
	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	v2 "github.com/MetaDataLab/web3-console-backend/internal/resource/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/sync/errgroup"
)

func init() {
	Register(CREATE_DEPLOYMENT_LABEL, NewCreateDeploymentTask, COOKIEAUTH)
}

var _ Task = (*DescribeTemplateTask)(nil)

type CreateDeploymentParamGroup struct {
	DeploymentParamGroup `mapstructure:",squash"`
	KeyPairId            string `validate:"required"`
	InstallAgent         bool
	Name                 string `mapstructure:"DeploymentName" validate:"required"`
}

type CreateDeploymentTask CreateDeploymentRequest

type CreateDeploymentRequest = AnyRequest[CreateDeploymentParamGroup]

type CreateDeploymentResp = BaseResponse

func NewCreateDeploymentTask(data *map[string]interface{}) (Task, error) {
	req := CreateDeploymentTask{}
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

func (t *CreateDeploymentTask) Run(c *gin.Context) (Response, error) {
	if isInvalidToken(&t.Body.AccessKey) {
		return nil, web3errors.ParamsError("invalid token")
	}
	_, dep, params, err := doLoadTemplate(&t.Body.DeploymentParamGroup, t.Body.Name, t.Body.KeyPairId, t.RequestUUID)
	if err != nil {
		return nil, err
	}
	deployment := &dao.Deployment{
		Address:     t.Address,
		ChainType:   params.chain,
		NetworkType: params.networkType,
		Da:          params.da,
		ProverType:  params.proverType,
		CloudVendor: params.cloudVendor,
		Region:      t.Body.Region,
		Az:          t.Body.Zone,
		Status:      DEPLOYING,
	}
	deployment.Name = t.Body.Name
	err = db.SaveDeployment(deployment)
	if err != nil {
		return nil, err
	}
	addServerWaitGroup(c)
	// asynchronized create resources
	go func() {
		defer doneServerWaitGroup(c)
		var resources []v2.Resource
		statusChan := dep.StatusChan()
		ctx, cancel := context.WithCancel(context.Background())
		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			defer func() {
				cancel()
			}()
			rs, err := dep.Build()
			if err != nil {
				return err
			}
			resources = rs
			return nil
		})

		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case stat := <-statusChan:
					deployment.Message = stat
					err = db.SaveDeployment(deployment)
					if err != nil {
						log.Errorw("saving deployment failed", "deployment id", deployment.ID, "err", err, "request id", t.RequestUUID)
					}
				}
			}

		})
		err = eg.Wait()
		if err != nil {
			deployment.Status = FAILED
			deployment.Message = err.Error()
			err = db.SaveDeployment(deployment)
			if err != nil {
				log.Errorw("saving deployment failed", "deployment id", deployment.ID, "err", err, "request id", t.RequestUUID)
			}
			return
		}
		dbResource := make([]dao.Resource, 0)
		for _, r := range resources {
			dbResource = append(dbResource, dao.Resource{
				ResourceId:   r.Id,
				ResourceType: r.Type,
				Role:         r.Role,
			})
		}
		deployment.Resources = dbResource
		// here the deployment needs to wait for user's key
		// the next status should be set by agent
		deployment.Status = UNKNOWN
		deployment.Message = CREATION_DONE
		err = db.SaveDeployment(deployment)
		if err != nil {
			log.Errorw("saving deployment failed after creation", "deployment id", deployment.ID, "err", err, "request id", t.RequestUUID)
		}
	}()
	return &CreateDeploymentResp{
		Action:      CREATE_DEPLOYMENT_LABEL + "Response",
		RequestUUID: t.RequestUUID,
	}, nil
}
