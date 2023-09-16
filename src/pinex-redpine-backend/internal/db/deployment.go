package db

import "github.com/MetaDataLab/web3-console-backend/internal/models/dao"

func ListDeployment(ids []uint, address string) ([]dao.Deployment, error) {
	deployments := []dao.Deployment{}
	var err error
	if ids != nil {
		err = db.Where("address = ? and id in (?)", address, ids).Preload("Resources").Find(&deployments).Error
	} else {
		err = db.Where("address = ?", address).Preload("Resources").Find(&deployments).Error
	}
	return deployments, err
}

func ListAllDeployment() ([]dao.Deployment, error) {
	deployments := []dao.Deployment{}
	var err error
	err = db.Preload("Resources").Find(&deployments).Error
	return deployments, err
}

func SaveDeployment(dep *dao.Deployment) error {
	err := db.Save(dep).Error
	return err
}

func DeleteDeployment(id uint) error {
	dep := &dao.Deployment{}
	dep.ID = id
	err := db.Select("Resources").Delete(dep).Error
	return err
}
