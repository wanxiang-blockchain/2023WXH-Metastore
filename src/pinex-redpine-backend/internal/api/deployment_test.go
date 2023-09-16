package api

import (
	"testing"
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/db"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testAddressId = "pinex_123"

var testToken = Token{
	// AccessKeyID:     "ak123",
	// AccessKeySecret: "IWillNotTellYou",
	AccessKeyID:     "4ZDughY8BbJ2NOEOTiG0u96p3rFZR0Nly",
	AccessKeySecret: "IJ2ozFmpjc3KKShxD2c1ZzH9Ceife1m8FCV2RAbjBLMI",
}

var testSuites = []CreateDeploymentParamGroup{
	{
		DeploymentParamGroup: DeploymentParamGroup{
			Chain:       "polygonzkevm",
			NetworkType: "devnet",
			ProverType:  "exclusive",
			DA:          "eth",
			CloudVendor: "surfercloud",
			Region:      "cn-bj2",
			Zone:        "cn-bj2-05",
			AccessKey:   testToken,
		},
		KeyPairId: "zkhh1h",
		Name:      "zkevm",
	},
	{
		DeploymentParamGroup: DeploymentParamGroup{
			Chain:       "opstack",
			NetworkType: "devnet",
			DA:          "eth",
			CloudVendor: "surfercloud",
			Region:      "cn-bj2",
			Zone:        "cn-bj2-05",
			AccessKey:   testToken,
		},
		KeyPairId: "zkhh1h",
		Name:      "opstack",
	},
}

func TestMain(m *testing.M) {
	config.InitConfig("/root/go/src/web3-console-backend/.vscode/config.yaml")
	log.InitGlobalLogger(&config.GConf.LogConfig, zap.AddCallerSkip(1))
	db.Init()
	m.Run()
}

func TestDescribeTemplate(t *testing.T) {
	// success
	for _, g := range testSuites {
		args := makeTestArgs(g)
		task, err := NewTask(DESCRIBE_DEPLOYMENT_TEMPLATE_LABEL, &args)
		require.Empty(t, err)
		require.NotEmpty(t, task)

		concreteTask, ok := task.(*DescribeTemplateTask)
		require.True(t, ok)
		require.NotEmpty(t, concreteTask)
		require.Equal(t, concreteTask.Body, g.DeploymentParamGroup)
		resp, err := task.Run(nil)
		require.Empty(t, err)
		require.NotEmpty(t, resp)
	}
	// params wrong
	for _, g := range testSuites {
		gLocal := g
		gLocal.Chain = "wrong_chain"
		args := makeTestArgs(gLocal)
		task, err := NewTask(DESCRIBE_DEPLOYMENT_TEMPLATE_LABEL, &args)
		require.NotEmpty(t, task)
		require.Empty(t, err)
		resp, err := task.Run(nil)
		require.NotEmpty(t, err)
		require.Empty(t, resp)
	}

	// params unsupported
	for _, g := range testSuites {
		gLocal := g
		gLocal.DA = "meta_store"
		args := makeTestArgs(gLocal)
		task, err := NewTask(DESCRIBE_DEPLOYMENT_TEMPLATE_LABEL, &args)
		require.NotEmpty(t, task)
		require.Empty(t, err)
		resp, err := task.Run(nil)
		require.NotEmpty(t, err)
		require.Empty(t, resp)
	}
	// params unsupported
	for _, g := range testSuites {
		gLocal := g
		gLocal.NetworkType = "main_net"
		args := makeTestArgs(gLocal)
		task, err := NewTask(DESCRIBE_DEPLOYMENT_TEMPLATE_LABEL, &args)
		require.NotEmpty(t, task)
		require.Empty(t, err)
		resp, err := task.Run(nil)
		require.NotEmpty(t, err)
		require.Empty(t, resp)
	}
}

func makeTestArgs(testParamGroup CreateDeploymentParamGroup) map[string]interface{} {
	var testArgs = map[string]interface{}{}
	testArgs[ADDRESS] = testAddressId
	testArgs[REQUEST_UUID] = uuid.NewString()
	testArgs["Chain"] = testParamGroup.Chain
	testArgs["NetworkType"] = testParamGroup.NetworkType
	testArgs["ProverType"] = testParamGroup.ProverType
	testArgs["DA"] = testParamGroup.DA
	testArgs["CloudVendor"] = testParamGroup.CloudVendor
	testArgs["Region"] = testParamGroup.Region
	testArgs["Zone"] = testParamGroup.Zone
	testArgs["KeyPairId"] = testParamGroup.KeyPairId
	testArgs["DeploymentName"] = testParamGroup.Name
	testArgs["AccessKeyID"] = testParamGroup.AccessKey.AccessKeyID
	testArgs["AccessKeySecret"] = testParamGroup.AccessKey.AccessKeySecret
	return testArgs
}

func TestCreateAndDeleteDeployment(t *testing.T) {
	ids := []uint{}
	oldIds := map[uint]struct{}{}
	oldDeps, err := db.ListDeployment(nil, testAddressId)
	require.Empty(t, err)
	for _, dep := range oldDeps {
		oldIds[dep.ID] = struct{}{}
	}

	for _, g := range testSuites {
		args := makeTestArgs(g)
		task, err := NewTask(CREATE_DEPLOYMENT_LABEL, &args)
		require.NotEmpty(t, task)
		require.Empty(t, err)
		concreteTask, ok := task.(*CreateDeploymentTask)
		require.True(t, ok)
		require.NotEmpty(t, concreteTask)
		resp, err := task.Run(&gin.Context{})
		require.Empty(t, err)
		require.NotEmpty(t, resp)
	}
	newDeps, err := db.ListDeployment(nil, testAddressId)
	require.Empty(t, err)
	for _, dep := range newDeps {
		if _, ok := oldIds[dep.ID]; !ok {
			ids = append(ids, dep.ID)
		}
	}
wait_loop:
	for {
		time.Sleep(10 * time.Second)
		deps, err := db.ListDeployment(ids, testAddressId)
		require.Empty(t, err)
		require.NotEmpty(t, deps)
		for _, dep := range deps {
			if dep.Status != UNKNOWN {
				continue wait_loop
			}
		}
		break
	}
	for _, id := range ids {
		args := map[string]interface{}{
			"DeploymentId":    id,
			"AccessKeyID":     testToken.AccessKeyID,
			"AccessKeySecret": testToken.AccessKeySecret,
			ADDRESS:           testAddressId,
			REQUEST_UUID:      "123",
		}
		task, err := NewTask(DELETE_DEPLOYMENT_LABEL, &args)
		require.Empty(t, err)
		require.NotEmpty(t, task)
		concreteTask, ok := task.(*DeleteDeploymentTask)
		require.True(t, ok)
		require.NotEmpty(t, concreteTask)
		resp, err := task.Run(&gin.Context{})
		require.Empty(t, err)
		require.NotEmpty(t, resp)
	}

	for {
		time.Sleep(10 * time.Second)
		deps, err := db.ListDeployment(ids, testAddressId)
		require.Empty(t, err)
		if len(deps) == 0 {
			return
		}
	}
}
