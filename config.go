// 2017.08.28 rjj: Basic application configuration data

package addressbook

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

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
	a.initializeRoutes()
}

func (a *Application) Run(hostPort string) {
	log.Fatal(http.ListenAndServe( hostPort , a.Router))
}

// **************** ROUTES ****************
func (a *Application) initializeRoutes() {
	a.Router.Handle("/favicon.ico", http.NotFoundHandler()).Methods("GET")
	a.Router.HandleFunc( "/addressbookentries", a.getAddressBookEntries).Methods("GET")
	a.Router.HandleFunc( "/addressbookentry", a.addAddressBookEntry).Methods("POST")
	a.Router.HandleFunc( "/addressbookentry/{id:[0-9]+}", a.getAddressBookEntry).Methods("GET")
	a.Router.HandleFunc( "/addressbookentry/{id:[0-9]+}", a.updateAddressBookEntry).Methods("PUT")
	a.Router.HandleFunc( "/addressbookentry/{id:[0-9]+}", a.deleteAddressBookEntry).Methods("DELETE")

	a.Router.HandleFunc( "/csvexport", a.getAddressBookEntriesAsCSV).Methods("GET")
	a.Router.HandleFunc( "/csvimport", a.addAddressBookEntriesFromCSV).Methods("POST")
}


// **************** HANDLERS ****************
// We package up error responses as JSON also
func respondWithError(w http.ResponseWriter, statusCode int, responseMessage string) {
	respondWithJSON(w, statusCode, map[string]string{"error": responseMessage})
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)

	// It is an interesting problem if the json.Marshal fails
	if nil != err {
		// Try to send back an error
		log.Fatal("FAILED: json.Marshal(%v)", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}


/*
	CRUD:
	C: Create
	R: Read
	U: Update
	D: Delete
 */
// Not quite the R in cRud, since this may return multiple entries.
func (a *Application) getAddressBookEntries(w http.ResponseWriter, r *http.Request) {
	// []*AddressBookEntry
	abes, err := a.DB.ListAddressBookEntries()
	//log.Printf("getAddressBookEntries:: abes(%v) err(%v)", abes, err)
	if nil != err {
		// NO data is OK
		if sql.ErrNoRows != err {
			// Something bad happened
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	respondWithJSON(w, http.StatusOK, abes)
}

// The C in Crud
func (a *Application) addAddressBookEntry(w http.ResponseWriter, r *http.Request) {
	var abe AddressBookEntry

	jd := json.NewDecoder(r.Body)
	if err := jd.Decode(&abe); nil != err {
		respondWithError(w, http.StatusBadRequest,
			fmt.Sprintf("Invalid request payload (%v)", r.Body))
		return
	}
	defer r.Body.Close()

	id, err := a.DB.AddAddressBookEntry(&abe)
	//log.Printf("addAddressBookEntry:: id(%v), err(%v)\n", id, err)

	if nil != err {
		respondWithError(w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to add AddressBookEntry (%v)", abe))
		return
	}
	// Add new ID to data
	abe.ID = id
	respondWithJSON(w, http.StatusCreated, abe)
}

// The R in cRud
func (a *Application) getAddressBookEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"],10,64)
	//log.Printf("getAddressBookEntry:: vars(%+v), err(%v)\n", vars, err)
	if nil != err {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Bad AddressBookEntry ID (%v)", vars["id"]))
		return
	}

	abe, err := a.DB.GetAddressBookEntry(id)
	if nil != err {
		// Differentiate between NO data found vs. another issue
		if sql.ErrNoRows == err {
			// We executed correctly, but the user asked for what is not there
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("AddressBookEntry with ID (%d) not found.", id))
		} else {
			// Something bad happened
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, abe)
}

// The U in crUd
func (a *Application) updateAddressBookEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"],10,64)
	//log.Printf("updateAddressBookEntry:: vars(%+v), err(%v)\n", vars, err)
	if nil != err {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Bad AddressBookEntry ID (%v)", vars["id"]))
		return
	}

	var abe AddressBookEntry
	jd := json.NewDecoder(r.Body)
	if err := jd.Decode(&abe); nil != err {
		respondWithError(w, http.StatusBadRequest,
			fmt.Sprintf("Invalid request payload (%v)", r.Body))
		return
	}
	defer r.Body.Close()

	abe.ID = id
	err = a.DB.UpdateAddressBookEntry(&abe)

	if nil != err {
		respondWithError(w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to update AddressBookEntry (%v)", abe))
		return
	}
	// We *could* do a DB read to find by this id, but I'll cheat here and *assume* the insert
	//	worked, and the returned ID with no errors, means we are OK to use this local data
	respondWithJSON(w, http.StatusOK, abe)
}

func (a *Application) deleteAddressBookEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"],10,64)
	//log.Printf("updateAddressBookEntry:: vars(%+v), err(%v)\n", vars, err)
	if nil != err {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Bad AddressBookEntry ID (%v)", vars["id"]))
		return
	}

	err = a.DB.DeleteAddressBookEntry(id)

	if nil != err {
		respondWithError(w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to delete AddressBookEntry (%v)", id))
		return
	}
	respondWithJSON(w, http.StatusOK, nil)
}

// **************** CSV Handlers ****************

// Write the records out.
// We include a header, so there will always be atleast one record returned
func respondWithCSV(w http.ResponseWriter, statusCode int, abes []*AddressBookEntry) {

	b := &bytes.Buffer{}
	csvWriter := csv.NewWriter( b )

	var abeStrings []string

	// TODO: Extract list of names from addressbook package
	abeHeaders := []string{
		"ID", "Firstname", "Lastname", "Email", "Phone",
	}
	err := csvWriter.Write(abeHeaders)
	if nil != err {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, abe := range abes {
		abeStrings = []string{
			fmt.Sprintf("%d", abe.ID),
			abe.Firstname,
			abe.Lastname,
			abe.Email,
			abe.Phone,
		}
		//log.Printf("respondWithCSV:: abe(%v)", abe)
		//log.Printf("respondWithCSV:: abeStrings(%v)", abeStrings)
		err = csvWriter.Write(abeStrings)

		// It is an interesting problem if the CSV write fails
		if nil != err {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	csvWriter.Flush()

	w.Header().Set("Content-Type", "text/csv")
	t := time.Now()
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment;filename=AddressBoookExport-%04d%s%02dT%02d%02d%02d.csv",
			t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()) )
	w.WriteHeader(statusCode)
	w.Write(b.Bytes())
}

func (a *Application) getAddressBookEntriesAsCSV(w http.ResponseWriter, r *http.Request) {
	// []*AddressBookEntry
	abes, err := a.DB.ListAddressBookEntries()
	//log.Printf("getAddressBookEntries:: abes(%v) err(%v)", abes, err)
	if nil != err {
		// NO data is OK
		if sql.ErrNoRows != err {
			// Something bad happened
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	respondWithCSV(w, http.StatusOK, abes)
}

// The request body should be our new addresses.
// Questions to consider:
// - Should current contents of the DB be dropped ?
//		I.e. delete existing records before import.		NO for this iteration
// - Will we always received a header record ?
// - Do we assume the ID field will be present ?		Assume YES
// - What do we do with the ID field ?	Ignore it
// - Should the entire import be atomic ?
//		Yes it should, but this implementation is not.
//		I.e. errors will be noted, but not terminate the input.
//		A better approach would be to allow specifying this behavior - but not today.
func (a *Application) addAddressBookEntriesFromCSV(w http.ResponseWriter, r *http.Request) {
	// []*AddressBookEntry
	csvReader := csv.NewReader( r.Body )

	var cnt, errCnt int
	msgs := []string{}
	for {
		cnt++
		// Extract a record
		record, err := csvReader.Read()
		if io.EOF == err {
			break
		}
		if nil != err {
			msgs = append(msgs, fmt.Sprintf("CSV read error, rcd# %d: %v", err))
			errCnt++
			continue
		}
		//log.Printf("addCSV:: record: %v", record)
		// 1 == cnt, check for header record
		if 1 == cnt && isABFCSVHeader(record) {
			// skip this one
			continue
		}

		// ! headerFound, insert record
		_, err = a.DB.AddAddressBookEntry( &AddressBookEntry{
			Firstname: record[1],
			Lastname:  record[2],
			Email:     record[3],
			Phone:     record[4],
			} )
		if nil != err {
			// TODO: Limit number of Add ABE failed msgs
			msgs = append(msgs, fmt.Sprintf("Add ABE failed, rcd# %d: %v", err))
			errCnt++
		}
	}

	// Respond with message of number successful and number failed imports
	msgs = append(msgs, fmt.Sprintf("Processed %d input records.  Errors: %d", cnt, errCnt))

	if cnt == errCnt {
		respondWithJSON(w, http.StatusBadRequest, msgs)
		return
	}
	if 0 < errCnt {	// We know errCnt < cnt, unless something really odd happened
		respondWithJSON(w, http.StatusPartialContent, msgs)
		return
	}
	respondWithJSON(w, http.StatusOK, msgs)
}

// TODO: verify all expected header fields are present in the expected order
func isABFCSVHeader(rcd []string) (bool) {
	return (5 <= len(rcd)) && ("ID" == rcd[0])
}


// **************** DATABASE SETUP ****************
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
