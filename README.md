# Simple Address Book REST API

# Address Book Entries
Each entry has the following fields (minimum requirement)
- First Name
- Last Name
- Email Address
- Phone Number

## Data Considerations
It is not specificed if any or all fields must be present, however we ensure that at minimum
- First Name
- Last Name
are present (note, done thru DB NOT NULL constraint).  Email and Phone are optional.
We also do not support multiple Email address, nor multiple Phone numbers in this simple implementation.

A future iteration might consider insisting that one of email or phone be present.

## The Database
MySQL is used in this implementation.
Environment variables are used to pass in the user name, user password and database instance (schema) to be used.
This initial configuration still has some hardcode references to the schema *yum_addressbook*.
This will be addressed in a later iteration.

There is a single database table, with the unwieldy name: *addressbookentries*

The table structure and an example data row is provided here:
```
mysql> desc addressbookentries;
+-------------+------------------+------+-----+-------------------+----------------+
| Field       | Type             | Null | Key | Default           | Extra          |
+-------------+------------------+------+-----+-------------------+----------------+
| id          | int(10) unsigned | NO   | PRI | NULL              | auto_increment |
| firstname   | varchar(255)     | NO   |     | NULL              |                |
| lastname    | varchar(255)     | NO   |     | NULL              |                |
| email       | varchar(255)     | YES  |     | NULL              |                |
| phone       | text             | YES  |     | NULL              |                |
| createdDate | datetime         | YES  |     | CURRENT_TIMESTAMP |                |
+-------------+------------------+------+-----+-------------------+----------------+
6 rows in set (0.00 sec)

mysql> select * from addressbookentries;
+----+-----------+----------+-----------------------+---------------+---------------------+
| id | firstname | lastname | email                 | phone         | createdDate         |
+----+-----------+----------+-----------------------+---------------+---------------------+
|  1 | Fn_0      | Ln_0     | Fn_0.LN_0@example.com | (000)000-0000 | 2017-08-29 04:51:10 |
+----+-----------+----------+-----------------------+---------------+---------------------+
1 row in set (0.00 sec)
```



# Getting it to work
## Install
```bash
git clone https://github.com/rjj-work/yum-address-book.git
```

## Pre-run Configuration
### MySQL Database
Plese configure your MySQL (or MariaDB) database with a user with sufficient privileges to:
- Create new databases,
- Create new tables,
- Full CRUD access on the tables

### Run-time Environment
Environment variables are used to configure the Database Username, Password and Instance/Schema Name

E.g.
```bash
export YUM_ADDRESSBOOK_DB_USERNAME=gouser
export YUM_ADDRESSBOOK_DB_PASSWORD=test123
export YUM_ADDRESSBOOK_DB_NAME=yum_addressbook
export YUM_ADDRESSBOOK_HOST_PORT=":8080"
```

## Start the Applicatiion
Nothing special here, typical Go start-up
```bash
cd app
go run main.go
```
At this point a browser could be used, e.g.

```http
http://localhost:8080/addressbookentries
```

## Testing
### To run the included Go test procedures

- Export test environment variables
```bash
export TEST_YUM_ADDRESSBOOK_DB_USERNAME=gouser
export TEST_YUM_ADDRESSBOOK_DB_PASSWORD=test123
export TEST_YUM_ADDRESSBOOK_DB_NAME=yum_addressbook_test
```
- Initiate the tests
```bash
cd app
go test -v
```
Test run output:
```bash
=== RUN   TestEmptyTable
--- PASS: TestEmptyTable (0.00s)
=== RUN   TestGetNonExistentAddressBookEntry
--- PASS: TestGetNonExistentAddressBookEntry (0.00s)
=== RUN   TestCreateAddressBookEntry
--- PASS: TestCreateAddressBookEntry (0.00s)
=== RUN   TestGetAddressBookEntry
--- PASS: TestGetAddressBookEntry (0.00s)
=== RUN   TestUpdateAddressBookEntry
--- PASS: TestUpdateAddressBookEntry (0.00s)
=== RUN   TestDeleteAddressBookEntry
--- PASS: TestDeleteAddressBookEntry (0.00s)
=== RUN   TestCSVExport
--- PASS: TestCSVExport (0.00s)
=== RUN   TestCSVImport
--- PASS: TestCSVImport (0.01s)
PASS
ok  	github.com/rjj-work/yum-address-book/app	0.040s

```

### Testing using *curl*
- Start the system normally, as described above
```bash
export YUM_ADDRESSBOOK_DB_USERNAME=gouser
export YUM_ADDRESSBOOK_DB_PASSWORD=test123
export YUM_ADDRESSBOOK_DB_NAME=yum_addressbook
export YUM_ADDRESSBOOK_HOST_PORT=":8080"
cd app
go run main.go
```

#### Inserting a record via *curl*
Sample JSON data file provided *data/new-addressbookentry-01.json*

```bash
gandalf17:yum-address-book rjj$ cd data
gandalf17:data rjj$ ls -l
total 16
-rw-r--r--  1 rjj  staff  780 Aug 30 02:57 import-data.csv
-rw-r--r--  1 rjj  staff   91 Aug 30 04:34 new-addressbookentry-01.json
gandalf17:data rjj$

gandalf17:data rjj$ curl -d@new-addressbookentry-01.json -H "Content-Type: application/json" http://localhost:8080/addressbookentry
```
The server will respond back with JSON, including the new ID value.
```JSON
{"id":1,"firstname":"fn1","lastname":"ln1","email":"fn1.ln1@example.com","phone":"(123)456-7890"}
```
Note, at this point, the browser will provide useful feedback for query.

#### List DB records via *curl*
Command: (Note I prefer to specify the -v option)
```bash
gandalf17:data rjj$ curl -v http://addressbookentries
```
Console output
```bash
* Rebuilt URL to: http://addressbookentries/
* Could not resolve host: addressbookentries
* Closing connection 0
curl: (6) Could not resolve host: addressbookentries
gandalf17:data rjj$ curl -v http://localhost:8080/addressbookentries
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8080 (#0)
> GET /addressbookentries HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Wed, 30 Aug 2017 09:10:47 GMT
< Content-Length: 99
<
* Connection #0 to host localhost left intact
[{"id":1,"firstname":"fn1","lastname":"ln1","email":"fn1.ln1@example.com","phone":"(123)456-7890"}]gandalf17:data rjj$
```

#### GET a single record via *curl*
Command:
```bash
gandalf17:data rjj$ curl -v http://localhost:8080/addressbookentry/1
```
Console output
```bash
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8080 (#0)
> GET /addressbookentry/1 HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Wed, 30 Aug 2017 09:13:27 GMT
< Content-Length: 97
<
* Connection #0 to host localhost left intact
{"id":1,"firstname":"fn1","lastname":"ln1","email":"fn1.ln1@example.com","phone":"(123)456-7890"}gandalf17:data rjj$
```

#### UPDATE via *curl*
A second JSON data file is provided that has updates to the firstname, lastname and email fields.
```bash
gandalf17:data rjj$ curl -v -X PUT -d@update-addressbookentry-01.json -H "Content-Type application/json" http://localhost:8080/addressbookentry/1
```
Console output
```bash
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8080 (#0)
> PUT /addressbookentry/1 HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.54.0
> Accept: */*
> Content-Type: application/json
> Content-Length: 125
>
* upload completely sent off: 125 out of 125 bytes
< HTTP/1.1 500 Internal Server Error
< Content-Type: application/json
< Date: Wed, 30 Aug 2017 09:26:38 GMT
< Content-Length: 121
<
* Connection #0 to host localhost left intact
{"error":"Failed to update AddressBookEntry ({1 fn1_edited ln1_edited fn1_edited.ln1_edited@example.com (123)456-7890})"}gandalf17:data rjj$```
```

#### DELETE via *curl*
Delete is straight forward.
- First show data is there via the list URL:
```bash
gandalf17:data rjj$ curl http://localhost:8080/addressbookentries
[{"id":1,"firstname":"fn1_edited","lastname":"ln1_edited","email":"fn1_edited.ln1_edited@example.com","phone":"(123)456-7890"}]gandalf17:data rjj$
```
- Next Delete ID==1
```bash
gandalf17:data rjj$ curl -X DELETE http://localhost:8080/addressbookentry/1
```
- Requery and list URL shows nothing
```bash
gandalf17:data rjj$ curl http://localhost:8080/addressbookentries
[]gandalf17:data rjj$
```
