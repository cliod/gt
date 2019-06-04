// @author  dreamlu
package der

import (
	"bytes"
	"fmt"
	"github.com/dreamlu/go-tool/util/lib"
	rf "github.com/dreamlu/go-tool/util/reflect"
	"reflect"
	"strconv"
	"strings"
)

//======================return tag=============================
//=============================================================

// select * replace
// select more tables
// tables : table name / table alias name
// 主表放在tables中第一个，紧接着为主表关联的外键表名(无顺序)
func GetMoreTableColumnSQL(model interface{}, tables ...string) (sql string) {
	typ := reflect.TypeOf(model)

	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		// foreign tables column
		for _, v := range tables {
			// tables
			switch {
			case !strings.Contains(tag, v+"_id") && strings.Contains(tag, v+"_"):
				sql += "`" + v + "`.`" + string([]byte(tag)[len(v)+1:]) + "` as " + tag + ","
				goto into
			}
		}
		sql += "`" + tables[0] + "`.`" + tag + "`,"
	into:
	}
	sql = string([]byte(sql)[:len(sql)-1]) //去点,
	return sql
}

// select *替换
// 两张表
func GetDoubleTableColumnSQL(model interface{}, table1, table2 string) (sql string) {
	typ := reflect.TypeOf(model)
	//var buffer bytes.Buffer
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		//table2的数据处理,去除table2_id
		if strings.Contains(tag, table2+"_") && !strings.Contains(tag, table2+"_id") {
			sql += table2 + ".`" + string([]byte(tag)[len(table2)+1:]) + "` as " + tag + "," //string([]byte(tag)[len(table2+1-1):])
			continue
		}
		sql += table1 + ".`" + tag + "`,"
	}
	sql = string([]byte(sql)[:len(sql)-1]) //去掉点,
	return sql
}

// 根据model中表模型的json标签获取表字段
// 将select* 中'*'变为对应的字段名
// 增加别名,表连接问题
func GetColAliasSQL(model interface{}, alias string) (sql string) {
	typ := reflect.TypeOf(model)
	// var buffer bytes.Buffer
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		// buffer.WriteString("`"+tag + "`,")
		sql += alias + ".`" + tag + "`,"
	}
	sql = string([]byte(sql)[:len(sql)-1]) //去掉点,
	return sql
}

// 根据model中表模型的json标签获取表字段
// 将select* 变为对应的字段名
func GetColSql(model interface{}) (sql string) {
	typ := reflect.TypeOf(model)
	//var buffer bytes.Buffer
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		//buffer.WriteString("`"+tag + "`,")
		sql += "`" + tag + "`,"
	}
	sql = string([]byte(sql)[:len(sql)-1]) //去掉点,
	return sql
}

//=======================================sql语句处理==========================================
//===========================================================================================

// More Table
// params: innerTables is inner join tables
// params: leftTables is left join tables
// return: select sql
// table1 as main table, include other tables_id(foreign key)
func GetMoreSearchSQL(model interface{}, params map[string][]string, innerTables []string, leftTables []string) (sqlnolimit, sql string, clientPage, everyPage int64) {

	// default pager
	//clientPage, _ = strconv.ParseInt(GetDevModeConfig("clientPage"), 10, 64)
	//everyPage, _ = strconv.ParseInt(GetDevModeConfig("everyPage"), 10, 64)

	var (
		clientPageStr = GetDevModeConfig("clientPage") // page number
		everyPageStr  = GetDevModeConfig("everyPage")  // Number of pages per page
		every         = ""                             // if every != "", it will return all data
		key           = ""                             // key like binary search
		tables        = innerTables                    // all tables
		buf           bytes.Buffer                     // sql bytes connect
	)

	tables = append(tables, leftTables...)
	// sql and sqlCount
	buf.WriteString("select ")
	buf.WriteString("count(`")
	buf.WriteString(tables[0])
	buf.WriteString("`.id) as sum_page ")
	//buf.WriteString(GetMoreTableColumnSQL(model, tables[:]...))
	buf.WriteString("from `")
	buf.WriteString(tables[0])
	buf.WriteString("`")
	// inner join
	for i := 1; i < len(innerTables); i++ {
		buf.WriteString(" inner join `")
		buf.WriteString(innerTables[i])
		buf.WriteString("` on `")
		buf.WriteString(tables[0])
		buf.WriteString("`.")
		buf.WriteString(innerTables[i])
		buf.WriteString("_id=`")
		buf.WriteString(innerTables[i])
		buf.WriteString("`.id ")
		//sql += " inner join ·" + innerTables[i] + "`"
	}
	// left join
	for i := 0; i < len(leftTables); i++ {
		buf.WriteString(" left join `")
		buf.WriteString(leftTables[i])
		buf.WriteString("` on `")
		buf.WriteString(tables[0])
		buf.WriteString("`.")
		buf.WriteString(leftTables[i])
		buf.WriteString("_id=`")
		buf.WriteString(leftTables[i])
		buf.WriteString("`.id ")
		//sql += " inner join ·" + innerTables[i] + "`"
	}
	buf.WriteString(" where 1=1 and ")

	//select* 变为对应的字段名
	sqlnolimit = buf.String()
	sql = strings.Replace(sqlnolimit, "count(`"+tables[0]+"`.id) as sum_page", GetMoreTableColumnSQL(model, tables[:]...), 1)
	for k, v := range params {
		switch k {
		case "clientPage":
			clientPageStr = v[0]
			continue
		case "everyPage":
			everyPageStr = v[0]
			continue
		case "every":
			every = v[0]
			continue
		case "key":
			key = v[0]
			var tablens = append(tables, tables[:]...)
			for k, v := range tablens {
				tablens[k] += ":" + v
			}
			// more tables key search
			sql, sqlnolimit = lib.GetMoreKeySQL(sql, sqlnolimit, key, model, tablens[:]...)
			continue
		case "":
			continue
		}

		// other tables, except tables[0]
		for _, table := range tables[1:] {
			switch {
			case !strings.Contains(table, table+"_id") && strings.Contains(table, table+"_"):
				v[0] = strings.Replace(v[0], "'", "\\'", -1)
				sql += "`" + table + "`.`" + string([]byte(k)[len(v)+1:]) + "`" + " = '" + v[0] + "' and "
				sqlnolimit += "`" + table + "`.`" + string([]byte(k)[len(v)+1:]) + "`" + " = '" + v[0] + "' and "
				goto into
			}
		}
		v[0] = strings.Replace(v[0], "'", "\\'", -1)
		sql += "`" + tables[0] + "`." + k + " = '" + v[0] + "' and "
		sqlnolimit += "`" + tables[0] + "`." + k + " = '" + v[0] + "' and "
	into:
	}

	clientPage, _ = strconv.ParseInt(clientPageStr, 10, 64)
	everyPage, _ = strconv.ParseInt(everyPageStr, 10, 64)

	sql = string([]byte(sql)[:len(sql)-4])                      //去and
	sqlnolimit = string([]byte(sqlnolimit)[:len(sqlnolimit)-4]) //去and
	if every == "" {
		sql += "order by `" + tables[0] + "`.id desc limit " + strconv.FormatInt((clientPage-1)*everyPage, 10) + "," + everyPageStr
	}

	return sqlnolimit, sql, clientPage, everyPage
}

// 两张表名,查询语句拼接
// 表1中有表2 id
func GetDoubleSearchSql(model interface{}, table1, table2 string, params map[string][]string) (sqlnolimit, sql string, clientPage, everyPage int64) {

	var (
		clientPageStr = GetDevModeConfig("clientPage") // page number
		everyPageStr  = GetDevModeConfig("everyPage")  // Number of pages per page
		every         = ""                             // if every != "", it will return all data
		key           = ""                             // key like binary search
	)

	//select* 变为对应的字段名
	sql = fmt.Sprintf("select %s from `%s` inner join `%s` on `%s`.%s_id=%s.id where 1=1 and ", GetDoubleTableColumnSQL(model, table1, table2), table1, table2, table1, table2, table2)

	sqlnolimit = fmt.Sprintf("select count(%s.id) as sum_page from `%s` inner join `%s` on `%s`.%s_id=%s.id where 1=1 and ", table1, table1, table2, table1, table2, table2)
	for k, v := range params {
		switch k {
		case "clientPage":
			clientPageStr = v[0]
			continue
		case "everyPage":
			everyPageStr = v[0]
			continue
		case "every":
			every = v[0]
			continue
		case "key":
			key = v[0]
			// 多表搜索
			sql, sqlnolimit = lib.GetMoreKeySQL(sql, sqlnolimit, key, model, table1+":"+table1, table2+":"+table2)
			//sql, sqlnolimit = lib.GetKeySql(sql, sqlnolimit, key, model , table2)
			continue
		case "":
			continue
		}

		//表2值查询
		if strings.Contains(k, table2+"_") && !strings.Contains(k, table2+"_id") {
			sql += table2 + ".`" + string([]byte(k)[len(table2)+1:]) + "`" + " = '" + v[0] + "' and " //string([]byte(tag)[len(table2+1-1):])
			sqlnolimit += table2 + ".`" + string([]byte(k)[len(table2)+1:]) + "`" + " = '" + v[0] + "' and "
			continue
		}

		v[0] = strings.Replace(v[0], "'", "\\'", -1) //转义
		sql += table1 + "." + k + " = '" + v[0] + "' and "
		sqlnolimit += table1 + "." + k + " = '" + v[0] + "' and "
	}

	clientPage, _ = strconv.ParseInt(clientPageStr, 10, 64)
	everyPage, _ = strconv.ParseInt(everyPageStr, 10, 64)

	sql = string([]byte(sql)[:len(sql)-4])                      //去and
	sqlnolimit = string([]byte(sqlnolimit)[:len(sqlnolimit)-4]) //去and
	if every == "" {
		sql += "order by " + table1 + ".id desc limit " + strconv.FormatInt((clientPage-1)*everyPage, 10) + "," + everyPageStr
	}

	return sqlnolimit, sql, clientPage, everyPage
}

// 传入表名,查询语句拼接
// 单张表
func GetSearchSQL(model interface{}, table string, params map[string][]string) (sqlnolimit, sql string, clientPage, everyPage int64, args []interface{}) {

	var (
		clientPageStr = GetDevModeConfig("clientPage") // page number
		everyPageStr  = GetDevModeConfig("everyPage")  // Number of pages per page
		every         = ""                             // if every != "", it will return all data
		key           = ""                             // key like binary search
	)

	//select* replace
	sql = fmt.Sprintf("select %s from `%s` where 1=1 and ", GetColSql(model), table)
	sqlnolimit = fmt.Sprintf("select count(id) as sum_page from `%s` where 1=1 and ", table)
	for k, v := range params {
		switch k {
		case "clientPage":
			clientPageStr = v[0]
			continue
		case "everyPage":
			everyPageStr = v[0]
			continue
		case "every":
			every = v[0]
			continue
		case "key":
			key = v[0]
			sql, sqlnolimit = lib.GetKeySQL(sql, sqlnolimit, key, model, table)
			continue
		case "":
			continue
		}

		//v[0] = strings.Replace(v[0], "'", "\\'", -1) //转义
		sql += k + " = ? and "
		sqlnolimit += k + " = ? and "
		args = append(args, v[0]) // args
	}

	clientPage, _ = strconv.ParseInt(clientPageStr, 10, 64)
	everyPage, _ = strconv.ParseInt(everyPageStr, 10, 64)

	sql = string([]byte(sql)[:len(sql)-4])                      //去and
	sqlnolimit = string([]byte(sqlnolimit)[:len(sqlnolimit)-4]) //去and
	if every == "" {
		sql += "order by id desc limit " + strconv.FormatInt((clientPage-1)*everyPage, 10) + "," + everyPageStr
	}

	return sqlnolimit, sql, clientPage, everyPage, args
}

// 传入数据库表名
// 更新语句拼接
func GetUpdateSQL(table string, params map[string][]string) (sql string, args []interface{}) {

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
func GetInsertSQL(table string, params map[string][]string) (sql string, args []interface{}) {

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
func GetDataBySQL(data interface{}, sql string, args ...interface{}) lib.GetInfo {
	var info lib.GetInfo

	dba := DB.Raw(sql, args[:]...).Scan(data)
	num := dba.RowsAffected

	//有数据是返回相应信息
	if dba.Error != nil {
		info = lib.GetInfoData(nil, lib.GetSqlError(dba.Error.Error()))
	} else if num == 0 && dba.Error == nil {
		info = lib.GetInfoData(nil, lib.MapNoResult)
	} else {
		// get data
		info = lib.GetMapDataSuccess(data)
	}
	return info
}

// 获得数据,根据name条件
func GetDataByName(data interface{}, name, value string) lib.GetInfo {
	var info lib.GetInfo

	dba := DB.Where(name+" = ?", value).Find(data) //只要一行数据时使用 LIMIT 1,增加查询性能
	num := dba.RowsAffected

	//有数据是返回相应信息
	if dba.Error != nil {
		info = lib.GetInfoData(nil, lib.GetSqlError(dba.Error.Error()))
	} else if num == 0 && dba.Error == nil {
		info = lib.GetInfoData(nil, lib.MapNoResult)
	} else {
		// get data
		info = lib.GetMapDataSuccess(data)
	}
	return info
}

// inner join
// 查询数据约定,表名_字段名(若有重复)
// 获得数据,根据id,两张表连接尝试
func GetDoubleTableDataByID(model, data interface{}, id, table1, table2 string) lib.GetInfo {
	sql := fmt.Sprintf("select %s from `%s` inner join `%s` "+
		"on `%s`.%s_id=`%s`.id where `%s`.id=? limit 1", GetDoubleTableColumnSQL(model, table1, table2), table1, table2, table1, table2, table2, table1)

	return GetDataBySQL(data, sql, id)
}

// left join
// 查询数据约定,表名_字段名(若有重复)
// 获得数据,根据id,两张表连接
func GetLeftDoubleTableDataByID(model, data interface{}, id, table1, table2 string) lib.GetInfo {

	sql := fmt.Sprintf("select %s from `%s` left join `%s` on `%s`.%s_id=`%s`.id where `%s`.id=? limit 1", GetDoubleTableColumnSQL(model, table1, table2), table1, table2, table1, table2, table2, table1)

	return GetDataBySQL(data, sql, id)
}

// 获得数据,根据id
func GetDataByID(data interface{}, id string) lib.GetInfo {
	var info lib.GetInfo

	dba := DB.First(data, id) //只要一行数据时使用 LIMIT 1,增加查询性能
	num := dba.RowsAffected

	//有数据是返回相应信息
	if dba.Error != nil {
		info = lib.GetInfoData(nil, lib.GetSqlError(dba.Error.Error()))
	} else if num == 0 && dba.Error == nil {
		info = lib.GetInfoData(nil, lib.MapNoResult)
	} else {
		// get data
		info = lib.GetMapDataSuccess(data)
	}
	return info
}

// More Table
// params: innerTables is inner join tables
// params: leftTables is left join tables
// return: search info
// table1 as main table, include other tables_id(foreign key)
func GetMoreDataBySearch(model, data interface{}, params map[string][]string, innerTables []string, leftTables []string, args ...interface{}) lib.GetInfoPager {
	// more table search
	sqlnolimit, sql, clientPage, everyPage := GetMoreSearchSQL(model, params, innerTables, leftTables)

	return GetDataBySQLSearch(data, sql, sqlnolimit, clientPage, everyPage, args[:]...)
}

// 获得数据,分页/查询,遵循一定查询规则,两张表,使用left join
// 如table2中查询,字段用table2_+"字段名",table1字段查询不变
func GetLeftDoubleTableDataBySearch(model, data interface{}, table1, table2 string, params map[string][]string) lib.GetInfoPager {
	//级联表的查询
	sqlnolimit, sql, clientPage, everyPage := GetDoubleSearchSql(model, table1, table2, params)
	sql = strings.Replace(sql, "inner join", "left join", 1)
	sqlnolimit = strings.Replace(sqlnolimit, "inner join", "left join", 1)

	return GetDataBySQLSearch(data, sql, sqlnolimit, clientPage, everyPage)
}

// 获得数据,分页/查询,遵循一定查询规则,两张表,默认inner join
// 如table2中查询,字段用table2_+"字段名",table1字段查询不变
func GetDoubleTableDataBySearch(model, data interface{}, table1, table2 string, params map[string][]string) lib.GetInfoPager {
	//级联表的查询以及
	sqlnolimit, sql, clientPage, everyPage := GetDoubleSearchSql(model, table1, table2, params)

	return GetDataBySQLSearch(data, sql, sqlnolimit, clientPage, everyPage)
}

// 获得数据,根据sql语句,分页
// args : sql参数'？'
// sql, sqlnolimit args 相同, 共用args
func GetDataBySQLSearch(data interface{}, sql, sqlnolimit string, clientPage, everyPage int64, args ...interface{}) lib.GetInfoPager {
	var info lib.GetInfoPager

	//// 完善可变长参数赋值问题
	//var value []interface{}
	//value = append(value, args[:]...)

	dba := DB.Raw(sqlnolimit, args[:]...).Scan(&info.Pager)
	num := info.Pager.SumPage
	if dba.Error != nil {

		info = lib.GetInfoPagerDataNil(nil, lib.GetSqlError(dba.Error.Error()))
	} else if num == 0 && dba.Error == nil {

		info = lib.GetInfoPagerDataNil(nil, lib.MapNoResult)
	} else {
		// DB.Debug().Find(&dest)
		dba = DB.Raw(sql, args[:]...).Scan(data)

		if dba.Error != nil {
			info = lib.GetInfoPagerDataNil(nil, lib.GetSqlError(dba.Error.Error()))
			return info
		}
		// get data
		info = lib.GetMapDataPager(data, clientPage, everyPage, num)

		// 去除pager
		//switch {
		//case strings.Contains(sql, "limit"): // maintain pager
		//	info = lib.GetMapDataPager(data, clientPage, everyPage, num)
		//default: // remove pager data
		//	info = lib.GetInfoPagerDataNil(data, lib.MapSuccess)
		//}
	}
	return info
}

// 获得数据,分页/查询
func GetDataBySearch(model, data interface{}, table string, params map[string][]string) lib.GetInfoPager {

	sqlnolimit, sql, clientPage, everyPage, args := GetSearchSQL(model, table, params)

	return GetDataBySQLSearch(data, sql, sqlnolimit, clientPage, everyPage, args[:]...)
}

// delete
///////////////////

// delete by sql
func DeleteDataBySQL(sql string, args ...interface{}) (info lib.MapData) {
	var num int64 //返回影响的行数

	dba := DB.Exec(sql, args[:]...)
	num = dba.RowsAffected
	switch {
	case dba.Error != nil:
		info = lib.GetSqlError(dba.Error.Error())
	case num == 0 && dba.Error == nil:
		info = lib.MapExistOrNo
	default:
		info = lib.MapDelete
	}
	return info
}

// 删除通用,任意参数
func DeleteDataByName(table string, key, value string) (info lib.MapData) {
	sql := fmt.Sprintf("delete from `%s` where %s=?", table, key)

	return DeleteDataBySQL(sql, value)
}

// update
///////////////////

// 修改数据,通用
func UpdateDataBySQL(sql string, args ...interface{}) (info lib.MapData) {
	var num int64 //返回影响的行数

	dba := DB.Exec(sql, args[:]...)
	num = dba.RowsAffected
	switch {
	case dba.Error != nil:
		info = lib.GetSqlError(dba.Error.Error())
	case num == 0 && dba.Error == nil:
		info = lib.MapExistOrNo
	default:
		info = lib.MapUpdate
	}
	return info
}

// 修改数据,通用
func UpdateData(table string, params map[string][]string) (info lib.MapData) {

	sql, args := GetUpdateSQL(table, params)

	return UpdateDataBySQL(sql, args[:]...)
}

// 结合struct修改
func UpdateStructData(data interface{}) (info lib.MapData) {
	var num int64 //返回影响的行数

	dba := DB.Save(data)
	num = dba.RowsAffected
	switch {
	case dba.Error != nil:
		info = lib.GetSqlError(dba.Error.Error())
	case num == 0 && dba.Error == nil:
		info = lib.MapExistOrNo
	default:
		info = lib.MapUpdate
	}
	return info
}

// create
////////////////////

// Create data by sql
func CreateDataBySQL(sql string, args ...interface{}) lib.MapData {
	var info lib.MapData
	var num int64 //返回影响的行数

	dba := DB.Exec(sql, args[:]...)
	num = dba.RowsAffected
	switch {
	case dba.Error != nil:
		info = lib.GetSqlError(dba.Error.Error())
	case num == 0 && dba.Error == nil:
		info = lib.MapError
	default:
		info = lib.MapCreate
	}
	return info
}

// 创建数据,通用
func CreateData(table string, params map[string][]string) (info lib.MapData) {

	sql, args := GetInsertSQL(table, params)

	return CreateDataBySQL(sql, args[:]...)
}

// 创建数据,通用
// 返回id,事务,慎用
// 业务少可用
func CreateDataResID(table string, params map[string][]string) (info lib.GetInfo) {

	//开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	sql, args := GetInsertSQL(table, params)
	dba := tx.Exec(sql, args[:]...)
	num := dba.RowsAffected

	var data ID
	tx.Raw("select max(id) as id from `%s`", table).Scan(&data)

	switch {
	case dba.Error != nil:
		info = lib.GetInfoData(nil, lib.GetSqlError(dba.Error.Error()))
	case num == 0 && dba.Error == nil:
		info = lib.GetInfoData(nil, lib.MapError)
	default:
		info = lib.GetMapDataSuccess(data)
	}

	if tx.Error != nil {
		tx.Rollback()
	}

	tx.Commit()
	return info
}

// select检查是否存在
func ValidateSQL(sql string) (info lib.MapData) {
	var num int64 //返回影响的行数

	var ve Value
	dba := DB.Raw(sql).Scan(&ve)
	num = dba.RowsAffected
	switch {
	case dba.Error != nil:
		info = lib.GetSqlError(dba.Error.Error())
	case num == 0 && dba.Error == nil:
		info = lib.MapValError
	default:
		info = lib.MapValSuccess
	}
	return info
}

//==============================================================
// json处理(struct data)
//==============================================================

// create
func CreateDataJ(data interface{}) (info lib.MapData) {
	var num int64 //返回影响的行数

	dba := DB.Create(data)
	num = dba.RowsAffected

	switch {
	case dba.Error != nil:
		info = lib.GetSqlError(dba.Error.Error())
	case num == 0 && dba.Error == nil:
		info = lib.MapError
	default:
		info = lib.MapCreate
	}
	return info
}

// more data create
// writing
func CreateMoreData(data interface{}) {

}

// create
// return insert id
func CreateDataJResID(data interface{}) (info lib.GetInfo) {
	var num int64 //返回影响的行数

	dba := DB.Create(data)
	num = dba.RowsAffected

	switch {
	case dba.Error != nil:
		info = lib.GetInfoData(nil, lib.GetSqlError(dba.Error.Error()))
	case num == 0 && dba.Error == nil:
		info = lib.GetInfoData(nil, lib.MapError)
	default:
		id, _ := rf.GetDataByFieldName(data, "ID")
		info = lib.GetInfoData(map[string]interface{}{"id": id}, lib.MapCreate)
	}
	return info
}

// update
func UpdateDataJ(data interface{}) (info lib.MapData) {

	dba := DB.Model(data).Update(data)

	switch {
	case dba.Error != nil:
		info = lib.GetSqlError(dba.Error.Error())
	default:
		info = lib.MapUpdate
	}
	return info
}
