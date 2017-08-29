# Simple Address Book REST API

# Address Book Entries
Each entry has the following fields (minimum requirement)
- First Name
- Last Name
- Email Address
- Phone Number

## Data considerations
It is not specificed if any or all fields must be present, however we will ensure that at minimum
- First Name
- Last Name
are present.  Email and Phone will be optional.
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
## Start it up

## Go testing
- Export some environment variables used during testing.
```bash
export TEST_YUM_ADDRESSBOOK_DB_USERNAME=gouser
export TEST_YUM_ADDRESSBOOK_DB_PASSWORD=test123
export TEST_YUM_ADDRESSBOOK_DB_NAME=yum_addressbook
```


## CURL testing
