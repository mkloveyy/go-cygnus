package models

import (
	"flag"
	"os"

	"go-cygnus/utils/db"
	"go-cygnus/utils/logging"
)

var syncDB bool

func init() {
	flag.BoolVar(&syncDB, "syncdb", false, "start app with initializing database")
}

func SyncDB() {
	// Sync db models automatically
	if syncDB {
		logging.GetLogger("root").Info("running syncdb procedure")
		tx := db.Engine.Debug()

		err := tx.AutoMigrate(&Action{}, &Account{})
		if err != nil {
			logging.GetLogger("root").WithError(err).Error("auto migrate")
		}

		logging.GetLogger("root").Info("done syncdb, exit")

		// NOTE: exit will not run any defer
		os.Exit(0)
	}
}
