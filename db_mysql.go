// 2017.08.28 rjj: Interface to MySQL DB
// Note that I've struggled with the convention to have database tables pluralized
//	but the conflict in the sematics of 'adressbook' vs. 'addressbooks'
// 	In the end I've punted by adopting the table name 'addressbookentries'

package addressbook

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
)

var createTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS yum_addressbook DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE yum_addressbook;`,
	`CREATE TABLE IF NOT EXISTS addressbookentries (
		id INT UNSIGNED NOT NULL AUTO_INCREMENT,
		firstname VARCHAR(255) NOT NULL,
		lastname VARCHAR(255) NOT NULL,
		email VARCHAR(255) NULL,
		phone TEXT NULL,
		createdDate datetime DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id)
	);`,
}

// mysqlDB persists AddressBookEntries to a MySQL instance.
type mysqlDB struct {
	conn *sql.DB

	list     *sql.Stmt
	insert   *sql.Stmt
	get      *sql.Stmt
	update   *sql.Stmt
	delete   *sql.Stmt
	// drop is for testing only
	drop     *sql.Stmt
	truncate *sql.Stmt
}

// Ensure mysqlDB conforms to the AddressBookDatabase interface.
var _ AddressBookDatabase = &mysqlDB{}

type MySQLConfig struct {
	// Optional.
	Username, Password string

	// Host of the MySQL instance.
	// If set, UnixSocket should be unset.
	Host string

	// Schema the MySQL instance
	Schema string

	// Port of the MySQL instance.
	// If set, UnixSocket should be unset.
	Port int

	// UnixSocket is the filepath to a unix socket.
	// If set, Host and Port should be unset.
	UnixSocket string
}

// dataStoreName returns a connection string suitable for sql.Open.
func (c MySQLConfig) dataStoreName(databaseName string) string {
	var cred string
	// [username[:password]@]
	if c.Username != "" {
		cred = c.Username
		if c.Password != "" {
			cred = cred + ":" + c.Password
		}
		cred = cred + "@"
	}

	if c.UnixSocket != "" {
		return fmt.Sprintf("%sunix(%s)/%s", cred, c.UnixSocket, databaseName)
	}
	return fmt.Sprintf("%stcp([%s]:%d)/%s", cred, c.Host, c.Port, databaseName)
}

// newMySQLDB creates a new AddressBookDatabase backed by a given MySQL server.
func newMySQLDB(config MySQLConfig) (AddressBookDatabase, error) {
	// Check database and table exists. If not, create it.
	if err := config.ensureTableExists(); err != nil {
		return nil, err
	}

	conn, err := sql.Open("mysql", config.dataStoreName("yum_addressbook"))
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("mysql: could not establish a good connection: %v", err)
	}

	db := &mysqlDB{
		conn: conn,
	}

	// Prepared statements. The actual SQL queries are in the code near the
	// relevant method (e.g. AddAddressBookEntry).
	if db.list, err = conn.Prepare(listStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare list: %v", err)
	}
	if db.get, err = conn.Prepare(getStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare get: %v", err)
	}
	if db.insert, err = conn.Prepare(insertStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare insert: %v", err)
	}
	if db.update, err = conn.Prepare(updateStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare update: %v", err)
	}
	if db.delete, err = conn.Prepare(deleteStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare delete: %v", err)
	}
	if db.drop, err = conn.Prepare(dropStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare drop: %v", err)
	}
	if db.truncate, err = conn.Prepare(truncateStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare truncate: %v", err)
	}

	return db, nil
}

// Close closes the database, freeing up any resources.
func (db *mysqlDB) Close() {
	db.conn.Close()
}

// rowScanner is implemented by sql.Row and sql.Rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanAddressBookEntry reads a AddressBookEntry from a sql.Row or sql.Rows
func scanAddressBookEntry(s rowScanner) (*AddressBookEntry, error) {
	var (
		id          int64
		firstname   sql.NullString
		lastname    sql.NullString
		email       sql.NullString
		phone       sql.NullString
	)
	if err := s.Scan(&id, &firstname, &lastname, &email, &phone); err != nil {
		return nil, err
	}

	abe := &AddressBookEntry{
		ID:          id,
		Firstname:   firstname.String,
		Lastname:    lastname.String,
		Email:       email.String,
		Phone:       phone.String,
	}
	return abe, nil
}

const listStatement = `SELECT * FROM addressbookentries ORDER BY lastname, firstname`

// ListAddressBookEntrys returns a list of AddressBookEntries, ordered by name.
func (db *mysqlDB) ListAddressBookEntries() ([]*AddressBookEntry, error) {
	rows, err := db.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var AddressBookEntries []*AddressBookEntry
	for rows.Next() {
		abe, err := scanAddressBookEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		AddressBookEntries = append(AddressBookEntries, abe)
	}

	return AddressBookEntries, nil
}


const getStatement = "SELECT * FROM addressbookentries WHERE id = ?"

// GetAddressBookEntry retrieves a addressbook by its ID.
func (db *mysqlDB) GetAddressBookEntry(id int64) (*AddressBookEntry, error) {
	abe, err := scanAddressBookEntry(db.get.QueryRow(id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("mysql: could not find address book entry with id %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get address book entry: %v", err)
	}
	return abe, nil
}

const insertStatement = `
  INSERT INTO addressbookentries (
    firstname, lastname, email, phone
  ) VALUES (?, ?, ?, ?)`

// AddAddressBookEntry saves a given addressbook, assigning it a new ID.
func (db *mysqlDB) AddAddressBookEntry(abe *AddressBookEntry) (id int64, err error) {
	r, err := execAffectingOneRow(db.insert, abe.Firstname, abe.Lastname, abe.Email, abe.Phone)
	if err != nil {
		return 0, err
	}

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("mysql: could not get last insert ID: %v", err)
	}
	return lastInsertID, nil
}

const deleteStatement = `DELETE FROM addressbookentries WHERE id = ?`

// DeleteAddressBookEntry removes a given addressbook by its ID.
func (db *mysqlDB) DeleteAddressBookEntry(id int64) error {
	if id == 0 {
		return errors.New("mysql: address book entry with unassigned ID passed into deleteAddressBookEntry")
	}
	_, err := execAffectingOneRow(db.delete, id)
	return err
}

const updateStatement = `
  UPDATE addressbookentries
  SET firstname=?, lastname=?, email=?, phone=?
  WHERE id = ?`

// UpdateAddressBookEntry updates the entry for a given addressbook.
func (db *mysqlDB) UpdateAddressBookEntry(abe *AddressBookEntry) error {
	if abe.ID == 0 {
		return errors.New("mysql: address book entry with unassigned ID passed into UpdateAddressBookEntry")
	}

	_, err := execAffectingOneRow(db.update, abe.Firstname, abe.Lastname, abe.Email, abe.Phone, abe.ID)
	return err
}


// ensureTableExists checks the table exists. If not, it creates it.
func (config MySQLConfig) ensureTableExists() error {
	conn, err := sql.Open("mysql", config.dataStoreName(""))
	if err != nil {
		return fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	defer conn.Close()

	// Check the connection.
	if conn.Ping() == driver.ErrBadConn {
		return fmt.Errorf("mysql: could not connect to the database. " +
			"could be bad address, or this address is not whitelisted for access.")
	}

	if _, err := conn.Exec("USE yum_addressbook"); err != nil {
		// MySQL error 1049 is "database does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1049 {
			fmt.Printf("ensureTableExists:: USE connection 2 err(%v)\n", mErr)
			return createTable(conn)
		}
		return fmt.Errorf("mysql: USE error %v", err)
	}

	if _, err := conn.Exec("DESCRIBE addressbookentries"); err != nil {
		// MySQL error 1146 is "table does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1146 {
			return createTable(conn)
		}
		// Unknown error.
		return fmt.Errorf("mysql: could not connect to the database: %v", err)
	}
	return nil
}

// createTable creates the table, and if necessary, the database.
func createTable(conn *sql.DB) error {
	for _, stmt := range createTableStatements {
		fmt.Printf("createTable:: Exec(%v)\n", stmt )
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// execAffectingOneRow executes a given statement, expecting one row to be affected.
func execAffectingOneRow(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	r, err := stmt.Exec(args...)
	if err != nil {
		return r, fmt.Errorf("mysql: could not execute statement: %v", err)
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return r, fmt.Errorf("mysql: could not get rows affected: %v", err)
	} else if rowsAffected != 1 {
		return r, fmt.Errorf("mysql: expected 1 row affected, got %d", rowsAffected)
	}
	return r, nil
}


// TESTING SUPPORT
const dropStatement = `
  DROP TABLE addressbookentries`

// DropTableAddressBookEntry drops the table from the DB
func (db *mysqlDB) DropAddressBookTable() error {
	_, err := db.conn.Exec(dropStatement)
	return err
}

const truncateStatement = `
  TRUNCATE TABLE addressbookentries`

// TruncateTableAddressBookEntry deleted all data in the DB, and ID sequence is reset to 1.
func (db *mysqlDB) TruncateTableAddressBookEntry() error {
	_, err := db.conn.Exec(truncateStatement)
	return err
}