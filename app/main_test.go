// 2017.08.28 rjj: Test file

package main_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	_ "log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rjj-work/yum-address-book"
)

var a addressbook.Application

func TestMain(m *testing.M) {
	a = addressbook.Application{}
	a.Initialize(
		os.Getenv( "TEST_YUM_ADDRESSBOOK_DB_USERNAME" ),
		os.Getenv( "TEST_YUM_ADDRESSBOOK_DB_PASSWORD" ),
		os.Getenv( "TEST_YUM_ADDRESSBOOK_DB_NAME" ),
	)

	// a.Initialize ensures the db and table exists

	code := m.Run()

	//resetTable()

	os.Exit( code )
}


func resetTable() {
	a.DB.TruncateTableAddressBookEntry()
}

func executeRequest(r *http.Request) (*httptest.ResponseRecorder) {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, r)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
    if expected != actual {
        t.Errorf("Expected response code %d != %d (actual)\n", expected, actual)
    }
}


func TestEmptyTable(t *testing.T) {
    resetTable()

    req, _ := http.NewRequest("GET", "/addressbookentries", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    if body := response.Body.String(); body != "[]" {
        t.Errorf("Expected an empty array. Got %s", body)
    }
}

// GET /addressbookentry/1
func TestGetNonExistentAddressBookEntry(t *testing.T) {
    resetTable()

    req, _ := http.NewRequest("GET", "/addressbookentry/1", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNotFound, response.Code)

    var m map[string]string
    json.Unmarshal(response.Body.Bytes(), &m)
    expected := fmt.Sprintf("AddressBookEntry with ID (%d) not found.", 1)
    if m["error"] != expected {
        t.Errorf("Expected 'error' key of the response to be set to '%s'. Got '%s'", expected, m["error"])
    }
}

func checkIt(t *testing.T, field string, expected, actual interface{}) {
	if expected != actual {
		t.Errorf("Expected %s '%s', but got '%v'", field, expected, actual)
	}
}
// POST /addressbookentry
func TestCreateAddressBookEntry(t *testing.T) {
    resetTable()

    payload := []byte(`{"firstname":"Fn1","lastname":"Ln1","email":"Fn1.Ln1@example.com","phone":"(123)456-7890"}`)

    req, _ := http.NewRequest("POST", "/addressbookentry", bytes.NewBuffer(payload))
    response := executeRequest(req)

    checkResponseCode(t, http.StatusCreated, response.Code)

    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

	checkIt(t, "firstname", "Fn1", m["firstname"])
	checkIt(t, "lastname", "Ln1", m["lastname"])
	checkIt(t, "email", "Fn1.Ln1@example.com", m["email"])
	checkIt(t, "phone", "(123)456-7890", m["phone"])

    // the id is compared to 1.0 because JSON unmarshaling converts numbers to
    // floats, when the target is a map[string]interface{}
    if m["id"] != 1.0 {
        t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
    }
}

// Create a list of AddressBookEntries on demand
func generateAddressBookEntries(t *testing.T, cnt int) ([]*addressbook.AddressBookEntry) {
	abes := []*addressbook.AddressBookEntry{}

	for i := 0; i < cnt; i++ {
		// Need to declare new ABE in the loop, since we append the reference
		var abe addressbook.AddressBookEntry
		abe.ID = int64(i)
		abe.Firstname = fmt.Sprintf( "Fn_%d", i )
		abe.Lastname  = fmt.Sprintf( "Ln_%d", i )
		abe.Email     = fmt.Sprintf( "Fn_%d.LN_%d@example.com", i, i )

		rawPhone := fmt.Sprintf( "%010d", i )
		abe.Phone     = fmt.Sprintf( "(%s)%s-%s", rawPhone[0:3], rawPhone[3:6], rawPhone[6:10] )

		abes = append(abes, &abe)
	}
	return abes
}

// Add as many entries as requested, minimum of 1
func addAddressBookEntries(t *testing.T, cnt int) {
	if 1 > cnt {
		cnt = 1
	}

	var abe addressbook.AddressBookEntry
	abes := generateAddressBookEntries(t, cnt)
	for idx, _ := range abes {
		id, err := a.DB.AddAddressBookEntry( abes[idx] )
		if nil != err {
			t.Errorf("Failed to add AddressBookEntry on index: %d, err: %v", idx, err)
		}
		// HACK, I really wanted to just comment out this line, but then get this build error:
		//	id declared and not used
		if 1==0 {fmt.Printf( "i: Added ID: %d, ABE: %v\n", id, abe )}
	}
}

// Insert a record in the DB and see if we can read it
func TestGetAddressBookEntry(t *testing.T) {
    resetTable()
    addAddressBookEntries(t, 1)

    req, _ := http.NewRequest("GET", "/addressbookentry/1", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)
}

// PUT /addressbookentry/1
func TestUpdateAddressBookEntry(t *testing.T) {
    resetTable()
    addAddressBookEntries(t, 1)

    req, _ := http.NewRequest("GET", "/addressbookentry/1", nil)
    response := executeRequest(req)
    checkResponseCode(t, http.StatusOK, response.Code)

	// Now assuming here means we were able to read the record inserted
    var original_abe map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &original_abe)

	new_firstname := "Fn1_a"
	new_lastname  := "Ln1_a"
	new_email     := "Fn1_a.Ln1_a@example.com"
	new_phone     := "(123)456-7890"
	new_abe := fmt.Sprintf(`{"firstname":"%s","lastname":"%s","email":"%s","phone":"%s"}`,
		new_firstname, new_lastname, new_email, new_phone)
	//fmt.Printf("%v\n", new_abe)
    //payload := []byte(`{"firstname":"Fn1_a","lastname":"Ln1_a","email":"Fn1_a.Ln1_a@example.com","phone":"(123)456-7890"}`)
    payload := []byte(new_abe)

    req, _ = http.NewRequest("PUT", "/addressbookentry/1", bytes.NewBuffer(payload))
    response = executeRequest(req)
    checkResponseCode(t, http.StatusOK, response.Code)

	// OK, here means we have successful PUT, but did the data change ?
    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

    if m["id"] != original_abe["id"] {
        t.Errorf("Expected the id to remain the same (%v). Got %v",
        	original_abe["id"], m["id"])
    }

    if m["firstname"] == original_abe["firstname"] {
        t.Errorf("Expected the firstname to change from '%v' to '%v'. Got '%v'",
        	original_abe["firstname"], new_firstname, m["firstname"])
    }

    if m["lastname"] == original_abe["lastname"] {
        t.Errorf("Expected the lastname to change from '%v' to '%v'. Got '%v'",
        	original_abe["lastname"], new_lastname, m["lastname"])
    }

    if m["email"] == original_abe["email"] {
        t.Errorf("Expected the email to change from '%v' to '%v'. Got '%v'",
        	original_abe["email"], new_email, m["email"])
    }

    if m["phone"] == original_abe["phone"] {
        t.Errorf("Expected the phone to change from '%v' to '%v'. Got '%v'",
        	original_abe["phone"], new_phone, m["phone"])
    }

}


// DELETE /addressbookentry/1
//	- reset table to empty
//	- add 1 new row
//	- read the row to verify there
//	- delete the row
//	- try to read row, should not be there
func TestDeleteAddressBookEntry(t *testing.T) {
    resetTable()
    addAddressBookEntries(t, 1)

    req, _ := http.NewRequest("GET", "/addressbookentry/1", nil)
    response := executeRequest(req)
    checkResponseCode(t, http.StatusOK, response.Code)

    req, _ = http.NewRequest("DELETE", "/addressbookentry/1", nil)
    response = executeRequest(req)
    checkResponseCode(t, http.StatusOK, response.Code)

    req, _ = http.NewRequest("GET", "/addressbookentry/1", nil)
    response = executeRequest(req)
    checkResponseCode(t, http.StatusNotFound, response.Code)
}


func TestCSVExport(t *testing.T) {
	resetTable()

	const numABEs = 10
	addAddressBookEntries(t, numABEs)

	req, _ := http.NewRequest("GET", "/csvexport", nil)
	response := executeRequest(req)
    checkResponseCode(t, http.StatusOK, response.Code)

	// To be a proper test, I should extract the body of the response and verify that:
	// - There are as many csv records as expected
	// - The records are actually the ones we expect.

	r := csv.NewReader( bytes.NewReader(response.Body.Bytes()) )
	var count int
	for {
		_, err := r.Read()
		if io.EOF == err {
			break
		}
		if nil != err {
			t.Errorf("CSV read error: %v", err)
		}
		count++
	}
	if 1+numABEs != count {
		t.Errorf("Expected %d CSV records, got %d", numABEs, count)
	}
}

func TestCSVImport(t *testing.T) {
	resetTable()

	// Create a request to the import endpoint that has a known number of records
	//	- generate ABEs
	//	- encode as CSV
	//	- create request with these in the body
	// Then verify those records are now in the db
	const numABEs = 15
	abes := generateAddressBookEntries(t, numABEs)

	b := &bytes.Buffer{}
	csvWriter := csv.NewWriter( b )

	var abeStrings []string

	// TODO: Extract list of names from addressbook package
	abeHeaders := []string{
		"ID", "Firstname", "Lastname", "Email", "Phone",
	}
	err := csvWriter.Write(abeHeaders)
	if nil != err {
		t.Errorf("CSV write header failed %v", err)
	}
	for _, abe := range abes {
		abeStrings = []string{
			fmt.Sprintf("%d", abe.ID),
			abe.Firstname,
			abe.Lastname,
			abe.Email,
			abe.Phone,
		}
		err = csvWriter.Write(abeStrings)

		// It is an interesting problem if the CSV write fails
		if nil != err {
			t.Errorf("CSV write header failed %v", err)
		}
	}
	csvWriter.Flush()


	req, _ := http.NewRequest("POST", "/csvimport", b)
	req.Header.Set("Content-Type", "text/csv")
    response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Query DB for the count of records, should be numABEs
	currentABEs, err := a.DB.ListAddressBookEntries()
	if nil != err {
		t.Errorf( "TestCSVImport:: failed to read ABEs: %v", err )
	}
	if numABEs != len(currentABEs) {
		t.Errorf( "Exected %d ABEs, but found %d", numABEs, len(currentABEs) )
	}
}
