// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package internal

import "reflect"

func Copy(dst, src interface{}) {
	aValue := reflect.ValueOf(dst)
	bValue := reflect.ValueOf(src)
	fieldNumsOfb := reflect.Indirect(bValue).NumField()
	for i := 0; i < fieldNumsOfb; i++ {
		bField := reflect.Indirect(bValue).Type().Field(i)
		bFieldValue := reflect.Indirect(bValue).Field(i)
		toFiled := reflect.Indirect(aValue).FieldByName(bField.Name)
		if toFiled.IsValid() {
			toFiled.Set(bFieldValue)
		}
	}
}
