// 2017.08.28 rjj: Basic application configuration data

package addressbook

import (
	//"database/sql"
	//"database/sql/driver"
	//"fmt"
	"log"

	"github.com/gorilla/mux"
)


type Application struct {
	Router	*mux.Router
	DB		AddressBookDatabase
}

func (a *Application) Initialize(user, passwd, dbname string) {
	var err error

	// [START sql]
	sqlConfig := SQLConfig{
		Username: user,
		Password: passwd,
		Instance: dbname,
		Port: 3306,
	}

	// Uncomment if you need to verify env vars are as you expect
	//log.Printf( "sqlConfig: %+v\n", sqlConfig )

	a.DB, err = configureSQL( sqlConfig )
	// [END sql]
	if nil != err {
		log.Fatal( err )
	}

	a.Router = mux.NewRouter()
}

func (a *Application) Run(hostPort string) {

}

// ******** ROUTES and HANDLERS ****************


// ******** DATABASE SETUP ****************
type SQLConfig struct {
	Username, Password, Instance string
	Port int
}

func configureSQL(config SQLConfig) (AddressBookDatabase, error) {
	// Running locally.
	return newMySQLDB( MySQLConfig{
		Username: config.Username,
		Password: config.Password,
		Schema: config.Instance,
		Host:     "localhost",
		// 3306 conflicts with local MySQL instance, so use different port for proxy
		// Port:     3306,
		Port:     config.Port,
	})
}
