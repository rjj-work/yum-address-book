// 2017.08.28 rjj: Define the basis structure for an Address Book Entry

package addressbook


// Even though the DB has the create timestamp, leaving this out for now
type AddressBookEntry struct {
	ID        int64     `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
}

// AddressBookDatabase provides thread-safe access to a database of contacts.
type AddressBookDatabase interface {
	// ListAddressBookEntries returns a list of AddressBookEntries, ordered by lastname, firstname.
	ListAddressBookEntries() ([]*AddressBookEntry, error)

	// GetAddressBookEntry retrieves a AddressBookEntry by its ID.
	GetAddressBookEntry(id int64) (*AddressBookEntry, error)

	// AddAddressBookEntry saves a given AddressBookEntry, assigning it a new ID.
	AddAddressBookEntry(abe *AddressBookEntry) (id int64, err error)

	// DeleteAddressBookEntry removes a given AddressBookEntry by its ID.
	DeleteAddressBookEntry(id int64) error

	// UpdateAddressBookEntry updates the entry for a given AddressBookEntry.
	UpdateAddressBookEntry(abe *AddressBookEntry) error

	// Close closes the database, freeing up any available resources.
	Close()

	// This is added for testing, should not be called otherwise
	DropAddressBookTable() (error)
	TruncateTableAddressBookEntry() (error)
}
