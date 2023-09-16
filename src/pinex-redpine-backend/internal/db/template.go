package db

import (
	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
)

type ChainType = dao.ChainType
type DaType = dao.DaType
type NetworkType = dao.NetworkType
type CloudVendor = dao.CloudVendor
type ProverType = dao.ProverType

func GetTemplate(chainType ChainType, daType DaType, networkType NetworkType,
	cloud CloudVendor, proverType ProverType, region string) (*dao.DeploymentTemplate, error) {
	deployment := &dao.DeploymentTemplate{}
	err := db.Where(&dao.DeploymentTemplate{ChainType: chainType}).Preload("CreationOrders").
		Preload("Dbs", "cloud_vendor = ?", cloud).Preload("Nodes").
		Preload("Dbs.DiskTemplate", "cloud_vendor = ?", cloud).
		Preload("Nodes.DiskTemplate", "cloud_vendor = ?", cloud).Preload("Nodes.EipTemplate").
		Preload("Nodes.MachineTemplate", "cloud_vendor = ?", cloud).Preload("Nodes.ImageTemplate", "cloud_vendor = ? and region = ?", cloud, region).
		Preload("Nodes.ImageTemplate.DiskTemplate", "cloud_vendor = ?", cloud).
		Preload("Nodes.ImageTemplate.Userdatas", "da_type = ? and network_type = ? and prover_type = ?", daType, networkType, proverType).
		Preload("Nodes.ImageTemplate.Userdatas.Args").
		Find(deployment).Error
	return deployment, err
}

func SaveTemplate(tmpl *dao.DeploymentTemplate) error {
	err := db.Create(tmpl).Error
	return err
}
