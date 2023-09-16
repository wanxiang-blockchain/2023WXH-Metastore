package api

import (
	"github.com/MetaDataLab/web3-console-backend/internal/db"
	"github.com/MetaDataLab/web3-console-backend/internal/errors"
	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

func init() {
	Register(DESCRIBE_AVAILABILITY_ZONES_LABEL, NewDescribeAvailabilityZonesTask, COOKIEAUTH)
}

type DescribeAvailabilityZonesRequest struct {
	BaseRequest
	CloudVendor string `mapstructure:"CloudVendor,required"`
}

type AvailabilityZone struct {
	Zone   string
	Region string
}

type DescribeAvailabilityZonesResponse struct {
	BaseResponse
	AvailabilityZones []AvailabilityZone
}

type DescribeAvailabilityZonesTask struct {
	Request  *DescribeAvailabilityZonesRequest
	Response *DescribeAvailabilityZonesResponse
}

func NewDescribeAvailabilityZonesRequest(data *map[string]interface{}) (*DescribeAvailabilityZonesRequest, error) {
	req := &DescribeAvailabilityZonesRequest{}
	err := mapstructure.Decode(*data, &req)
	if err != nil {
		return nil, err
	}
	req.BaseRequest.RequestUUID = (*data)["RequestUUID"].(string)

	return req, nil
}

func NewDescribeAvailabilityZonesResponse(sessionId string) *DescribeAvailabilityZonesResponse {
	return &DescribeAvailabilityZonesResponse{
		BaseResponse: BaseResponse{
			Action:      DESCRIBE_AVAILABILITY_ZONES_LABEL + "Response",
			RequestUUID: sessionId,
		},
	}
}

func NewDescribeAvailabilityZonesTask(data *map[string]interface{}) (Task, error) {
	req, err := NewDescribeAvailabilityZonesRequest(data)
	if err != nil {
		return nil, err
	}
	task := &DescribeAvailabilityZonesTask{
		Request:  req,
		Response: NewDescribeAvailabilityZonesResponse(req.BaseRequest.RequestUUID),
	}

	validate := validator.New()
	err = validate.Struct(task.Request)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (task *DescribeAvailabilityZonesTask) Run(c *gin.Context) (Response, error) {

	cloudVendorValue, ok := dao.CloudVendor_value[task.Request.CloudVendor]
	if !ok {
		task.Response.BaseResponse.RetCode = int(errors.ParamsError("CloudVendor").Code())
		task.Response.BaseResponse.Message = errors.ParamsError("CloudVendor").String()
		return task.Response, nil
	}

	zones, err := db.GetZones(cloudVendorValue)
	if err != nil {
		task.Response.BaseResponse.RetCode = int(errors.GetZonesFailed().Code())
		task.Response.BaseResponse.Message = errors.GetZonesFailed().String()
		return task.Response, nil
	}

	for _, zone := range zones {
		az := AvailabilityZone{Zone: zone.Zone, Region: zone.Region}
		task.Response.AvailabilityZones = append(task.Response.AvailabilityZones, az)
	}

	return task.Response, nil
}
