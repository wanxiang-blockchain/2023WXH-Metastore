package resource

import (
	"fmt"

	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
)

type ResourceType uint8

const (
	_ ResourceType = iota
	POLYGON_ZKEVM_PROVER
	POLYGON_ZKEVM_NODE
	POLYGON_ZKEVM_DB
	OP_NODE
)

func (r ResourceType) String() string {
	switch r {
	case POLYGON_ZKEVM_PROVER:
		return "POLYGON_ZKEVM_PROVER"
	case POLYGON_ZKEVM_NODE:
		return "POLYGON_ZKEVM_NODE"
	case POLYGON_ZKEVM_DB:
		return "POLYGON_ZKEVM_DB"
	case OP_NODE:
		return "OP_NODE"
	}
	return fmt.Sprintf("unknown resource type: %d", r)
}

type Resource struct {
	Id   string
	Type ResourceType
}

type DeploymentType uint8

const (
	_ DeploymentType = iota
	POLYGON_ZKEVM
	OP_STACK
)

func (r DeploymentType) String() string {
	switch r {
	case POLYGON_ZKEVM:
		return "POLYGON_ZKEVM"
	case OP_STACK:
		return "OP_STACK"

	}
	return fmt.Sprintf("unknown deployment type: %d", r)
}

type Token struct {
	AccessKeyId     string
	SecretAccessKey string
}

type CreateDeploymentRequest struct {
	Token
	Name      string
	Type      DeploymentType
	Region    string
	ProjectId string
	Az        string
	KeyPairId string
}

func CreateDeployment(req CreateDeploymentRequest) ([]Resource, string, error) {
	creq := createDeploymentReq{
		CreateDeploymentRequest: req,
	}
	log.Debugf("creating deployment: %+v", req)
	return creq.create()
}

func DeleteHost(id string, region string, az string, projectId string, token Token) error {
	return doDeleteHost(id, region, az, projectId, token)
}

func DeleteDB(id string, region string, az string, projectId string, token Token) error {
	return doDeleteDb(id, region, az, projectId, token)
}
