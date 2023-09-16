package resource

import (
	"testing"

	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/stretchr/testify/require"
)

var params = &commonParams{
	nameBase: "web3_test_resource",
	region:   "cn-bj2",
	zone:     "cn-bj2-05",
	token: Token{
		AccessKeyId:     "ak123",
		SecretAccessKey: "IWillNotTellYou",
	},
}

var testTemplate = "#cloud-config\n" +
	"\n" +
	"cloud_final_modules:\n" +
	"    - scripts-per-boot\n" +
	"\n" +
	"write_files:\n" +
	"- content: |\n" +
	"    #!/bin/bash\n" +
	"    echo \"Hello World.  The time is now $(date -R)!\" >> /root/success.txt \n" +
	"  path: /var/lib/cloud/scripts/per-boot/success.sh\n" +
	"  permissions: \"0755\"\n"

func TestMain(m *testing.M) {
	cfg := &config.LogConfig{}
	log.InitGlobalLogger(cfg)
	m.Run()
}

func TestCreateVm(t *testing.T) {
	vpcId, subnetId, ips, err := getRandomIps(params, 1)
	require.Empty(t, err)
	require.Len(t, ips, 1)
	if params.vpcId != "" {
		require.Equal(t, vpcId, params.vpcId)
	}
	if params.subnetId != "" {
		require.Equal(t, subnetId, params.subnetId)
	}
	params.vpcId = vpcId
	params.subnetId = subnetId
	vm := vmResource{
		commonParams: params,
		keyPairId:    "zkhh1h",
		cpu:          1,
		//2 G
		memory: 2 * 1024,
		imgId:  "uimage-cn8em8zrtxw",
		rootDisk: vmDisk{
			// 20G
			size:     20,
			diskType: "CLOUD_RSSD",
		},
		dataDisk: vmDisk{
			// 20G
			size:     20,
			diskType: "CLOUD_RSSD",
		},
		ip: ips[0],
		ud: &userdata{
			template: testTemplate,
		},
	}
	err = vm.create()
	require.Empty(t, err)
	require.NotEmpty(t, vm.vmId)
	t.Log(vm.vmId)
}

func TestCreatePg(t *testing.T) {
	pg := dbResource{
		commonParams: params,
		version:      "postgresql-13.4",
		port:         5432,
		machineType:  "o.pgsql2m.medium",
		//paramGroupId: 54,
		password: "08AACA3C990F47BE91D4AACCED90DBF5",
	}
	err := pg.create()
	require.Empty(t, err)
	require.NotEmpty(t, pg.dbId)
	require.NotEmpty(t, pg.ip)
}

func TestCreateZkEvm(t *testing.T) {
	req := CreateDeploymentRequest{
		Name:   "web3_zkevm",
		Region: "cn-bj2",
		Az:     "cn-bj2-05",
		Token: Token{
			AccessKeyId:     "ak123",
			SecretAccessKey: "IWillNotTellYou",
		},
		KeyPairId: "zkhh1h",
		Type:      POLYGON_ZKEVM,
	}
	r, passwd, err := CreateDeployment(req)
	require.Empty(t, err)
	require.NotEmpty(t, r)
	require.NotEmpty(t, passwd)
}

func TestCreateOpStack(t *testing.T) {
	req := CreateDeploymentRequest{
		Name:   "web3_op",
		Region: "cn-bj2",
		Az:     "cn-bj2-05",
		Token: Token{
			AccessKeyId:     "ak123",
			SecretAccessKey: "IWillNotTellYou",
		},
		KeyPairId: "ef2036",
		Type:      OP_STACK,
	}
	r, passwd, err := CreateDeployment(req)
	require.Empty(t, err)
	require.NotEmpty(t, r)
	require.Empty(t, passwd)
}
