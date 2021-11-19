package util

import (
	"errors"
	"fmt"
	"reflect"

	json "github.com/json-iterator/go"
)

func isPointer(val interface{}) bool {
	if reflect.ValueOf(val).Kind() == reflect.Ptr {
		return true
	}
	return false
}

func MapToStruct(m map[string]interface{}, val interface{}) (ok bool, err error) {
	if ! isPointer(val) {
		return false, nil
	}
	tmp, err := json.Marshal(m)
	if err != nil {
		return false, err
	}
	fmt.Println(string(tmp))
	err = json.Unmarshal(tmp, val)
	if err != nil {
		return false, err
	}
	return true, nil
}

func StructToMap(val interface{}, m *map[string]interface{}) error {
	tmp, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = json.Unmarshal(tmp, m)
	if err != nil {
		return err
	}
	return nil
}

func SetField(v interface{}, name string, value string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("v must be pointer to struct")
	}
	rv = rv.Elem()
	fv := rv.FieldByName(name)
	if !fv.IsValid() {
		return fmt.Errorf("not a field name: %s", name)
	}
	if !fv.CanSet() {
		return fmt.Errorf("cannot set field %s", name)
	}
	if fv.Kind() != reflect.String {
		return fmt.Errorf("%s is not a string field", name)
	}
	fv.SetString(value)
	return nil
}
