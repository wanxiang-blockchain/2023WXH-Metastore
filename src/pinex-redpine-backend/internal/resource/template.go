package resource

var gb = 1024 * 1024 * 1024
var mb = 1024 * 1024

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

func zkEvmProverVm() vmResource {
	return vmResource{
		commonParams: &commonParams{},
		cpu:          96,
		// 768G
		memory: 768 * 1024,
		rootDisk: vmDisk{
			// 200G
			size:     200,
			diskType: "CLOUD_RSSD",
		},
		dataDisk: vmDisk{
			// 800G
			size:     800,
			diskType: "CLOUD_RSSD",
		},
		ud: &userdata{
			template: zkEvmProverTemplate,
		},
		imgId:       "uimage-n5h4jrfsbzb",
		cpuPlatform: "Amd/Auto",
	}
}

func zkEvmNodeVm() vmResource {
	return vmResource{
		commonParams: &commonParams{},
		cpu:          8,
		// 16G
		memory: 16 * 1024,
		rootDisk: vmDisk{
			// 200G
			size:     200,
			diskType: "CLOUD_RSSD",
		},
		dataDisk: vmDisk{
			// 500G
			size:     500,
			diskType: "CLOUD_RSSD",
		},
		ud: &userdata{
			template: zkEvmNodeTemplate,
		},
		imgId: "uimage-n5gj0rjxda0",
	}
}

func opNodeVm() vmResource {
	return vmResource{
		commonParams: &commonParams{},
		cpu:          2,
		// 4G
		memory: 4 * 1024,
		rootDisk: vmDisk{
			// 20G
			size:     50,
			diskType: "CLOUD_RSSD",
		},
		dataDisk: vmDisk{
			// 100G
			size:     100,
			diskType: "CLOUD_RSSD",
		},
		ud: &userdata{
			template: opNodeVmTemplate,
		},
		imgId: "uimage-n5hw1xusxvw",
		// use default
		// cpuPlatform: "Intel/Auto",
	}
}

func pgDb() dbResource {
	return dbResource{
		version:     "postgresql-13.4",
		port:        5432,
		dbMode:      "Normal",
		diskSize:    512,
		machineType: "o.pgsql2m.medium",
	}
}

// func starkNetProverVm() vmResource {
// 	return vmResource{
// 		commonParams: &commonParams{},
// 		cpu:          96,
// 		// 768G
// 		memory: 768 * 1024,
// 		rootDisk: vmDisk{
// 			// 200G
// 			size:     200,
// 			diskType: "CLOUD_RSSD",
// 		},
// 		dataDisk: vmDisk{
// 			// 500G
// 			size:     500,
// 			diskType: "CLOUD_RSSD",
// 		},
// 		ud: &userdata{
// 			template: starkNetProverTemplate,
// 		},
// 		cpuPlatform: "Amd/Auto",
// 	}
// }
