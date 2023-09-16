package db

import (
	"strings"
	"testing"

	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func toArgKey(r dao.Role, suffix string) string {
	return strings.ToLower(r.String()) + "_" + suffix
}

var zkEvmProverTemplate = "#cloud-config\n" +
	"\n" +
	"cloud_final_modules:\n" +
	"    - scripts-per-boot\n" +
	"\n" +
	"write_files:\n" +
	"- content: |\n" +
	"    #!/bin/bash\n" +
	"    cd /root/ && ./launchProver.sh \"%s\" \"%s\" \"%s\" \"%s\"\n" +
	"  path: /var/lib/cloud/scripts/per-boot/run_zkevm_prover.sh\n" +
	"  permissions: \"0755\"\n"

var zkEvmNodeTemplate = "#cloud-config\n" +
	"\n" +
	"cloud_final_modules:\n" +
	"    - scripts-per-boot\n" +
	"\n" +
	"write_files:\n" +
	"- content: |\n" +
	"    #!/bin/bash\n" +
	"    cp -r /root/* /data/ && cd /data && nohup ./raas.sh \"%s\" \"%s\" \"%s\" \"%s\" \"%s\" 2>&1 >> raas.out &\n" +
	"  path: /var/lib/cloud/scripts/per-boot/run_zkevm_node.sh\n" +
	"  permissions: \"0755\"\n"

var opNodeVmTemplate = "#cloud-config\n" +
	"\n" +
	"cloud_final_modules:\n" +
	"    - scripts-per-boot\n" +
	"\n" +
	"write_files:\n" +
	"- content: |\n" +
	"    #!/bin/bash\n" +
	"    cd /home/ubuntu/network/op/ && su ubuntu -c \"./init.sh DEFAULT\" && su ubuntu -c ./start.sh \n" +
	"  path: /var/lib/cloud/scripts/per-boot/run_opstack_node.sh\n" +
	"  permissions: \"0755\"\n"

var starkNetProverTemplate = "#cloud-config\n" +
	"\n" +
	"cloud_final_modules:\n" +
	"    - scripts-per-boot\n" +
	"\n" +
	"write_files:\n" +
	"- content: |\n" +
	"    #!/bin/bash\n" +
	"    echo \"Hello World.  The time is now $(date -R)!\"\n" +
	"  path: /var/lib/cloud/scripts/per-boot/run_zkevm_prover.sh\n" +
	"  permissions: \"0755\"\n"

func TestMain(m *testing.M) {
	config.InitConfig("/root/go/src/web3-console-backend/.vscode/config.yaml")
	log.InitGlobalLogger(&config.GConf.LogConfig, zap.AddCallerSkip(1), zap.AddStacktrace(zap.DebugLevel))
	Init()
	m.Run()
}

func TestGetTemplate(t *testing.T) {
	tmpl, err := GetTemplate(dao.POLYGON_ZKEVM, dao.ETH, dao.DEV_NET, dao.SURFER_CLOUD, dao.EXCLUSIVE, "cn-bj2")
	require.NotEmpty(t, tmpl)
	require.Empty(t, err)
}

func TestGetOpTemplate(t *testing.T) {
	tmpl, err := GetTemplate(dao.OP_STACK, dao.ETH, dao.DEV_NET, dao.SURFER_CLOUD, dao.EXCLUSIVE, "cn-bj2")
	require.NotEmpty(t, tmpl)
	require.Empty(t, err)
}

func TestSaveTemplate(t *testing.T) {
	tmpl := &dao.DeploymentTemplate{
		BaseTemplateModel: dao.BaseTemplateModel{
			Name: "polygon_zkevm",
		},
		ChainType: dao.POLYGON_ZKEVM,
		Dbs: []dao.DbTemplate{{
			BaseTemplateModel: dao.BaseTemplateModel{
				Name: "polygon_zkevm_db",
			},
			CloudVendor: dao.SURFER_CLOUD,
			Role:        dao.POLYGON_ZKEVM_DB,
			DbType:      "postgresql-13.4",
			MachineType: "o.pgsql2m.medium",
			Port:        5432,
			DiskTemplate: dao.DiskTemplate{
				BaseTemplateModel: dao.BaseTemplateModel{
					Name: "test_db_disk",
				},
				CloudVendor: dao.SURFER_CLOUD,
				Type:        "CLOUD_RSSD",
				Size:        512,
			},
			Mode: "Normal",
		}},
		Nodes: []dao.NodeTemplate{
			{
				BaseTemplateModel: dao.BaseTemplateModel{
					Name: "polygon_zkevm_prover",
				},
				Role: dao.POLYGON_ZKEVM_PROVER,
				DiskTemplate: dao.DiskTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "polygon_zkevm_prover_data_disk",
					},
					CloudVendor: dao.SURFER_CLOUD,
					Type:        "CLOUD_RSSD",
					Size:        800,
				},
				MachineTemplate: dao.MachineTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "polygon_zkevm_prover_machine_template",
					},
					Platform:    "Amd/Auto",
					CloudVendor: dao.SURFER_CLOUD,
					CoreNum:     96,
					Memory:      768 * 1024,
					MachineType: "O",
				},
				ImageTemplate: dao.ImageTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "polygon_zkevm_prover_image",
					},
					CloudVendor: dao.SURFER_CLOUD,
					ImageId:     "uimage-n9l61ivcli0",
					Region:      "cn-bj2",
					AZ:          "cn-bj2-05",
					DiskTemplate: dao.DiskTemplate{
						BaseTemplateModel: dao.BaseTemplateModel{
							Name: "polygon_zkevm_prover_root_disk",
						},
						CloudVendor: dao.SURFER_CLOUD,
						Type:        "CLOUD_RSSD",
						Size:        200,
					},
					Userdatas: []dao.Userdata{
						{
							BaseTemplateModel: dao.BaseTemplateModel{
								Name: "polygon_zkevm_prover_userdata",
							},
							DaType:      dao.ETH,
							NetworkType: dao.DEV_NET,
							ProverType:  dao.EXCLUSIVE,
							Content:     zkEvmProverTemplate,
							Args: []dao.UserDataArg{
								{
									BaseTemplateModel: dao.BaseTemplateModel{
										Name: toArgKey(dao.POLYGON_ZKEVM_NODE, "ip"),
									},
									Position: 0,
								},
								{
									BaseTemplateModel: dao.BaseTemplateModel{
										Name: toArgKey(dao.POLYGON_ZKEVM_DB, "ip"),
									},
									Position: 1,
								},
								{
									BaseTemplateModel: dao.BaseTemplateModel{
										Name: toArgKey(dao.POLYGON_ZKEVM_DB, "user"),
									},
									Position: 2,
								},
								{
									BaseTemplateModel: dao.BaseTemplateModel{
										Name: toArgKey(dao.POLYGON_ZKEVM_DB, "password"),
									},
									Position: 3,
								},
							},
						},
					},
				},
			},
			{
				BaseTemplateModel: dao.BaseTemplateModel{
					Name: "polygon_zkevm_node",
				},
				Role: dao.POLYGON_ZKEVM_NODE,
				DiskTemplate: dao.DiskTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "polygon_zkevm_node_data_disk",
					},
					CloudVendor: dao.SURFER_CLOUD,
					Type:        "CLOUD_RSSD",
					Size:        500,
				},
				MachineTemplate: dao.MachineTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "polygon_zkevm_node_machine_template",
					},
					Platform:    "Intel/Auto",
					CloudVendor: dao.SURFER_CLOUD,
					CoreNum:     8,
					Memory:      16 * 1024,
					MachineType: "O",
				},
				ImageTemplate: dao.ImageTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "polygon_zkevm_node_image",
					},
					CloudVendor: dao.SURFER_CLOUD,
					ImageId:     "uimage-n5gj0rjxda0",
					Region:      "cn-bj2",
					AZ:          "cn-bj2-05",
					DiskTemplate: dao.DiskTemplate{
						BaseTemplateModel: dao.BaseTemplateModel{
							Name: "polygon_zkevm_node_root_disk",
						},
						CloudVendor: dao.SURFER_CLOUD,
						Type:        "CLOUD_RSSD",
						Size:        200,
					},
					Userdatas: []dao.Userdata{
						{
							BaseTemplateModel: dao.BaseTemplateModel{
								Name: "polygon_zkevm_node_userdata",
							},
							DaType:      dao.ETH,
							NetworkType: dao.DEV_NET,
							ProverType:  dao.EXCLUSIVE,
							Content:     zkEvmNodeTemplate,
							Args: []dao.UserDataArg{
								{
									BaseTemplateModel: dao.BaseTemplateModel{
										Name: toArgKey(dao.POLYGON_ZKEVM_PROVER, "ip"),
									},
									Position: 0,
								},
								{
									BaseTemplateModel: dao.BaseTemplateModel{
										Name: toArgKey(dao.POLYGON_ZKEVM_DB, "user"),
									},
									Position: 1,
								},
								{
									BaseTemplateModel: dao.BaseTemplateModel{
										Name: toArgKey(dao.POLYGON_ZKEVM_DB, "password"),
									},
									Position: 2,
								},
								{
									BaseTemplateModel: dao.BaseTemplateModel{
										Name: toArgKey(dao.POLYGON_ZKEVM_DB, "ip"),
									},
									Position: 3,
								},
								{
									BaseTemplateModel: dao.BaseTemplateModel{
										Name: toArgKey(dao.POLYGON_ZKEVM_DB, "port"),
									},
									Position: 4,
								},
							},
						},
					},
				},
				EipTemplate: &dao.EipTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "polygon_zkevm_node_eip",
					},
					BindWidth: 10,
				},
			},
		},
		CreationOrders: []dao.CreationOrder{
			{
				Role:        dao.POLYGON_ZKEVM_DB,
				Sequence:    0,
				ShouldBlock: true,
			},
			{
				Role:        dao.POLYGON_ZKEVM_NODE,
				Sequence:    1,
				ShouldBlock: true,
			},
			{
				Role:        dao.POLYGON_ZKEVM_PROVER,
				Sequence:    2,
				ShouldBlock: true,
			},
		},
	}
	err := SaveTemplate(tmpl)
	require.Empty(t, err)
}

func TestSaveOpTemplate(t *testing.T) {
	tmpl := &dao.DeploymentTemplate{
		BaseTemplateModel: dao.BaseTemplateModel{
			Name: "op_stack",
		},
		ChainType: dao.OP_STACK,
		Nodes: []dao.NodeTemplate{
			{
				BaseTemplateModel: dao.BaseTemplateModel{
					Name: "op_stack_node",
				},
				Role: dao.OP_STACK_NODE,
				DiskTemplate: dao.DiskTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "op_stack_node_data_disk",
					},
					CloudVendor: dao.SURFER_CLOUD,
					Type:        "CLOUD_RSSD",
					Size:        100,
				},
				MachineTemplate: dao.MachineTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "op_stack_node_machine_template",
					},
					Platform:    "Intel/Auto",
					CloudVendor: dao.SURFER_CLOUD,
					CoreNum:     2,
					Memory:      4 * 1024,
					MachineType: "O",
				},
				ImageTemplate: dao.ImageTemplate{
					BaseTemplateModel: dao.BaseTemplateModel{
						Name: "op_stack_node_image",
					},
					CloudVendor: dao.SURFER_CLOUD,
					ImageId:     "uimage-n5hw1xusxvw",
					Region:      "cn-bj2",
					AZ:          "cn-bj2-05",
					DiskTemplate: dao.DiskTemplate{
						BaseTemplateModel: dao.BaseTemplateModel{
							Name: "op_stack_node_root_disk",
						},
						CloudVendor: dao.SURFER_CLOUD,
						Type:        "CLOUD_RSSD",
						Size:        50,
					},
					Userdatas: []dao.Userdata{
						{
							BaseTemplateModel: dao.BaseTemplateModel{
								Name: "op_stack_node_userdata",
							},
							DaType:      dao.ETH,
							NetworkType: dao.DEV_NET,
							ProverType:  dao.EXCLUSIVE,
							Content:     opNodeVmTemplate,
						},
					},
				},
			},
		},
		CreationOrders: []dao.CreationOrder{
			{
				Role:        dao.OP_STACK_NODE,
				Sequence:    0,
				ShouldBlock: true,
			},
		},
	}
	err := SaveTemplate(tmpl)
	require.Empty(t, err)
}
