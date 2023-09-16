package db

import (
	"fmt"
	"testing"

	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
)

func TestSaveZones(t *testing.T) {
	zones := []dao.AvailabilityZone{
		{CloudVendor: dao.SURFER_CLOUD, Zone: "cn-bj2-05", Region: "cn-bj2"},
		{CloudVendor: dao.SURFER_CLOUD, Zone: "cn-bj2-04", Region: "cn-bj2"},
	}

	err := db.AutoMigrate(&dao.AvailabilityZone{})
	if err != nil {
		t.Fatalf("failed to migrate database %s", err.Error())
	}

	err = SaveZones(zones)
	if err != nil {
		t.Fatalf("create err %s", err.Error())
	}
}

func TestGetZones(t *testing.T) {
	zones, err := GetZones(dao.SURFER_CLOUD)
	if err != nil {
		t.Fatalf("find err %s", err.Error())
	}
	fmt.Printf("zones: %v\n", zones)
}
