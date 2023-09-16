package db

import "github.com/MetaDataLab/web3-console-backend/internal/models/dao"

func Migrate() error {
	var migrates = []any{&dao.DbTemplate{}, &dao.DeploymentTemplate{}, &dao.NodeTemplate{}, &dao.DiskTemplate{},
		&dao.MachineTemplate{}, &dao.ImageTemplate{}, &dao.Userdata{}, &dao.EipTemplate{}, &dao.UserDataArg{}, &dao.CreationOrder{},
		dao.Deployment{}, dao.Resource{}}
	return db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(migrates...)
}
