package v2

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

type VmUcloud struct {
	*dao.NodeTemplate
	*commonParams
	keyPairId    string
	ip           string
	userdataArgs map[string]any
	client       *uhost.UHostClient
	requestId    string
}

func NewVm(params *commonParams, tmpl *dao.NodeTemplate,
	keyPairId string, ip string, requestId string, userdataArgs map[string]any) *VmUcloud {
	cfg, credential := params.newUcloudConfigs()
	vmClient := uhost.NewClient(&cfg, &credential)
	return &VmUcloud{
		NodeTemplate: tmpl,
		commonParams: params,
		keyPairId:    keyPairId,
		ip:           ip,
		userdataArgs: userdataArgs,
		client:       vmClient,
		requestId:    requestId,
	}
}

func (v *VmUcloud) WaitReady(id string) error {
	log.Debugw("waiting for uhost instance ready", "request id", v.requestId)
	describeReq := v.makeDescribeReq(id)
	for {
		describeResp, err := v.client.DescribeUHostInstance(describeReq)
		if err != nil {
			return err
		}
		if describeResp.GetRetCode() != 0 {
			return fmt.Errorf("failed describe uhost instance: %s", describeResp.GetMessage())
		}
		if len(describeResp.UHostSet) == 1 {
			instanceMeta := describeResp.UHostSet[0]
			log.Debugw("found target uhost from response list", "current instance status", instanceMeta.State, "request id", v.requestId)
			if instanceMeta.State == "Running" {
				log.Debugw("uhost initialization done", "host id", instanceMeta.UHostId, "request id", v.requestId)
				return nil
			}
			if instanceMeta.State == "Install Fail" {
				return errors.New("create instance failed")
			}
			if instanceMeta.State == "Stopping" || instanceMeta.State == "Stopped" || instanceMeta.State == "Install Fail" || instanceMeta.State == "" {
				return errors.New("bad instance status, maybe modified manually")
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func (v *VmUcloud) Role() dao.Role {
	return v.NodeTemplate.Role
}

func (v *VmUcloud) Create() (string, string, error) {
	req, err := v.makeCreateReq()
	if err != nil {
		return "", "", err
	}
	resp, err := v.client.CreateUHostInstance(req)
	if err != nil {
		return "", "", err
	}
	if resp.GetRetCode() != 0 {
		return "", "", fmt.Errorf("failed creating uhost: %s", resp.GetMessage())
	}
	if len(resp.UHostIds) != 1 {
		return "", "", errors.New("get vm id failed")
	}
	vmId := resp.UHostIds[0]
	ip := resp.IPs[0]
	log.Debugw("vm creation done", "host id", vmId, "request id", v.requestId)
	if v.EipTemplate != nil {
		v.createEip(vmId)
	}
	return vmId, ip, nil
}

func (v *VmUcloud) makeCreateReq() (*uhost.CreateUHostInstanceRequest, error) {
	req := v.client.NewCreateUHostInstanceRequest()
	var name string
	if v.nameBase == "" {
		name = strings.ToLower(v.NodeTemplate.Role.String())
	} else {
		name = v.nameBase + "_" + strings.ToLower(v.NodeTemplate.Role.String())
	}
	req.Name = ucloud.String(name)
	req.Zone = ucloud.String(v.az)
	req.Region = ucloud.String(v.region)
	req.ChargeType = ucloud.String("Month")
	if v.vpcId != "" {
		req.VPCId = ucloud.String(v.vpcId)
	}
	if v.subnetId != "" {
		req.SubnetId = ucloud.String(v.subnetId)
	}
	if v.projectId != "" {
		req.ProjectId = ucloud.String(v.projectId)
	}
	req.PrivateIp = append(req.PrivateIp, v.ip)
	req.Disks = make([]uhost.UHostDisk, 2)
	req.MachineType = ucloud.String(v.MachineTemplate.MachineType)
	req.Disks[0] = uhost.UHostDisk{
		Size:   ucloud.Int(int(v.ImageTemplate.DiskTemplate.Size)),
		IsBoot: ucloud.String("True"),
		Type:   ucloud.String(v.ImageTemplate.DiskTemplate.Type),
	}
	req.Disks[1] = uhost.UHostDisk{
		Size:   ucloud.Int(int(v.NodeTemplate.DiskTemplate.Size)),
		IsBoot: ucloud.String("False"),
		Type:   ucloud.String(v.NodeTemplate.DiskTemplate.Type),
	}
	req.ImageId = ucloud.String(v.ImageTemplate.ImageId)
	req.LoginMode = ucloud.String("KeyPair")
	req.KeyPairId = &v.keyPairId
	req.MinimalCpuPlatform = ucloud.String(v.MachineTemplate.Platform)
	req.CPU = ucloud.Int(v.MachineTemplate.CoreNum)
	req.Memory = ucloud.Int(v.MachineTemplate.Memory)
	if len(v.ImageTemplate.Userdatas) != 0 {
		ud, err := v.renderUserdata()
		if err != nil {
			return nil, err
		}
		req.UserData = ucloud.String(ud)
	}
	log.Debugw("creating uhost instance",
		"image id", req.ImageId,
		"ip", req.PrivateIp,
		"KeyPairId", req.KeyPairId,
		"machine type", req.MachineType,
		"cpu", req.CPU,
		"memory", req.Memory,
		"vpc", req.VPCId,
		"subnet", req.SubnetId,
		"disks", req.Disks,
		"request id", v.requestId,
	)
	return req, nil
}

func (v *VmUcloud) renderUserdata() (string, error) {
	if len(v.ImageTemplate.Userdatas) == 0 {
		return "", nil
	}
	ud := v.ImageTemplate.Userdatas[0]
	if len(ud.Args) == 0 {
		return base64.StdEncoding.EncodeToString([]byte(ud.Content)), nil
	}
	var args = make([]interface{}, len(ud.Args))
	for _, arg := range ud.Args {
		argVal := v.userdataArgs[arg.Name]
		if arg.Position >= uint8(len(ud.Args)) {
			return "", fmt.Errorf("bad userdata arg position: %+v", arg)
		}
		args[arg.Position] = fmt.Sprint(argVal)
	}
	udStr := fmt.Sprintf(ud.Content, args...)
	return base64.StdEncoding.EncodeToString([]byte(udStr)), nil
}

func (v *VmUcloud) Delete(id string) error {
	if err := v.stop(id); err != nil {
		if err == resourceNotFoundError {
			return nil
		}
		return err
	}
	log.Debugw("terminating uhost", "host id", id, "request id", v.requestId)
	termReq := v.makeTerminateReq(id)

	_, err := v.client.TerminateUHostInstance(termReq)
	return err
}

func (v *VmUcloud) stop(id string) error {
	closed, err := v.isUhostStopped(id)
	if err != nil {
		return err
	}
	if closed {
		return nil
	}
	stopReq := v.makeStopReq(id)
	log.Debugw("stopping uhost", "host id", id, "request id", v.requestId)
	resp, err := v.client.StopUHostInstance(stopReq)
	if err != nil {
		return err
	}
	if resp.GetRetCode() != 0 {
		return fmt.Errorf("stop uhost failed: %s", resp.GetMessage())
	}
	log.Debugw("waiting for uhost stop", "host id", id, "request id", v.requestId)
	for {
		closed, err := v.isUhostStopped(id)
		if err != nil {
			return err
		}
		if closed {
			log.Debugw("uhost stopped", "host id", id, "request id", v.requestId)
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (v *VmUcloud) makeDescribeReq(id string) *uhost.DescribeUHostInstanceRequest {
	describeReq := v.client.NewDescribeUHostInstanceRequest()
	describeReq.Region = ucloud.String(v.region)
	describeReq.Zone = ucloud.String(v.az)
	describeReq.UHostIds = []string{id}
	describeReq.ProjectId = ucloud.String(v.projectId)
	return describeReq
}
func (v *VmUcloud) makeStopReq(id string) *uhost.StopUHostInstanceRequest {
	stopReq := v.client.NewStopUHostInstanceRequest()
	stopReq.Region = ucloud.String(v.region)
	stopReq.Zone = ucloud.String(v.az)
	stopReq.UHostId = ucloud.String(id)
	stopReq.ProjectId = ucloud.String(v.projectId)
	return stopReq
}
func (v *VmUcloud) makeTerminateReq(id string) *uhost.TerminateUHostInstanceRequest {
	termReq := v.client.NewTerminateUHostInstanceRequest()
	termReq.Region = ucloud.String(v.region)
	termReq.Zone = ucloud.String(v.az)
	termReq.UHostId = ucloud.String(id)
	termReq.ProjectId = ucloud.String(v.projectId)
	return termReq
}

func (v *VmUcloud) isUhostStopped(id string) (bool, error) {
	req := v.makeDescribeReq(id)
	describeResp, err := v.client.DescribeUHostInstance(req)
	if err != nil {
		return false, err
	}
	if describeResp.GetRetCode() != 0 {
		return false, fmt.Errorf("failed list uhost instance: %s", describeResp.GetMessage())
	}
	if len(describeResp.UHostSet) == 1 {
		instanceMeta := describeResp.UHostSet[0]
		if instanceMeta.State == "Stopped" {
			return true, nil
		}
	} else if len(describeResp.UHostSet) == 0 {
		return true, resourceNotFoundError
	}
	return false, nil
}

func (v *VmUcloud) createEip(id string) error {
	cfg, credential := v.commonParams.newUcloudConfigs()
	client := unet.NewClient(&cfg, &credential)
	allocReq := client.NewAllocateEIPRequest()
	allocReq.Bandwidth = ucloud.Int(v.EipTemplate.BindWidth)
	var name string
	if v.nameBase == "" {
		name = strings.ToLower(v.NodeTemplate.Role.String()) + "_eip"
	} else {
		name = v.nameBase + "_" + strings.ToLower(v.NodeTemplate.Role.String()) + "_eip"
	}
	allocReq.Name = ucloud.String(name)
	allocReq.OperatorName = ucloud.String("Bgp")
	allocReq.Region = ucloud.String(v.region)
	allocReq.Zone = ucloud.String(v.az)
	allocReq.ProjectId = ucloud.String(v.projectId)
	allocResp, err := client.AllocateEIP(allocReq)
	if err != nil {
		return err
	}
	bindReq := client.NewBindEIPRequest()
	bindReq.Region = ucloud.String(v.region)
	bindReq.Zone = ucloud.String(v.az)
	bindReq.ProjectId = ucloud.String(v.projectId)
	if len(allocResp.EIPSet) == 0 {
		return errors.New("alloc eip failed")
	}
	bindReq.EIPId = ucloud.String(allocResp.EIPSet[0].EIPId)
	bindReq.ResourceType = ucloud.String("uhost")
	bindReq.ResourceId = ucloud.String(id)
	_, err = client.BindEIP(bindReq)
	if err != nil {
		return err
	}
	return nil
}

func (v *VmUcloud) InstancePrice() (float64, error) {
	req := v.client.NewGetUHostInstancePriceRequest()
	req.CPU = ucloud.Int(v.MachineTemplate.CoreNum)
	req.Memory = ucloud.Int(v.MachineTemplate.Memory)
	req.ChargeType = ucloud.String("Month")
	req.Disks = make([]uhost.UHostDisk, 2)
	req.MachineType = ucloud.String(v.MachineTemplate.MachineType)
	req.Disks[0] = uhost.UHostDisk{
		Size:   ucloud.Int(int(v.ImageTemplate.DiskTemplate.Size)),
		IsBoot: ucloud.String("True"),
		Type:   ucloud.String(v.ImageTemplate.DiskTemplate.Type),
	}
	req.Disks[1] = uhost.UHostDisk{
		Size:   ucloud.Int(int(v.NodeTemplate.DiskTemplate.Size)),
		IsBoot: ucloud.String("False"),
		Type:   ucloud.String(v.NodeTemplate.DiskTemplate.Type),
	}
	req.MachineType = ucloud.String(v.MachineTemplate.MachineType)
	req.Count = ucloud.Int(1)
	req.Zone = ucloud.String(v.az)
	req.Region = ucloud.String(v.region)
	req.ProjectId = ucloud.String(v.projectId)
	if strings.Contains(v.MachineTemplate.Platform, "Amd") {
		req.CpuPlatform = ucloud.String("Amd")
	}
	resp, err := v.client.GetUHostInstancePrice(req)
	if err != nil {
		return 0, err
	}
	for _, ps := range resp.PriceSet {
		if ps.ChargeType == "Month" {
			return ps.Price, nil
		}
	}
	return 0, errors.New("get uhost price failed from ucloud api failed, price not found")
}

func (v *VmUcloud) NetworkPrice() (float64, error) {
	if v.EipTemplate == nil {
		return 0, nil
	}
	cfg, credential := v.commonParams.newUcloudConfigs()
	client := unet.NewClient(&cfg, &credential)
	req := client.NewGetEIPPriceRequest()
	req.ChargeType = ucloud.String("Month")
	req.Zone = ucloud.String(v.az)
	req.Region = ucloud.String(v.region)
	req.OperatorName = ucloud.String("Bgp")
	req.Bandwidth = ucloud.Int(v.EipTemplate.BindWidth)
	req.ProjectId = ucloud.String(v.projectId)
	resp, err := client.GetEIPPrice(req)
	if err != nil {
		return 0, err
	}
	for _, ps := range resp.PriceSet {
		if ps.ChargeType == "Month" {
			return ps.Price, nil
		}
	}
	return 0, errors.New("get eip price failed from ucloud api failed, price not found")
}

func DeleteUhostVm(hostId string, region string, az string, projectId string, token AccessToken, requestId string) error {
	params := commonParams{
		region:    region,
		az:        az,
		token:     token,
		projectId: projectId,
	}
	v := NewVm(&params, nil, "", "", requestId, nil)
	return v.Delete(hostId)
}
