// 2017.08.28 rjj: Simple Go REST server for minimal address book
// 	Work done as partial fullfilment towards earning an interview for a Go (golang) development position

package main

import(
	//_ "encoding/json"
	//_ "errors"
	//"fmt"
	//_ "io"
	//"log"
	//"net/http"
	"os"
	//_ "path"
	//"strconv"

	//_ "golang.org/x/net/context"

	//"github.com/gorilla/handlers"
	//"github.com/gorilla/mux"
	//_ "github.com/satori/go.uuid"

	"github.com/rjj-work/yum-address-book"
)


func main() {

	a := addressbook.Application{}
	a.Initialize(
		os.Getenv( "YUM_ADDRESSBOOK_DB_USERNAME" ),
		os.Getenv( "YUM_ADDRESSBOOK_DB_PASSWORD" ),
		os.Getenv( "YUM_ADDRESSBOOK_DB_NAME" ),
	)

	// Run from localhost for
	a.Run( os.Getenv( "YUM_ADDRESSBOOK_HOST_PORT" ) )
}
