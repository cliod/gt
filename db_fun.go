// package gt

package gt

import (
	"bytes"
	"fmt"
	"github.com/dreamlu/gt/tool/mock"
	reflect2 "github.com/dreamlu/gt/tool/reflect"
	"github.com/dreamlu/gt/tool/result"
	sq "github.com/dreamlu/gt/tool/sql"
	"github.com/dreamlu/gt/tool/type/cmap"
	"github.com/dreamlu/gt/tool/type/te"
	"github.com/dreamlu/gt/tool/util"
	"github.com/dreamlu/gt/tool/util/str"
	"reflect"
	"strconv"
	"strings"
)

//======================return tag=============================
//=============================================================

var coMap = cmap.CMap{}

// select * replace
// select more tables
// tables : table name / table alias name
// 主表放在tables中第一个, 紧接着为主表关联的外键表名(无顺序)
func GetMoreTableColumnSQL(model interface{}, tables ...string) (sql string) {
	typ := reflect.TypeOf(model)
	key := typ.PkgPath() + typ.Name()
	sql = coMap.Get(key)
	if sql != "" {
		//Logger().Info("[USE coMap GET ColumnSQL]")
		return
	}
	var buf bytes.Buffer
	//typ := reflect.TypeOf(model)
	GetReflectTagMore(typ, &buf, tables[:]...)
	sql = string(buf.Bytes()[:buf.Len()-1]) //去点,
	coMap.Add(key, sql)
	//Logger().Info("[ADD ColumnSQL TO coMap]")
	return
}

// 层级递增解析tag, more tables
func GetReflectTagMore(ref reflect.Type, buf *bytes.Buffer, tables ...string) {

	var (
		oTag, tag, tagTable string
		b                   bool
	)
	if ref.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < ref.NumField(); i++ {
		tag = ref.Field(i).Tag.Get("json")
		if tag == "" {
			GetReflectTagMore(ref.Field(i).Type, buf, tables[:]...)
			continue
		}
		if oTag, tag, tagTable, b = sq.GtTag(ref.Field(i).Tag, tag); b == true {
			continue
		}
		if tagTable != "" {
			buf.WriteString("`")
			buf.WriteString(tagTable)
			buf.WriteString("`.`")
			buf.WriteString(tag)
			//buf.WriteString("`,")
			buf.WriteString("` as ")
			if oTag != "" && oTag != "-" {
				buf.WriteString(oTag)
			} else {
				buf.WriteString(tag)
			}
			buf.WriteString(",")
			continue
		}

		if b = otherTableTagSQL(oTag, tag, buf, tables...); b == false {
			buf.WriteString("`")
			buf.WriteString(tables[0])
			buf.WriteString("`.`")
			buf.WriteString(tag)
			buf.WriteString("`,")
		}
	}
}

// if there is tag gt and json, select json tag first
// 多表的其他表解析处理
func otherTableTagSQL(oTag, tag string, buf *bytes.Buffer, tables ...string) bool {
	// foreign tables column
	for _, v := range tables {
		if strings.Contains(tag, v+"_id") {
			break
		}
		// tables
		switch {
		case strings.HasPrefix(tag, v+"_") &&
			// 下面两种条件参考db_test.go==>TestGetReflectTagMore()
			!strings.Contains(tag, "_id") &&
			!strings.EqualFold(v, tables[0]):
			buf.WriteString("`")
			buf.WriteString(v)
			buf.WriteString("`.`")
			buf.Write([]byte(tag)[len(v)+1:])
			buf.WriteString("` as ")
			if oTag != "" && oTag != "-" {
				buf.WriteString(oTag)
			} else {
				buf.WriteString(tag)
			}
			buf.WriteString(",")
			return true
		}
	}
	return false
}

// 根据model中表模型的json标签获取表字段
// 将select* 中'*'变为对应的字段名
// 增加别名,表连接问题
func GetColSQLAlias(model interface{}, alias string) (sql string) {
	typ := reflect.TypeOf(model)
	key := typ.PkgPath() + typ.Name()
	sql = coMap.Get(key)
	if sql != "" {
		//Logger().Info("[USE coMap GET ColumnSQL]")
		return
	}
	var buf bytes.Buffer
	GetReflectTagAlias(typ, &buf, alias)
	sql = string(buf.Bytes()[:buf.Len()-1]) //去掉点,
	coMap.Add(key, sql)
	return
}

// 层级递增解析tag, 别名
func GetReflectTagAlias(ref reflect.Type, buf *bytes.Buffer, alias string) {

	if ref.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < ref.NumField(); i++ {
		tag := ref.Field(i).Tag.Get("json")
		if tag == "" {
			GetReflectTagAlias(ref.Field(i).Type, buf, alias)
			continue
		}
		// sub sql
		gtTag := ref.Field(i).Tag.Get("gt")
		if strings.Contains(gtTag, str.GtSubSQL) {
			continue
		}
		buf.WriteString(alias)
		buf.WriteString(".`")
		buf.WriteString(tag)
		buf.WriteString("`,")
	}
}

// 根据model中表模型的json标签获取表字段
// 将select* 变为对应的字段名
func GetColSQL(model interface{}) (sql string) {
	typ := reflect.TypeOf(model)
	key := typ.PkgPath() + typ.Name()
	sql = coMap.Get(key)
	if sql != "" {
		//Logger().Info("[USE coMap GET ColumnSQL]")
		return
	}
	var buf bytes.Buffer
	//typ := reflect.TypeOf(model)
	GetReflectTag(reflect.TypeOf(model), &buf)
	sql = string(buf.Bytes()[:buf.Len()-1]) // remove ,
	coMap.Add(key, sql)
	return
}

// 层级递增解析tag
func GetReflectTag(reflectType reflect.Type, buf *bytes.Buffer) {

	if reflectType.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < reflectType.NumField(); i++ {
		tag := reflectType.Field(i).Tag.Get("json")
		if tag == "" {
			GetReflectTag(reflectType.Field(i).Type, buf)
			continue
		}
		// sub sql
		gtTag := reflectType.Field(i).Tag.Get("gt")
		if strings.Contains(gtTag, str.GtSubSQL) {
			continue
		}
		buf.WriteString("`")
		buf.WriteString(tag)
		buf.WriteString("`,")
	}
}

// get col ?
func GetColParamSQL(model interface{}) (sql string) {
	var buf bytes.Buffer

	typ := reflect.TypeOf(model)
	for i := 0; i < typ.NumField(); i++ {
		buf.WriteString("?,")
	}
	sql = string(buf.Bytes()[:buf.Len()-1]) //去掉点,
	return sql
}

// get data value
// like GetColSQL
func GetParams(data interface{}) (params []interface{}) {

	typ := reflect.ValueOf(data)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		value := typ.Field(i).Interface() //.Tag.Get("json")
		params = append(params, value)
	}
	return
}

// GT SQL struct
type GT struct {
	*Params
	// CMap
	CMaps cmap.CMap // params

	// select sql
	Select string // select sql
	From   string // only once
	Group  string // the last group
	Args   []interface{}
	ArgsNt []interface{}

	// mock data
	isMock bool
}

//=======================================sql语句处理==========================================
//===========================================================================================

// More Table
// params: innerTables is inner join tables, must even number
// params: leftTables is left join tables
// return: select sql
// table1 as main table, include other tables_id(foreign key)
func GetMoreSearchSQL(gt *GT) (sqlNt, sql string, clientPage, everyPage int64, args []interface{}) {

	sqlNt, tables := moreSql(gt)
	// select* 变为对应的字段名
	sql = strings.Replace(sqlNt, "count(`"+tables[0]+"`.id) as total_num", GetMoreTableColumnSQL(gt.Model, tables[:]...)+gt.SubSQL, 1)
	var (
		order = "`" + tables[0] + "`.id desc" // order by
		key   = ""                            // key like binary search
		bufW  bytes.Buffer                    // sql bytes connect
	)
	for k, v := range gt.CMaps {
		if v[0] == "" {
			continue
		}
		switch k {
		case str.GtClientPage, str.GtClientPageUnderLine:
			clientPage, _ = strconv.ParseInt(v[0], 10, 64)
			continue
		case str.GtEveryPage, str.GtEveryPageUnderLine:
			everyPage, _ = strconv.ParseInt(v[0], 10, 64)
			continue
		case str.GtOrder:
			order = v[0]
			continue
		case str.GtKey:
			key = v[0]
			if gt.KeyModel == nil {
				gt.KeyModel = gt.Model
			}
			var tablens = append(tables, tables[:]...)
			for k, v := range tablens {
				tablens[k] += ":" + v
			}
			// more tables key search
			sqlKey, argsKey := sq.GetMoreKeySQL(key, gt.KeyModel, tablens[:]...)
			bufW.WriteString(sqlKey)
			args = append(args, argsKey[:]...)
			continue
		case str.GtMock:
			mock.Mock(gt.Data)
			gt.isMock = true
			continue
		case "":
			continue
		}

		if b := otherTableWhere(&bufW, tables[1:], k); b != true {
			v[0] = strings.Replace(v[0], "'", "\\'", -1)
			//bufW.WriteString("`" + tables[0] + "`." + k + " = ? and ")
			bufW.WriteString("`")
			bufW.WriteString(tables[0])
			bufW.WriteString("`.`")
			bufW.WriteString(k)
			bufW.WriteString("` = ? and ")
		}
		args = append(args, v[0])
		//into:
	}

	if bufW.Len() != 0 {
		sql += fmt.Sprintf("where %s ", bufW.Bytes()[:bufW.Len()-4])
		sqlNt += fmt.Sprintf("where %s", bufW.Bytes()[:bufW.Len()-4])
		if gt.SubWhereSQL != "" {
			sql += fmt.Sprintf("and %s ", gt.SubWhereSQL)
			sqlNt += fmt.Sprintf("and %s", gt.SubWhereSQL)
		}
	} else if gt.SubWhereSQL != "" {
		sql += fmt.Sprintf("where %s ", gt.SubWhereSQL)
		sqlNt += fmt.Sprintf("where %s", gt.SubWhereSQL)
	}
	sql += fmt.Sprintf(" order by %s ", order)

	return
}

func otherTableWhere(bufW *bytes.Buffer, tables []string, k string) (b bool) {
	// other tables, except tables[0]
	for _, v := range tables {
		switch {
		case !strings.Contains(k, v+"_id") && strings.Contains(k, v+"_"):
			//bufW.WriteString("`" + table + "`.`" + string([]byte(k)[len(v)+1:]) + "` = ? and ")
			bufW.WriteString("`")
			bufW.WriteString(v)
			bufW.WriteString("`.`")
			bufW.WriteString(string([]byte(k)[len(v)+1:]))
			bufW.WriteString("` = ? and ")
			//args = append(args, v[0])
			b = true
			return
		}
	}
	return
}

// more sql
func moreSql(gt *GT) (sqlNt string, tables []string) {

	// read ram
	typ := reflect.TypeOf(gt.Model)
	keyNt := typ.PkgPath() + "/sqlNt/" + typ.Name()
	keyTs := typ.PkgPath() + "/sqlNtTables/" + typ.Name()
	sqlNt = coMap.Get(keyNt)
	if tables = strings.Split(coMap.Get(keyTs), ","); tables[0] == "" {
		tables = []string{}
	}
	if sqlNt != "" {
		//Logger().Info("[USE coMap GET ColumnSQL]")
		return
	}

	innerTables, leftTables, innerField, leftField, DBS := moreTables(gt)
	tables = append(tables, innerTables...)
	tables = append(tables, leftTables...)
	tables = util.RemoveDuplicateString(tables)

	var (
		//tables = innerTables // all tables
		bufNt bytes.Buffer // sql bytes connect
	)
	// sql and sqlCount
	bufNt.WriteString("select count(`")
	bufNt.WriteString(tables[0])
	bufNt.WriteString("`.id) as total_num from ")
	if tb := DBS[tables[0]]; tb != "" {
		bufNt.WriteString("`" + tb + "`.")
	}
	bufNt.WriteString("`")
	bufNt.WriteString(tables[0])
	bufNt.WriteString("` ")
	// inner join
	for i := 1; i < len(innerTables); i += 2 {
		bufNt.WriteString("inner join ")
		// innerDB not support ``
		if tb := DBS[innerTables[i]]; tb != "" {
			bufNt.WriteString("`" + tb + "`.")
		}
		bufNt.WriteString("`")
		bufNt.WriteString(innerTables[i])
		bufNt.WriteString("` on `")
		bufNt.WriteString(innerTables[i-1])
		bufNt.WriteString("`.`")
		bufNt.WriteString(innerField[i-1])
		bufNt.WriteString("`=`")
		bufNt.WriteString(innerTables[i])
		bufNt.WriteString("`.`")
		bufNt.WriteString(innerField[i])
		bufNt.WriteString("` ")
	}
	// left join
	for i := 1; i < len(leftTables); i += 2 {
		bufNt.WriteString("left join ")
		if tb := DBS[leftTables[i]]; tb != "" {
			bufNt.WriteString("`" + tb + "`.")
		}
		bufNt.WriteString("`")
		bufNt.WriteString(leftTables[i])
		bufNt.WriteString("` on `")
		bufNt.WriteString(leftTables[i-1])
		bufNt.WriteString("`.`")
		bufNt.WriteString(leftField[i-1])
		bufNt.WriteString("`=`")
		bufNt.WriteString(leftTables[i])
		bufNt.WriteString("`.`")
		bufNt.WriteString(leftField[i])
		bufNt.WriteString("` ")
	}
	sqlNt = bufNt.String()
	coMap.Add(keyNt, sqlNt)
	coMap.Add(keyTs, strings.Join(tables, ","))
	return
}

// more sql tables
// can read by ram
func moreTables(gt *GT) (innerTables, leftTables, innerField, leftField []string, DBS map[string]string) {

	for k, v := range gt.InnerTable {
		st := strings.Split(v, ":")

		if strings.Contains(st[0], ".") {
			sts := strings.Split(st[0], ".")
			if DBS == nil {
				DBS = make(map[string]string)
			}
			DBS[sts[1]] = sts[0]
			st[0] = sts[1]
		}
		innerTables = append(innerTables, st[0])
		if len(st) == 1 { // default
			field := "id"
			if k%2 == 0 {
				// default other table_id
				otb := strings.Split(gt.InnerTable[k+1], ":")[0]
				if strings.Contains(otb, ".") {
					otb = strings.Split(otb, ".")[1]
				}
				field = otb + "_id"
			}
			innerField = append(innerField, field)
		} else {
			innerField = append(innerField, st[1])
		}
	}
	// left
	for k, v := range gt.LeftTable {
		st := strings.Split(v, ":")

		if strings.Contains(st[0], ".") {
			sts := strings.Split(st[0], ".")
			if DBS == nil {
				DBS = make(map[string]string)
			}
			DBS[sts[1]] = sts[0]
			st[0] = sts[1]
		}
		leftTables = append(leftTables, st[0])
		if len(st) == 1 {
			field := "id"
			if k%2 == 0 {
				// default other table_id
				otb := strings.Split(gt.LeftTable[k+1], ":")[0]
				if strings.Contains(otb, ".") {
					otb = strings.Split(otb, ".")[1]
				}
				field = otb + "_id"
			}
			leftField = append(leftField, field)
		} else {
			leftField = append(leftField, st[1])
		}
	}
	return
}

// 分页参数不传, 查询所有
// 默认根据id倒序
// 单张表
func GetSearchSQL(gt *GT) (sqlNt, sql string, clientPage, everyPage int64, args []interface{}) {

	var (
		order        = "id desc"  // default order by
		key          = ""         // key like binary search
		bufW, bufNtW bytes.Buffer // where sql, sqlNt bytes sql
		table        = sq.Table(gt.Table)
	)

	// select* replace
	sql = fmt.Sprintf("select %s%s from %s ", GetColSQL(gt.Model), gt.SubSQL, table)
	sqlNt = fmt.Sprintf("select count(id) as total_num from %s ", table)
	for k, v := range gt.CMaps {
		if v[0] == "" {
			continue
		}
		switch k {
		case str.GtClientPage, str.GtClientPageUnderLine:
			clientPage, _ = strconv.ParseInt(v[0], 10, 64)
			continue
		case str.GtEveryPage, str.GtEveryPageUnderLine:
			everyPage, _ = strconv.ParseInt(v[0], 10, 64)
			continue
		case str.GtOrder:
			order = v[0]
			continue
		case str.GtKey:
			key = v[0]
			if gt.KeyModel == nil {
				gt.KeyModel = gt.Model
			}
			sqlKey, argsKey := sq.GetKeySQL(key, gt.KeyModel, table)
			bufW.WriteString(sqlKey)
			bufNtW.WriteString(sqlKey)
			args = append(args, argsKey[:]...)
			continue
		case str.GtMock:
			mock.Mock(gt.Data)
			gt.isMock = true
			continue
		case "":
			continue
		}
		bufW.WriteString(k)
		bufW.WriteString(" = ? and ")
		bufNtW.WriteString(k)
		bufNtW.WriteString(" = ? and ")
		args = append(args, v[0]) // args
	}

	if bufW.Len() != 0 {
		sql += fmt.Sprintf("where %s ", bufW.Bytes()[:bufW.Len()-4])
		sqlNt += fmt.Sprintf("where %s", bufNtW.Bytes()[:bufNtW.Len()-4])
		if gt.SubWhereSQL != "" {
			sql += fmt.Sprintf("and %s ", gt.SubWhereSQL)
			sqlNt += fmt.Sprintf("and %s", gt.SubWhereSQL)
		}
	} else if gt.SubWhereSQL != "" {
		sql += fmt.Sprintf(" where %s ", gt.SubWhereSQL)
		sqlNt += fmt.Sprintf(" where %s", gt.SubWhereSQL)
	}
	sql += fmt.Sprintf(" order by %s ", order)
	return
}

// get data sql
func GetDataSQL(gt *GT) (sql string, args []interface{}) {

	var (
		order = ""         // default no order by
		key   = ""         // key like binary search
		bufW  bytes.Buffer // where sql, sqlNt bytes sql
		table = sq.Table(gt.Table)
	)

	// select* replace
	sql = fmt.Sprintf("select %s%s from %s ", GetColSQL(gt.Model), gt.SubSQL, table)
	for k, v := range gt.CMaps {
		if v[0] == "" {
			continue
		}
		switch k {
		case str.GtOrder:
			order = v[0]
			continue
		case str.GtKey:
			key = v[0]
			if gt.KeyModel == nil {
				gt.KeyModel = gt.Model
			}
			sqlKey, argsKey := sq.GetKeySQL(key, gt.KeyModel, table)
			bufW.WriteString(sqlKey)
			args = append(args, argsKey[:]...)
			continue
		case str.GtMock:
			mock.Mock(gt.Data)
			gt.isMock = true
			continue
		case "":
			continue
		}
		bufW.WriteString(k)
		bufW.WriteString(" = ? and ")
		args = append(args, v[0]) // args
	}

	if bufW.Len() != 0 {
		sql += fmt.Sprintf(" where %s ", bufW.Bytes()[:bufW.Len()-4])
		if gt.SubWhereSQL != "" {
			sql += fmt.Sprintf("and %s ", gt.SubWhereSQL)
		}
	} else if gt.SubWhereSQL != "" {
		sql += fmt.Sprintf(" where %s ", gt.SubWhereSQL)
	}
	if order != "" {
		sql += fmt.Sprintf(" order by %s ", order)
	}
	return
}

// select sql
func GetSelectSearchSQL(gt *GT) (sqlNt, sql string, clientPage, everyPage int64) {

	for k, v := range gt.CMaps {
		switch k {
		case str.GtClientPage, str.GtClientPageUnderLine:
			clientPage, _ = strconv.ParseInt(v[0], 10, 64)
			continue
		case str.GtEveryPage, str.GtEveryPageUnderLine:
			everyPage, _ = strconv.ParseInt(v[0], 10, 64)
			continue
		}
	}

	sql = gt.Select
	if gt.From == "" {
		gt.From = "from"
	}
	sqlNt = "select count(*) as total_num " + gt.From + strings.Join(strings.Split(sql, gt.From)[1:], "")
	if gt.Group != "" {
		sql += gt.Group
	}
	return
}

// 传入数据库表名
// 更新语句拼接
func GetUpdateSQL(table string, params cmap.CMap) (sql string, args []interface{}) {

	// sql connect
	var (
		id  string       // id
		buf bytes.Buffer // sql bytes connect
	)
	buf.WriteString("update `")
	buf.WriteString(table)
	buf.WriteString("` set ")
	for k, v := range params {
		if k == "id" {
			id = v[0]
			continue
		}
		buf.WriteString("`")
		buf.WriteString(k)
		buf.WriteString("` = ?,")
		args = append(args, v[0])
	}
	args = append(args, id)
	sql = string(buf.Bytes()[:buf.Len()-1]) + " where id = ?"
	return sql, args
}

// 传入数据库表名
// 插入语句拼接
func GetInsertSQL(table string, params cmap.CMap) (sql string, args []interface{}) {

	// sql connect
	var (
		sqlv string
		buf  bytes.Buffer // sql bytes connect
	)
	buf.WriteString("insert `")
	buf.WriteString(table)
	buf.WriteString("`(")
	//sql = "insert `" + table + "`("

	for k, v := range params {
		buf.WriteString("`")
		buf.WriteString(k)
		buf.WriteString("`,")

		args = append(args, v[0])
		sqlv += "?,"
	}
	//sql = buf.Bytes()[:buf.Len()-1]
	sql = buf.String()
	sql = string([]byte(sql)[:len(sql)-1]) + ") value(" + sqlv
	sql = string([]byte(sql)[:len(sql)-1]) + ")" // remove ','

	return sql, args
}

// ===================================================================================
// ==========================common crud=========made=by=lucheng======================
// ===================================================================================

// get
// relation get
////////////////

// 获得数据,根据sql语句,无分页
func (db *DBTool) GetDataBySQL(data interface{}, sql string, args ...interface{}) {

	db.DB.Raw(sql, args[:]...).Scan(data)
}

// 获得数据,根据name条件
func (db *DBTool) GetDataByName(data interface{}, name, value string) (err error) {

	dba := db.DB.Where(name+" = ?", value).Find(data) //只要一行数据时使用 LIMIT 1,增加查询性能

	//有数据是返回相应信息
	if dba.Error != nil {
		err = sq.GetSQLError(dba.Error.Error())
	} else {
		// get data
		err = nil
	}
	return
}

//// inner join
//// 查询数据约定,表名_字段名(若有重复)
//// 获得数据,根据id,两张表连接尝试
//func (db *DBTool) GetDoubleTableDataByID(model, data interface{}, id, table1, table2 string) error {
//	sql := fmt.Sprintf("select %s from `%s` inner join `%s` "+
//		"on `%s`.%s_id=`%s`.id where `%s`.id=? limit 1", GetDoubleTableColumnSQL(model, table1, table2), table1, table2, table1, table2, table2, table1)
//
//	return db.GetDataBySQL(data, sql, id)
//}
//
//// left join
//// 查询数据约定,表名_字段名(若有重复)
//// 获得数据,根据id,两张表连接
//func (db *DBTool) GetLeftDoubleTableDataByID(model, data interface{}, id, table1, table2 string) error {
//
//	sql := fmt.Sprintf("select %s from `%s` left join `%s` on `%s`.%s_id=`%s`.id where `%s`.id=? limit 1", GetDoubleTableColumnSQL(model, table1, table2), table1, table2, table1, table2, table2, table1)
//
//	return db.GetDataBySQL(data, sql, id)
//}

// 获得数据,根据id
func (db *DBTool) GetDataByID(gt *GT, id interface{}) {

	db.GetDataBySQL(gt.Data, fmt.Sprintf("select %s from %s where id = ?", GetColSQL(gt.Model), sq.Table(gt.Table)), id)
}

// More Table
// params: innerTables is inner join tables
// params: leftTables is left join tables
// return: search info
// table1 as main table, include other tables_id(foreign key)
func (db *DBTool) GetMoreDataBySearch(gt *GT) (pager result.Pager) {
	// more table search
	sqlNt, sql, clientPage, everyPage, args := GetMoreSearchSQL(gt)
	// isMock
	if gt.isMock {
		return
	}
	return db.GetDataBySQLSearch(gt.Data, sql, sqlNt, clientPage, everyPage, args, args)
}

// 获得数据,分页/查询
func (db *DBTool) GetDataBySearch(gt *GT) (pager result.Pager) {

	sqlNt, sql, clientPage, everyPage, args := GetSearchSQL(gt)
	// isMock
	if gt.isMock {
		return
	}
	return db.GetDataBySQLSearch(gt.Data, sql, sqlNt, clientPage, everyPage, args, args)
}

// 获得数据, no search
func (db *DBTool) GetData(gt *GT) {

	sql, args := GetDataSQL(gt)
	// isMock
	if gt.isMock {
		return
	}
	db.GetDataBySQL(gt.Data, sql, args[:]...)
}

// select sql search
func (db *DBTool) GetDataBySelectSQLSearch(gt *GT) (pager result.Pager) {

	sqlNt, sql, clientPage, everyPage := GetSelectSearchSQL(gt)

	return db.GetDataBySQLSearch(gt.Data, sql, sqlNt, clientPage, everyPage, gt.Args, gt.ArgsNt)
}

// 获得数据,根据sql语句,分页
// args : sql参数'？'
// sql, sqlNt args 相同, 共用args
func (db *DBTool) GetDataBySQLSearch(data interface{}, sql, sqlNt string, clientPage, everyPage int64, args []interface{}, argsNt []interface{}) (pager result.Pager) {

	// if no clientPage or everyPage
	// return all data
	if clientPage != 0 && everyPage != 0 {
		sql += fmt.Sprintf("limit %d, %d", (clientPage-1)*everyPage, everyPage)
	}
	// sqlNt += limit
	db.DB.Raw(sqlNt, argsNt[:]...).Scan(&pager)
	if db.Error == nil {
		db.DB.Raw(sql, args[:]...).Scan(data)
		// pager data
		pager.ClientPage = clientPage
		pager.EveryPage = everyPage
		return
	}
	return
}

// exec common
////////////////////

// exec sql
func (db *DBTool) ExecSQL(sql string, args ...interface{}) {

	db.Exec(sql, args...)
}

// delete
///////////////////

// delete
func (db *DBTool) Delete(table string, id interface{}) {
	switch id.(type) {
	case string:
		if strings.Contains(id.(string), ",") {
			id = strings.Split(id.(string), ",")
		}
	}
	db.ExecSQL(fmt.Sprintf("delete from %s where id in (?)", sq.Table(table)), id)
}

// update
///////////////////

// via form data update
func (db *DBTool) UpdateFormData(table string, params cmap.CMap) (err error) {

	sql, args := GetUpdateSQL(table, params)
	db.ExecSQL(sql, args...)
	return db.Error
}

// 结合struct修改
func (db *DBTool) UpdateStructData(data interface{}) (err error) {
	var num int64

	dba := db.DB.Save(data)
	num = dba.RowsAffected
	switch {
	case dba.Error != nil:
		err = sq.GetSQLError(dba.Error.Error())
	case num == 0 && dba.Error == nil:
		err = &te.TextError{Msg: result.MsgExistOrNo}
	default:
		err = nil
	}
	return
}

// create
////////////////////

// via form data create
func (db *DBTool) CreateFormData(table string, params cmap.CMap) error {

	sql, args := GetInsertSQL(table, params)
	db.ExecSQL(sql, args...)
	return db.Error
}

// param.CMap 形式批量创建问题: 顺序对应, 使用json形式批量创建

// 创建数据,通用
// 返回id,事务,慎用
// 业务少可用
func (db *DBTool) CreateDataResID(table string, params cmap.CMap) (id str.ID, err error) {

	//开启事务
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	sql, args := GetInsertSQL(table, params)
	dba := tx.Exec(sql, args[:]...)

	tx.Raw("select max(id) as id from `%s`", table).Scan(&id)

	switch {
	case dba.Error != nil:
		err = sq.GetSQLError(dba.Error.Error())
	default:
		err = nil
	}

	if tx.Error != nil {
		tx.Rollback()
	}

	tx.Commit()
	return
}

// select检查是否存在
// == nil 即存在
func (db *DBTool) ValidateSQL(sql string) (err error) {
	var num int64 //返回影响的行数

	var ve str.Value
	dba := db.DB.Raw(sql).Scan(&ve)
	num = dba.RowsAffected
	switch {
	case dba.Error != nil:
		err = sq.GetSQLError(dba.Error.Error())
	case num == 0 && dba.Error == nil:
		err = &te.TextError{Msg: result.MsgExistOrNo}
	default:
		err = nil
	}
	return
}

//==============================================================
// json处理(struct data)
//==============================================================

// create
func (db *DBTool) CreateData(table string, data interface{}) {

	db.Table(table).Create(data)
}

// data must array type
// more data create
// single table
func (db *DBTool) CreateMoreData(table string, model interface{}, data interface{}) {
	var (
		buf    bytes.Buffer
		params []interface{}
	)

	// array data
	arrayData := reflect2.ToSlice(data)

	for _, v := range arrayData {
		// buf
		buf.WriteString("(")
		buf.WriteString(GetColParamSQL(model))
		buf.WriteString("),")
		// params
		params = append(params, GetParams(v)[:]...)
	}
	values := string(buf.Bytes()[:buf.Len()-1])

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", sq.Table(table), GetColSQL(model), values)
	db.DB.Exec(sql, params[:]...)
}

// update
func (db *DBTool) UpdateData(gt *GT) {

	if gt.Model == nil {
		gt.Model = gt.Data
	}

	if gt.Select != "" {
		db.Table(gt.Table).Model(gt.Model).Where(gt.Select, gt.Args).Update(gt.Data)
	} else {
		db.Table(gt.Table).Model(gt.Data).Update(gt.Data)
	}

	//db.Update(gt.Data)
}
