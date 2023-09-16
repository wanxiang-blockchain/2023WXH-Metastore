package api

import (
	"sync"

	"github.com/MetaDataLab/web3-console-backend/internal/pkg"
	"github.com/gin-gonic/gin"
)

const codeTemplate = "Welcome to PINEX!\n\nThis request will not trigger a blockchain transaction or cost any gas fees. It is only used to authorise logging into PINEX.\n\nYour authentication status will reset after 12 hours.\n\nWallet address:\nADDRESS\n\nNonce:\nNONCE"

// action labels
const (
	LOGIN_PINEX_LABEL                  = "LoginPinex"
	GET_PINEX_CODE_LABEL               = "GetPinexCode"
	DESCRIBE_AVAILABILITY_ZONES_LABEL  = "DescribeAvailabilityZones"
	CREATE_DEPLOYMENT_LABEL            = "CreateDeployment"
	DELETE_DEPLOYMENT_LABEL            = "DeleteDeployment"
	LIST_DEPLOYMENTS_LABEL             = "ListDeployments"
	DESCRIBE_DEPLOYMENT_TEMPLATE_LABEL = "DescribeDeploymentTemplate"
)

// param labels
const (
	ADDRESS      = "Address"
	REQUEST_UUID = "RequestUUID"
)

type DeploymentStatus = string

const (
	DEPLOYING DeploymentStatus = "Deploying"
	FAILED                     = "Failed"
	UNKNOWN                    = "Unknown"
	DELETING                   = "Deleting"
	RUNNING                    = "Running"
)

const (
	CREATION_DONE = "Resources creation done"
)

func extractServerWaitGroup(c *gin.Context) *sync.WaitGroup {
	wgI := c.Value(pkg.SERVER_WAIT_GROUP_LABEL)
	if wgI != nil {
		wg, ok := wgI.(*sync.WaitGroup)
		if ok {
			return wg
		}
	}
	return nil
}

// mind that this function should be put in request goroutine, not in a child goroutines
// otherwise it may panic when shutting down the server
func addServerWaitGroup(c *gin.Context) {
	wg := extractServerWaitGroup(c)
	if wg != nil {
		wg.Add(1)
	}
}
func doneServerWaitGroup(c *gin.Context) {
	wg := extractServerWaitGroup(c)
	if wg != nil {
		wg.Done()
	}
}

func isInvalidToken(t *Token) bool {
	return t.AccessKeyID == "" || t.AccessKeySecret == ""
}
