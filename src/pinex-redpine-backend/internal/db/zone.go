package db

import (
	"fmt"

	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
)

func GetZones(cloud interface{}) ([]dao.AvailabilityZone, error) {
	var cloudValue dao.CloudVendor
	switch cloud := cloud.(type) {
	case string:
		cloudValue = dao.CloudVendor_value[cloud]
	case dao.CloudVendor:
		cloudValue = cloud
	default:
		return nil, fmt.Errorf("cloud type %T not supported", cloud)
	}

	zones := make([]dao.AvailabilityZone, 1000)
	result := db.Find(&zones, "cloud_vendor = ?", cloudValue)

	return zones, result.Error
}

func SaveZones(zones []dao.AvailabilityZone) error {
	result := db.Create(zones)
	return result.Error
}
