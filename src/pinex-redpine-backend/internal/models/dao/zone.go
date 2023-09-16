package dao

import "gorm.io/gorm"

type AvailabilityZone struct {
	gorm.Model
	CloudVendor `gorm:"uniqueIndex:idx_cloud_vendor_zone"`
	Zone        string `gorm:"uniqueIndex:idx_cloud_vendor_zone;size:30"`
	Region      string `gorm:"size:30"`
}
