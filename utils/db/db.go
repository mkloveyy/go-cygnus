package db

import (
	"fmt"
	"gorm.io/gorm/logger"
	"io/ioutil"

	"sigs.k8s.io/yaml"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"go-cygnus/constants"
	"go-cygnus/utils/logging"
)

var (
	Engine *gorm.DB
)

type Config struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Port     string `json:"port"`
}

func Init() {
	dbString := GetDSN()

	// init engine
	db, err := gorm.Open(mysql.Open(dbString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		logging.GetLogger("root").WithError(err).Fatal("New Engine failed")
	}

	Engine = db
}

func GetConfig() (dbConfig Config) {
	dbYmlFile := fmt.Sprintf("%s/database.yml", constants.ConfigPath)
	dbContent, readErr := ioutil.ReadFile(dbYmlFile)

	if readErr != nil {
		panic("Read database.yml error, file path:" + dbYmlFile)
	}

	dbConfig = Config{}
	unmarshalErr := yaml.Unmarshal(dbContent, &dbConfig)

	if unmarshalErr != nil {
		panic("Unmarshal database content error.")
	}

	return
}

func GetDSN() string {
	dbConfig := GetConfig()

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?"+
		"charset=utf8&parseTime=true&loc=Local&timeout=10s",
		dbConfig.Username, dbConfig.Password,
		dbConfig.Host, dbConfig.Port, dbConfig.Database)
}

// TxErrDefer commit or revert tx based on err, passing err to return
func TxErrDefer(tx *gorm.DB, err error) error {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}

	var commitErr error

	if err != nil {
		commitErr = tx.Rollback().Error
	} else {
		commitErr = tx.Commit().Error
	}

	if commitErr == nil {
		return errors.WithStack(err)
	}

	return errors.Wrap(errors.WithStack(err), "TxErrDefer err: "+commitErr.Error())
}

func IsDBError(err error) bool {
	return err != nil && err != gorm.ErrRecordNotFound
}
