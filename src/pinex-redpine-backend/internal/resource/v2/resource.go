package v2

import (
	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type ResourceType = dao.ResourceType

type Resource struct {
	Id   string
	Type ResourceType
	Role Role
}

type DeploymentType = dao.ChainType
type CloudVendor = dao.CloudVendor
type ProverType = dao.ProverType
type ChainType = dao.ChainType
type DaType = dao.DaType
type NetworkType = dao.NetworkType
type Role = dao.Role

type AccessToken struct {
	AccessKeyID     string `validate:"required"`
	AccessKeySecret string `validate:"required"`
}

type commonParams struct {
	nameBase   string
	region     string
	az         string
	projectId  string
	vpcId      string
	subnetId   string
	token      AccessToken
	cloud      CloudVendor
	proverType ProverType
}

func (p *commonParams) newUcloudConfigs() (ucloud.Config, auth.Credential) {
	cfg := ucloud.NewConfig()
	cfg.UserAgent = config.USER_AGENT
	if config.GConf.ApiBaseUrl != "" {
		log.Debugf("using api base url: %s", config.GConf.ApiBaseUrl)
		cfg.BaseUrl = config.GConf.ApiBaseUrl
	}

	credential := auth.NewCredential()
	credential.PublicKey = p.token.AccessKeyID
	credential.PrivateKey = p.token.AccessKeySecret
	return cfg, credential
}
