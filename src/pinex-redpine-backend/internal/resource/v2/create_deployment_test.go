package v2

import (
	"testing"

	"github.com/MetaDataLab/web3-console-backend/internal/db"
	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var params = &commonParams{
	nameBase: "web3_test_resource",
	region:   "cn-bj2",
	//az:       "cn-bj2-02",
	token: AccessToken{
		AccessKeyID:     "4ZDughY8BbJ2NOEOTiG0u96p3rFZR0Nly",
		AccessKeySecret: "IJ2ozFmpjc3KKShxD2c1ZzH9Ceife1m8FCV2RAbjBLMI",
	},
}

var myKey = "zkhh1h"
var weijiaKey = "ef2036"

func TestMain(m *testing.M) {
	config.InitConfig("/root/go/src/web3-console-backend/.vscode/config.yaml")
	log.InitGlobalLogger(&config.GConf.LogConfig, zap.AddCallerSkip(1), zap.AddStacktrace(zap.DebugLevel))
	db.Init()
	m.Run()
}

func TestCreateZkEvm(t *testing.T) {
	dep := NewDeployment(dao.POLYGON_ZKEVM, dao.ETH, dao.DEV_NET, dao.EXCLUSIVE, dao.SURFER_CLOUD,
		"cn-bj2", "cn-bj2-05",
		AccessToken{"ak123", "IWillNotTellYou"}, "",
		WithKeyPairId(myKey), WithName("test_zkevm"),
	)
	tmpl, err := dep.LoadTemplate()
	require.Empty(t, err)
	require.NotEmpty(t, tmpl)
	resources, err := dep.Build()
	require.Empty(t, err)
	require.NotEmpty(t, resources)
}

func TestCreateOpStack(t *testing.T) {
	dep := NewDeployment(dao.OP_STACK, dao.ETH, dao.DEV_NET, dao.EXCLUSIVE, dao.SURFER_CLOUD,
		"cn-bj2", "cn-bj2-05",
		AccessToken{"ak123", "IWillNotTellYou"}, "",
		WithKeyPairId(weijiaKey), WithName("test_op"),
	)
	tmpl, err := dep.LoadTemplate()
	require.Empty(t, err)
	require.NotEmpty(t, tmpl)
	resources, err := dep.Build()
	require.Empty(t, err)
	require.NotEmpty(t, resources)
}

func TestGetPrice(t *testing.T) {
	dep := NewDeployment(dao.POLYGON_ZKEVM, dao.ETH, dao.DEV_NET, dao.EXCLUSIVE, dao.SURFER_CLOUD,
		"cn-bj2", "cn-bj2-05",
		AccessToken{"ak123", "IWillNotTellYou"}, "",
		WithKeyPairId(myKey), WithName("test_zkevm"),
	)
	tmpl, err := dep.LoadTemplate()
	require.Empty(t, err)
	require.NotEmpty(t, tmpl)
	priceTags, err := dep.LoadPrice()
	require.Empty(t, err)
	require.NotEmpty(t, priceTags)
}

func TestDescribePgsql(t *testing.T) {
	db := NewDb(params, nil, "", "")
	err := db.stopDb("upgsql-nqeexrmkgsc")
	require.Empty(t, err)
}
