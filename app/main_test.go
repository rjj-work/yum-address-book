// 2017.08.28 rjj: Test file

package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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
    expected := "AddressBookEntry not found"
    if m["error"] != expected {
        t.Errorf("Expected the 'error' key of the response to be set to '%s'. Got '%s'", expected, m["error"])
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

// Add as many entries as requested, minimum of 1
func addAddressBookEntries(t *testing.T, cnt int) {
	if 1 > cnt {
		cnt = 1
	}

	var abe addressbook.AddressBookEntry

	for i := 0; i < cnt; i++ {
		abe.Firstname = fmt.Sprintf( "Fn_%d", i )
		abe.Lastname  = fmt.Sprintf( "Ln_%d", i )
		abe.Email     = fmt.Sprintf( "Fn_%d.LN_%d@example.com", i, i )

		rawPhone := fmt.Sprintf( "%010d", i )
		abe.Phone     = fmt.Sprintf( "(%s)%s-%s", rawPhone[0:3], rawPhone[3:6], rawPhone[6:10] )

		id, err := a.DB.AddAddressBookEntry( &abe )
		if nil != err {
			t.Errorf("Failed to add AddressBookEntry on index: %d, err: %v", i, err)
		}
		fmt.Printf( "i: Added ID: %d, ABE: %v\n", id, abe )
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