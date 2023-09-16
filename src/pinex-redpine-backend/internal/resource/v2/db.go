package v2

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/ucloud/ucloud-sdk-go/services/upgsql"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

type PgsqlUcloud struct {
	*dao.DbTemplate
	*commonParams
	password  string
	client    *upgsql.UPgSQLClient
	requestId string
}

func NewDb(params *commonParams, tmpl *dao.DbTemplate, password string, requestId string) *PgsqlUcloud {
	cfg, credential := params.newUcloudConfigs()
	client := upgsql.NewClient(&cfg, &credential)
	return &PgsqlUcloud{
		DbTemplate:   tmpl,
		commonParams: params,
		password:     password,
		client:       client,
		requestId:    requestId,
	}
}

func (d *PgsqlUcloud) WaitReady(id string) error {
	req := d.makeGetDbReq(id)
	for {
		resp, err := d.client.GetUPgSQLInstance(req)
		if err != nil {
			return err
		}
		if resp.GetRetCode() != 0 {
			return fmt.Errorf("failed list pgsql instance: %s", resp.GetMessage())
		}
		meta := resp.DataSet
		if meta.State == "Fail" {
			return errors.New("created db failed")
		}
		if meta.State == "Shutdown" || meta.State == "Shutoff" || meta.State == "Delete" {
			return errors.New("bad db status, maybe modified manually")
		}
		if meta.State == "Running" {
			log.Debugw("db initialization done", "request id", d.requestId)
			return nil
		}
		log.Debugw("found target pgsql from response list", "current instance status", meta.State, "request id", d.requestId)
		time.Sleep(20 * time.Second)
	}
}

func (d *PgsqlUcloud) Role() dao.Role {
	return d.DbTemplate.Role
}

func (d *PgsqlUcloud) Create() (string, string, error) {
	parameterGroupId, err := d.findDefaultParameterGroupId()
	if err != nil {
		return "", "", err
	}
	req := d.makeCreatePgsqlRequest(parameterGroupId)
	resp, err := d.client.CreateUPgSQLInstance(req)
	if err != nil {
		return "", "", err
	}
	if resp.GetRetCode() != 0 {
		return "", "", fmt.Errorf("failed creating pgsql: %s", resp.GetMessage())
	}
	log.Debugw("pgsql creation done, pending pgsql initialization", "request id", d.requestId)

	dbId := resp.InstanceID
	getDbReq := d.makeGetDbReq(dbId)
	getResp, err := d.client.GetUPgSQLInstance(getDbReq)
	if err != nil {
		return "", "", err
	}
	return dbId, getResp.DataSet.IP, nil
}

func (d *PgsqlUcloud) makeListParameterGroupReq() request.GenericRequest {
	listParamsReq := d.client.NewGenericRequest()
	listParamsReq.SetAction("ListUPgSQLParamTemplate")
	listParamsReq.SetZone(d.az)
	listParamsReq.SetRegion(d.region)
	listParamsReq.SetProjectId(d.projectId)
	return listParamsReq
}

func (d *PgsqlUcloud) findDefaultParameterGroupId() (int, error) {
	listParamsReq := d.makeListParameterGroupReq()
	log.Debugw("listing parameter groups", "request id", d.requestId)
	listResp, err := d.client.GenericInvoke(listParamsReq)
	if err != nil {
		return 0, err
	}
	if listResp.GetRetCode() != 0 {
		return 0, fmt.Errorf("failed list param template: %s", listResp.GetMessage())
	}
	iDataSet := listResp.GetPayload()["Data"]
	dataSet, ok := iDataSet.([]interface{})
	if !ok {
		return 0, errors.New("unmarshal param template response dataset failed")
	}

	log.Debugw("get list response, finding default parameter group id", "db version", d.DbType, "request id", d.requestId)
	for _, data := range dataSet {
		if attrs, ok := data.(map[string]interface{}); ok {
			modifiable, ok := attrs["Modifiable"].(bool)
			if !ok {
				return 0, errors.New("unmarshal param template response dataset failed")
			}
			if modifiable {
				continue
			}
			dbVersion, ok := attrs["DBVersion"].(string)
			if !ok {
				return 0, errors.New("unmarshal param template response dataset failed")
			}
			if dbVersion != d.DbType {
				continue
			}
			var groupId int
			groupIdF, ok := attrs["GroupID"].(float64)
			if ok {
				groupId = int(groupIdF)
			} else {
				groupId, ok = attrs["GroupID"].(int)
				if !ok {
					return 0, errors.New("unmarshal param template response dataset failed")
				}
			}
			log.Debugw("parameter group found", "group id", groupId, "request id", d.requestId)
			return groupId, nil
		}
	}
	return 0, fmt.Errorf("default parameter group not found for db version: %s", d.DbType)
}

func (d *PgsqlUcloud) makeCreatePgsqlRequest(parameterGroupId int) *upgsql.CreateUPgSQLInstanceRequest {
	req := d.client.NewCreateUPgSQLInstanceRequest()
	var name string
	if d.nameBase == "" {
		name = strings.ToLower(d.DbTemplate.Role.String())
	} else {
		name = d.nameBase + "_" + strings.ToLower(d.DbTemplate.Role.String())
	}
	req.Name = ucloud.String(name)
	req.Zone = ucloud.String(d.az)
	req.Region = ucloud.String(d.region)
	if d.vpcId != "" {
		req.VPCID = ucloud.String(d.vpcId)
	}
	if d.subnetId != "" {
		req.SubnetID = ucloud.String(d.subnetId)
	}
	if d.projectId != "" {
		req.ProjectId = ucloud.String(d.projectId)
	}
	req.ParamGroupID = ucloud.Int(parameterGroupId)
	req.DBVersion = ucloud.String(d.DbType)
	req.AdminPassword = ucloud.String(d.password)

	req.MachineType = ucloud.String(d.MachineType)
	req.Port = ucloud.Int(d.Port)
	req.DiskSpace = ucloud.String(fmt.Sprint(d.DiskTemplate.Size))
	req.InstanceMode = ucloud.String(d.Mode)

	log.Debugw("creating pgsql instance",
		"disk space", req.DiskSpace,
		"port", req.Port,
		"instance mode", req.InstanceMode,
		"machine type", req.MachineType,
		"password", req.AdminPassword,
		"parameter group", req.ParamGroupID,
		"vpc", req.VPCID,
		"mode", req.InstanceMode,
		"subnet", req.SubnetID,
		"request id", d.requestId,
	)
	return req
}

func (d *PgsqlUcloud) makeGetDbReq(id string) *upgsql.GetUPgSQLInstanceRequest {
	req := d.client.NewGetUPgSQLInstanceRequest()
	req.InstanceID = ucloud.String(id)
	req.ProjectId = ucloud.String(d.projectId)
	req.Region = ucloud.String(d.region)
	req.Zone = ucloud.String(d.az)
	return req
}

func (d *PgsqlUcloud) makeListDbReq() *upgsql.ListUPgSQLInstanceRequest {
	req := d.client.NewListUPgSQLInstanceRequest()
	req.ProjectId = ucloud.String(d.projectId)
	req.Region = ucloud.String(d.region)
	req.Zone = ucloud.String(d.az)
	return req
}

func (d *PgsqlUcloud) Delete(id string) error {
	if err := d.stopDb(id); err != nil {
		if err == resourceNotFoundError {
			return nil
		}
		return err
	}
	log.Debugw("terminating pgsql", "db id", id, "request id", d.requestId)
	termReq := d.makeTerminateDbReq(id)
	termResp, err := d.client.DeleteUPgSQLInstance(termReq)
	if err != nil {
		return err
	}
	if termResp.RetCode != 0 {
		return fmt.Errorf("terminate db failed: %s", termResp.GetMessage())
	}
	return err
}

func (d *PgsqlUcloud) stopDb(id string) error {
	closed, err := d.findDb(id)
	if err != nil {
		return err
	}
	if closed {
		return nil
	}
	log.Debugw("stopping db", "db id", id, "request id", d.requestId)
	stopReq := d.makeCloseDbReq(id)
	stopResp, err := d.client.StopUPgSQLInstance(stopReq)
	if err != nil {
		return err
	}
	if stopResp.RetCode != 0 {
		return fmt.Errorf("stop db failed: %s", stopResp.GetMessage())
	}
	log.Debugw("waiting db stop", "db id", id, "request id", d.requestId)
	for {
		stopped, err := d.ifDbStopped(id)
		if err != nil {
			return err
		}
		if stopped {
			log.Debugw("db stopped", "request id", d.requestId)
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (d *PgsqlUcloud) findDb(id string) (bool, error) {
	req := d.makeListDbReq()
	resp, err := d.client.ListUPgSQLInstance(req)
	if err != nil {
		return false, err
	}
	for _, meta := range resp.DataSet {
		if meta.InstanceID == id {
			if meta.State == "Stopped" {
				return true, nil
			} else {
				return false, nil
			}
		}
	}
	return true, resourceNotFoundError
}

func (d *PgsqlUcloud) ifDbStopped(id string) (bool, error) {
	req := d.makeGetDbReq(id)
	resp, err := d.client.GetUPgSQLInstance(req)
	if err != nil {
		return false, err
	}
	if resp.DataSet.State == "Stopped" {
		log.Debugw("db stopped", "db id", id, "request id", d.requestId)
		return true, nil
	}
	return false, nil
}

func (d *PgsqlUcloud) makeCloseDbReq(id string) *upgsql.StopUPgSQLInstanceRequest {
	req := d.client.NewStopUPgSQLInstanceRequest()
	req.Region = ucloud.String(d.region)
	req.Zone = ucloud.String(d.az)
	req.InstanceID = ucloud.String(id)
	req.ProjectId = ucloud.String(d.projectId)
	return req
}

func (d *PgsqlUcloud) makeTerminateDbReq(id string) *upgsql.DeleteUPgSQLInstanceRequest {
	req := d.client.NewDeleteUPgSQLInstanceRequest()
	req.Region = ucloud.String(d.region)
	req.Zone = ucloud.String(d.az)
	req.InstanceID = ucloud.String(id)
	req.ProjectId = ucloud.String(d.projectId)
	return req
}

func (d *PgsqlUcloud) InstancePrice() (float64, error) {
	req := d.client.NewGetUPgSQLInstancePriceRequest()
	req.ChargeType = ucloud.String("Month")
	req.MachineType = ucloud.String(d.MachineType)
	req.Zone = ucloud.String(d.az)
	req.Region = ucloud.String(d.region)
	req.DiskSpace = ucloud.Int(int(d.DiskTemplate.Size))
	req.InstanceMode = ucloud.String(d.Mode)
	req.ProjectId = ucloud.String(d.projectId)
	req.Quantity = ucloud.Int(1)

	resp, err := d.client.GetUPgSQLInstancePrice(req)
	if err != nil {
		return 0, err
	}
	for _, ps := range resp.PriceSet {
		if ps.ChargeType == "Month" {
			return ps.Price, nil
		}
	}
	return 0, errors.New("get db price failed from ucloud api failed, price not found")
}

func (d *PgsqlUcloud) NetworkPrice() (float64, error) {
	return 0, nil
}

func DeleteUhostDb(dbId string, region string, az string, projectId string, token AccessToken, requestId string) error {
	params := commonParams{
		region:    region,
		az:        az,
		token:     token,
		projectId: projectId,
	}
	v := NewDb(&params, nil, "", requestId)
	return v.Delete(dbId)
}
