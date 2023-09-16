package resource

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func firstIfEx[T any](array []T) T {
	if len(array) > 0 {
		return array[0]
	}
	return *new(T)
}

func getRandomIps(params *commonParams, number int) (string, string, []string, error) {
	cfg := ucloud.NewConfig()
	cfg.UserAgent = config.USER_AGENT
	if config.GConf.ApiBaseUrl != "" {
		cfg.BaseUrl = config.GConf.ApiBaseUrl
	}

	credential := auth.NewCredential()
	credential.PublicKey = params.token.AccessKeyId
	credential.PrivateKey = params.token.SecretAccessKey

	client := vpc.NewClient(&cfg, &credential)
	describeSubnetReq := client.NewDescribeSubnetRequest()
	describeSubnetReq.Zone = ucloud.String(params.zone)
	describeSubnetReq.Region = ucloud.String(params.region)
	describeSubnetReq.ShowAvailableIPs = ucloud.Bool(true)
	if params.vpcId != "" {
		describeSubnetReq.VPCId = ucloud.String(params.vpcId)
	}
	if params.subnetId != "" {
		describeSubnetReq.SubnetId = ucloud.String(params.subnetId)
	}
	if params.projectId != "" {
		describeSubnetReq.ProjectId = ucloud.String(params.projectId)
	}
	resp, err := client.DescribeSubnet(describeSubnetReq)
	if err != nil {
		return "", "", nil, err
	}
	if resp.GetRetCode() != 0 {
		return "", "", nil, fmt.Errorf("failed describe subnet: %s", resp.GetMessage())
	}
	if len(resp.DataSet) == 0 {
		return "", "", nil, errors.New("allocate ip failed, cannot get subnet info")
	}

	targetSubnet := resp.DataSet[0]

	if params.vpcId != "" && params.vpcId != targetSubnet.VPCId {
		return "", "", nil, errors.New("allocate ip failed, returned vpc is not the specified one")
	}
	if params.subnetId != "" && params.subnetId != targetSubnet.SubnetId {
		return "", "", nil, errors.New("allocate ip failed, returned subnet is not the specified one")
	}

	subnetId := targetSubnet.SubnetId
	vpcId := targetSubnet.VPCId

	describeSubnetResourceReq := client.NewDescribeSubnetResourceRequest()
	describeSubnetResourceReq.Zone = ucloud.String(params.zone)
	describeSubnetResourceReq.Region = ucloud.String(params.region)

	describeSubnetResourceReq.SubnetId = ucloud.String(targetSubnet.SubnetId)

	if params.projectId != "" {
		describeSubnetResourceReq.ProjectId = ucloud.String(params.projectId)
	}
	respResource, err := client.DescribeSubnetResource(describeSubnetResourceReq)
	if err != nil {
		return "", "", nil, err
	}
	if respResource.GetRetCode() != 0 {
		return "", "", nil, fmt.Errorf("failed describe subnet resource: %s", respResource.GetMessage())
	}
	used := make(map[string]struct{})
	for _, resource := range respResource.DataSet {
		used[resource.IP] = struct{}{}
	}
	cidr := targetSubnet.Subnet + "/" + targetSubnet.Netmask

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ipString := ip.String()
		// skip occupied ips
		if strings.HasSuffix(ipString, ".0") || strings.HasSuffix(ipString, ".1") || strings.HasSuffix(ipString, ".255") || strings.HasSuffix(ipString, ".2") {
			continue
		}
		_, ok := used[ipString]
		if !ok {
			ips = append(ips, ipString)
		}
		// the first one is the network address
		if len(ips) == number {
			break
		}
	}
	if len(ips) != number {
		return "", "", nil, errors.New("cannot select enough ips in subnet")
	}
	return vpcId, subnetId, ips, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ#!%_&*()+=1234567890"

func randPassword(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	builder := strings.Builder{}
	for i := 0; i < n; i++ {
		c := charset[r.Intn(len(charset))]
		builder.WriteByte(c)
	}
	return builder.String()
}
