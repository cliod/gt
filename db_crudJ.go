// @author  dreamlu
package der

import (
	"github.com/dreamlu/go-tool/tool/result"
)

// implement DBCrud
// form data
type DbCrudJ struct {
	// attributes
	InnerTables []string    // inner join tables
	LeftTables  []string    // left join tables
	Table       string      // table name
	Model       interface{} // table model, like User{}
	ModelData   interface{} // table model data, like var user User{}, it is 'user'

	// pager info
	ClientPage int64 // page number
	EveryPage  int64 // Number of pages per page
}

// create
func (c *DbCrudJ) Create(data interface{}) error {

	return CreateDataJ(data)
}

// update
func (c *DbCrudJ) Update(data interface{}) error {

	return UpdateDataJ(data)
}

// delete
func (c *DbCrudJ) Delete(id string) error {

	return DeleteDataByName(c.Table, "id", id)
}

// search
// pager info
// clientPage : default 1
// everyPage : default 10
func (c *DbCrudJ) GetBySearch(params map[string][]string) (pager result.Pager, err error) {

	return GetDataBySearch(c.Model, c.ModelData, c.Table, params)
}

// by id
func (c *DbCrudJ) GetByID(id string) error {

	//DB.AutoMigrate(&c.Model)
	return GetDataByID(c.ModelData, id)
}

// the same as search
// more tables
func (c *DbCrudJ) GetMoreBySearch(params map[string][]string) (pager result.Pager, err error) {

	return GetMoreDataBySearch(c.Model, c.ModelData, params, c.InnerTables, c.LeftTables)
}

// common sql
// through sql get data
func (c *DbCrudJ) GetDataBySQL(sql string, args ...interface{}) error {

	return GetDataBySQL(c.ModelData, sql, args[:]...)
}

// common sql
// through sql get data
// args not include limit ?, ?
// args is sql and sqlNt common params
func (c *DbCrudJ) GetDataBySearchSQL(sql, sqlNt string, args ...interface{}) (pager result.Pager, err error) {

	return GetDataBySQLSearch(c.ModelData, sql, sqlNt, c.ClientPage, c.EveryPage, args)
}

// delete by sql
func (c *DbCrudJ) DeleteBySQL(sql string, args ...interface{}) error {

	return DeleteDataBySQL(sql, args[:]...)
}

// update by sql
func (c *DbCrudJ) UpdateBySQL(sql string, args ...interface{}) error {

	return UpdateDataBySQL(sql, args[:]...)
}

// create by sql
func (c *DbCrudJ) CreateBySQL(sql string, args ...interface{}) error {

	return CreateDataBySQL(sql, args[:]...)
}
