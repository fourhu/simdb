// Package db A simple library to persist structs in json file and perform queries and CRUD operations
package simdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.aporeto.io/elemental"
	"sync"
)

var ErrRecordNotFound = errors.New("record not found")
var ErrUpdateFailed = errors.New("update failed, no record(s) to update")

// empty represents an empty result
var empty interface{}

// query describes a query
type query struct {
	key, operator string
	value         interface{}
}

// Driver contains all the state of the db.
type Driver struct {
	dir               string    //directory name to store the db
	queries           [][]query // nested queries
	queryIndex        int
	queryMap          map[string]QueryFunc // contains query functions
	jsonContent       interface{}          // copy of original decoded json data for further processing
	errors            []error              // contains all the errors when processing
	originalJSON      interface{}          // actual json when opening the json file
	isOpened          bool
	entityDealingWith elemental.Identity
	mutex             *sync.Mutex
}

// New creates a new database driver. Accepts the directory name to store the db files.
// If the passed directory not exist then will create one.
//
//	driver, err:=db.New("customer")
func New(dir string) (*Driver, error) {
	driver := &Driver{
		dir:      dir,
		queryMap: loadDefaultQueryMap(),
		mutex:    &sync.Mutex{},
	}
	err := createDirIfNotExist(dir)
	return driver, err
}

// Open will open the json db based on the entity passed.
// Once the file is open you can apply where conditions or get operation.
//
//	driver.Open(Customer{})
//
// Open returns a pointer to Driver, so you can chain methods like Where(), Get(), etc
func (d *Driver) Open(entity elemental.Identifiable) *Driver {
	d.queries = nil
	d.entityDealingWith = entity.Identity()
	db, err := d.openDB(entity)
	d.originalJSON = db
	d.jsonContent = d.originalJSON
	d.isOpened = true
	if err != nil {
		d.addError(err)
	}
	return d
}

// Errors will return errors encountered while performing any operations
func (d *Driver) Errors() []error {
	return d.errors
}

// Insert the entity to the json db. Insert will identify the type of the
// entity and insert the entity to the specific json file based on the type of the entity.
// If the db file not exist then will create a new db file
//
//		customer:=Customer {
//			CustID:"CUST1",
//			Name:"sarouje",
//			Address: "address",
//			Contact: Contact {
//				Phone:"45533355",
//				Email:"someone@gmail.com",
//			},
//		}
//	 err:=driver.Insert(customer)
func (d *Driver) Insert(entity elemental.Identifiable) (err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.entityDealingWith = entity.Identity()
	err = d.readAppend(entity)
	return
}

// Where builds a where clause to filter the records.
//
//	driver.Open(Customer{}).Where("custid","=","CUST1")
func (d *Driver) Where(key, cond string, val interface{}) *Driver {
	q := query{
		key:      key,
		operator: cond,
		value:    val,
	}
	if d.queryIndex == 0 && len(d.queries) == 0 {
		qq := []query{}
		qq = append(qq, q)
		d.queries = append(d.queries, qq)
	} else {
		d.queries[d.queryIndex] = append(d.queries[d.queryIndex], q)
	}

	return d
}

// Get the result from the json db as an array. If no where condition then return all the data from json db
//
// Get based on a where condition
//
//	driver.Open(Customer{}).Where("name","=","sarouje").Get()
//
// Get all records
//
//	driver.Open(Customer{}).Get()
func (d *Driver) Get() *Driver {
	if !d.isDBOpened() {
		return d
	}
	if len(d.queries) > 0 {
		d.processQuery()
	} else {
		d.jsonContent = d.originalJSON
	}
	d.queryIndex = 0

	return d
}

// First return the first record matching the condtion.
//
//	driver.Open(Customer{}).Where("custid","=","CUST1").First()
func (d *Driver) First() *Driver {
	if !d.isDBOpened() {
		return d
	}
	records := d.Get().RawArray()
	if len(records) > 0 {
		d.jsonContent = records[0]
	} else {
		d.addError(fmt.Errorf("no records to perform First operation"))
	}

	return d
}

// Raw will return the data in map type
func (d *Driver) Raw() interface{} {
	return d.jsonContent
}

// RawArray will return the data in map array type
func (d *Driver) RawArray() []interface{} {
	if aa, ok := d.jsonContent.([]interface{}); ok {
		return aa
	}
	return nil
}

// AsEntity will converts the map to the passed structure pointer.
// should call this function after calling Get() or First(). This function will convert
// the result of Get or First operation to the passed structure type
// 'output' variable should be a pointer to a structure or stucture array. Function returns error in case
// of any errors in conversion.
//
// First()
//
//	var custOut Customer
//	err:=driver.Open(Customer{}).First().AsEntity(&custOut)
//	fmt.Printf("%#v", custOut)
//	this function will fill the custOut with the values from the map
//
// Get()
//
//	var customers []Customer
//	err:=driver.Open(Customer{}).Get().AsEntity(&customers)
func (d *Driver) AsEntity(output interface{}) (err error) {
	if !d.isDBOpened() {
		return fmt.Errorf("should call Open() before calling AsEntity()")
	}
	switch t := d.jsonContent.(type) {
	case []interface{}:
		if len(t) <= 0 {
			return ErrRecordNotFound
		}
	case interface{}:
		if t == nil {
			return ErrRecordNotFound
		}
	}

	outByte, err := json.Marshal(d.jsonContent)
	if err != nil {
		return err
	}
	err = json.Unmarshal(outByte, output)
	return
}

// Update the json data based on the id field/value pair
//
//	customerToUpdate:=driver.Open(Customer{}).Where("custid","=","CUST1").First()
//	customerToUpdate.Name="Sony Arouje"
//	err:=driver.Update(customerToUpdate)
//
// Should not change the ID field when updating the record.
func (d *Driver) Update(entity elemental.Identifiable) (err error) {
	d.queries = nil
	d.entityDealingWith = entity.Identity()
	entityID := entity.Identifier()
	couldUpdate := false
	// entName, _ := d.getEntityName()

	d.mutex.Lock()
	defer d.mutex.Unlock()
	records := d.Open(entity).Get().RawArray()

	if len(records) > 0 {
		for indx, item := range records {
			if record, ok := item.(map[string]interface{}); ok {
				if v, ok := record[IdentifierKey]; ok && fmt.Sprintf("%v", v) == fmt.Sprintf("%v", entityID) {
					records[indx] = entity
					couldUpdate = true
				}
			}
		}
	}
	if couldUpdate {
		err = d.writeAll(records)
	} else {
		err = ErrUpdateFailed
	}

	return
}

// Upsert function will try updating the passed entity. If no records to update then
// do the Insert operation.
//
//	   	customer := Customer{
//			CustID:  "CU4",
//			Name:    "Sony Arouje",
//			Address: "address",
//			Contact: Contact{
//				Phone: "45533355",
//				Email: "someone@gmail.com",
//			},
//		}
//	 driver.Upsert(customer)
func (d *Driver) Upsert(entity elemental.Identifiable) (err error) {
	err = d.Update(entity)
	if errors.Is(err, ErrUpdateFailed) {
		err = d.Insert(entity)
	}
	return
}

// Delete the record from the json db based on the id field/value pair
//
//	  custToDelete:=Customer {
//		   CustID:"CUST1",
//	  }
//	  err:=driver.Delete(custToDelete)
func (d *Driver) Delete(entity elemental.Identifiable) (err error) {
	d.queries = nil
	d.entityDealingWith = entity.Identity()
	entityID := entity.Identifier()
	entName, _ := d.getEntityName()
	couldDelete := false
	newRecordArray := make([]interface{}, 0, 0)

	d.mutex.Lock()
	defer d.mutex.Unlock()
	records := d.Open(entity).Get().RawArray()

	if len(records) > 0 {
		for indx, item := range records {
			if record, ok := item.(map[string]interface{}); ok {
				if v, ok := record[IdentifierKey]; ok && v != entityID {
					records[indx] = entity
					newRecordArray = append(newRecordArray, record)
				} else {
					couldDelete = true
				}
			}
		}
	}
	if couldDelete {
		err = d.writeAll(newRecordArray)
	} else {
		err = fmt.Errorf("failed to delete, unable to find any %s record with %s %s", entName, IdentifierKey, entityID)
	}
	return
}
