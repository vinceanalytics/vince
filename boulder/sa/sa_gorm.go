package sa

import (
	"github.com/gernest/vince/boulder/core"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(
		&issuedNameModel{},
		&core.Certificate{},
		&core.CertificateStatus{},
		&core.FQDNSet{},
		&orderModel{},
		&orderToAuthzModel{},
		&requestedNameModel{},
		&orderFQDNSet{},
		&authzModel{},
		&recordedSerialModel{},
		&precertificateModel{},
		&keyHashModel{},
		&incidentModel{},
	)
	return db, err
}
