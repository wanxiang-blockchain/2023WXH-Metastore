package resource

import (
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

func doDeleteHost(id string, region string, az string, projectId string, token Token) error {
	cfg := ucloud.NewConfig()
	cfg.UserAgent = config.USER_AGENT
	if config.GConf.ApiBaseUrl != "" {
		cfg.BaseUrl = config.GConf.ApiBaseUrl
		log.Debugf("using api base url: %s", config.GConf.ApiBaseUrl)
	}

	credential := auth.NewCredential()
	credential.PublicKey = token.AccessKeyId
	credential.PrivateKey = token.SecretAccessKey

	client := uhost.NewClient(&cfg, &credential)
	if err := stopUhost(client, id, region, az, projectId); err != nil {
		return err
	}
	log.Debugf("terminating uhost %s", id)
	termReq := client.NewTerminateUHostInstanceRequest()
	termReq.Region = ucloud.String(region)
	termReq.Zone = ucloud.String(az)
	termReq.UHostId = ucloud.String(id)
	termReq.ProjectId = ucloud.String(projectId)

	_, err := client.TerminateUHostInstance(termReq)
	return err
}

func stopUhost(client *uhost.UHostClient, id string, region string, az string, projectId string) error {
	describeReq := client.NewDescribeUHostInstanceRequest()
	describeReq.Region = ucloud.String(region)
	describeReq.Zone = ucloud.String(az)
	describeReq.UHostIds = []string{id}
	describeReq.ProjectId = ucloud.String(projectId)
	closed, err := isUhostStopped(id, client, describeReq)
	if err != nil {
		return err
	}
	if closed {
		return nil
	}
	stopReq := client.NewStopUHostInstanceRequest()
	stopReq.Region = ucloud.String(region)
	stopReq.Zone = ucloud.String(az)
	stopReq.UHostId = ucloud.String(id)
	stopReq.ProjectId = ucloud.String(projectId)
	log.Debugf("stopping uhost %s", id)
	resp, err := client.StopUHostInstance(stopReq)
	if err != nil {
		return err
	}
	if resp.GetRetCode() != 0 {
		return fmt.Errorf("stop uhost failed: %s", resp.GetMessage())
	}
	log.Debugf("waiting uhost stop, id: %s", id)
	for {

		closed, err := isUhostStopped(id, client, describeReq)
		if err != nil {
			return err
		}
		if closed {
			log.Debugf("uhost stopped, id: %s", id)
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func isUhostStopped(id string, client *uhost.UHostClient, req *uhost.DescribeUHostInstanceRequest) (bool, error) {
	describeResp, err := client.DescribeUHostInstance(req)
	if err != nil {
		return false, err
	}
	if describeResp.GetRetCode() != 0 {
		return false, fmt.Errorf("failed list uhost instance: %s", describeResp.GetMessage())
	}
	if len(describeResp.UHostSet) == 1 {
		instanceMeta := describeResp.UHostSet[0]
		log.Debugf("uhost %s status: %s", id, instanceMeta.State)
		if instanceMeta.State == "Stopped" {
			return true, nil
		}
	}
	return false, nil
}

func doDeleteDb(id string, region string, az string, projectId string, token Token) error {
	cfg := ucloud.NewConfig()
	cfg.UserAgent = config.USER_AGENT
	if config.GConf.ApiBaseUrl != "" {
		cfg.BaseUrl = config.GConf.ApiBaseUrl
		log.Debugf("using api base url: %s", config.GConf.ApiBaseUrl)
	}

	credential := auth.NewCredential()
	credential.PublicKey = token.AccessKeyId
	credential.PrivateKey = token.SecretAccessKey

	client := upgsql.NewClient(&cfg, &credential)
	if err := stopDb(client, id, region, az, projectId); err != nil {
		return err
	}
	log.Debugf("terminating pgsql %s", id)
	termReq := client.NewDeleteUPgSQLInstanceRequest()
	termReq.Region = ucloud.String(region)
	termReq.Zone = ucloud.String(az)
	termReq.InstanceID = ucloud.String(id)
	termReq.ProjectId = ucloud.String(projectId)
	termResp, err := client.DeleteUPgSQLInstance(termReq)
	if err != nil {
		return err
	}
	if termResp.RetCode != 0 {
		return fmt.Errorf("terminate db failed: %s", termResp.GetMessage())
	}
	return err
}

func stopDb(client *upgsql.UPgSQLClient, id string, region string, az string, projectId string) error {
	listReq := client.NewListUPgSQLInstanceRequest()
	listReq.Region = ucloud.String(region)
	listReq.Zone = ucloud.String(az)

	listReq.ProjectId = ucloud.String(projectId)
	closed, err := ifDbStopped(id, client, listReq)
	if err != nil {
		return err
	}
	if closed {
		return nil
	}
	log.Debugf("stopping db %s", id)
	closeReq := client.NewStopUPgSQLInstanceRequest()
	closeReq.Region = ucloud.String(region)
	closeReq.Zone = ucloud.String(az)
	closeReq.InstanceID = ucloud.String(id)
	closeReq.ProjectId = ucloud.String(projectId)
	stopResp, err := client.StopUPgSQLInstance(closeReq)
	if err != nil {
		return err
	}
	if stopResp.RetCode != 0 {
		return fmt.Errorf("stop db failed: %s", stopResp.GetMessage())
	}
	log.Debugf("waiting db stop, id: %s", id)
	for {
		closed, err := ifDbStopped(id, client, listReq)
		if err != nil {
			return err
		}
		if closed {
			log.Debugf("db stopped, id: %s", id)
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func ifDbStopped(id string, client *upgsql.UPgSQLClient, req *upgsql.ListUPgSQLInstanceRequest) (bool, error) {
	listResp, err := client.ListUPgSQLInstance(req)
	if err != nil {
		return false, err
	}
	if listResp.GetRetCode() != 0 {
		return false, fmt.Errorf("failed list pgsql instance: %s", listResp.GetMessage())
	}
	for _, meta := range listResp.DataSet {
		if meta.InstanceID == id {
			if meta.State == "Fail" {
				return false, errors.New("created db failed")
			}
			if meta.State == "Stopped" {
				return true, nil
			}
		}
	}
	return false, nil
}
