// package der

package der

import (
	"errors"
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// go tool yaml
// use go-yaml
type Yaml struct {
	// yaml data
	data map[interface{}]interface{}
}

// load the default app.yaml data
func (c *Yaml) loadYaml(path string) error {

	yamlS, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return readErr
	}
	// yaml解析的时候c.data如果没有被初始化，会自动为你做初始化
	err := yaml.Unmarshal(yamlS, &c.data)
	if err != nil {
		return errors.New("can not parse " + path + " config")
	}
	return nil
}

// 从配置文件中获取值
func (c *Yaml) Get(name string) interface{} {
	path := strings.Split(name, ".")
	data := c.data
	for key, value := range path {
		v, ok := data[value]
		if !ok {
			break
		}
		if (key + 1) == len(path) {
			return v
		}
		// print yaml v3 problem!
		// use go-yaml v2 replace
		//log.Println(name, "&&",reflect.TypeOf(v).String())
		if reflect.TypeOf(v).String() == "map[interface {}]interface {}" {
			data = v.(map[interface{}]interface{})
		}
	}
	return nil
}

// string
func (c *Yaml) GetString(name string) string {
	value := c.Get(name)
	switch value := value.(type) {
	case string:
		return value
	case bool, float64, int:
		return fmt.Sprint(value)
	default:
		return ""
	}
}

// int
func (c *Yaml) GetInt(name string) int {
	value := c.Get(name)
	switch value := value.(type) {
	case string:
		i, err := strconv.Atoi(value)
		log.Println("[YAML type error]: ", err)
		return i
	case int:
		return value
	case bool:
		if value {
			return 1
		}
		return 0
	case float64:
		return int(value)
	default:
		return 0
	}
}

// bool
func (c *Yaml) GetBool(name string) bool {
	value := c.Get(name)
	switch value := value.(type) {
	case string:
		str, _ := strconv.ParseBool(value)
		return str
	case int:
		if value != 0 {
			return true
		}
		return false
	case bool:
		return value
	case float64:
		if value != 0.0 {
			return true
		}
		return false
	default:
		return false
	}
}

// 从配置文件中获取Struct类型的值,这里的struct是你自己定义的根据配置文件
func (c *Yaml) GetStruct(name string, s interface{}) interface{} {
	d := c.Get(name)
	switch d.(type) {
	case string:
		c.setField(s, name, d)
	case map[interface{}]interface{}:
		c.mapToStruct(d.(map[interface{}]interface{}), s)
	}
	return s
}

func (c *Yaml) mapToStruct(m map[interface{}]interface{}, s interface{}) interface{} {
	for key, value := range m {
		switch key.(type) {
		case string:
			c.setField(s, key.(string), value)
		}
	}
	return s
}

func (c *Yaml) setField(obj interface{}, name string, value interface{}) error {
	// reflect.Indirect 返回value对应的值
	structValue := reflect.Indirect(reflect.ValueOf(obj))
	structFieldValue := structValue.FieldByName(name)

	// isValid 显示的测试一个空指针
	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	// CanSet判断值是否可以被更改
	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	// 获取要更改值的类型
	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)

	if structFieldType.Kind() == reflect.Struct && val.Kind() == reflect.Map {
		vint := val.Interface()

		switch vint.(type) {
		case map[interface{}]interface{}:
			for key, value := range vint.(map[interface{}]interface{}) {
				c.setField(structFieldValue.Addr().Interface(), key.(string), value)
			}
		case map[string]interface{}:
			for key, value := range vint.(map[string]interface{}) {
				c.setField(structFieldValue.Addr().Interface(), key, value)
			}
		}

	} else {
		if structFieldType != val.Type() {
			return errors.New("Provided value type didn't match obj field type")
		}

		structFieldValue.Set(val)
	}

	return nil

}
