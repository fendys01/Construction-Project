package array

import "reflect"

// ArrStr struct
type ArrStr string

// InArray check is in array
func (s ArrStr) InArray(val string, array []string) (exists bool, index int) {
	exists = false
	index = -1

	for i, s := range array {
		if s == val {
			exists = true
			index = i
			return
		}
	}

	return
}

// Remove member array
func (s ArrStr) Remove(array []string, value string) []string {
	isExist, index := s.InArray(value, array)
	if isExist {
		array = append(array[:index], array[(index+1):]...)
	}

	return array
}

// Array Unique filtered
func (s ArrStr) Unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

// ArrUint struct
type ArrUint uint

// InArray check is in array
func (s ArrUint) InArray(val uint, array []uint) (exists bool, index int) {
	exists = false
	index = -1

	for i, s := range array {
		if s == val {
			exists = true
			index = i
			return
		}
	}

	return
}

// Remove member array
func (s ArrUint) Remove(array []uint, value uint) []uint {
	isExist, index := s.InArray(value, array)
	if isExist {
		array = append(array[:index], array[(index+1):]...)
	}

	return array
}

// ArrUint32 struct
type ArrUint32 uint32

// InArray check is in array
func (s ArrUint32) InArray(val uint32, array []uint32) (exists bool, index int) {
	exists = false
	index = -1

	for i, s := range array {
		if s == val {
			exists = true
			index = i
			return
		}
	}

	return
}

// Remove member array
func (s ArrUint32) Remove(array []uint32, value uint32) []uint32 {
	isExist, index := s.InArray(value, array)
	if isExist {
		array = append(array[:index], array[(index+1):]...)
	}

	return array
}

// ArrUint64 struct
type ArrUint64 uint64

// InArray check is in array
func (s ArrUint64) InArray(val uint64, array []uint64) (exists bool, index int) {
	exists = false
	index = -1

	for i, s := range array {
		if s == val {
			exists = true
			index = i
			return
		}
	}

	return
}

// Remove member array
func (s ArrUint64) Remove(array []uint64, value uint64) []uint64 {
	isExist, index := s.InArray(value, array)
	if isExist {
		array = append(array[:index], array[(index+1):]...)
	}

	return array
}

// ArrInt64 struct
type ArrInt64 int64

// InArray check is in array
func (s ArrInt64) InArray(val int64, array []int64) (exists bool, index int) {
	exists = false
	index = -1

	for i, s := range array {
		if s == val {
			exists = true
			index = i
			return
		}
	}

	return
}

// Remove member array
func (s ArrInt64) Remove(array []int64, value int64) []int64 {
	isExist, index := s.InArray(value, array)
	if isExist {
		array = append(array[:index], array[(index+1):]...)
	}

	return array
}

// ArrUint32 struct
type ArrInt32 int32

// InArray check is in array
func (s ArrInt32) InArray(val int32, array []int32) (exists bool, index int) {
	exists = false
	index = -1

	for i, s := range array {
		if s == val {
			exists = true
			index = i
			return
		}
	}

	return
}

// Remove member array
func (s ArrInt32) Remove(array []int32, value int32) []int32 {
	isExist, index := s.InArray(value, array)
	if isExist {
		array = append(array[:index], array[(index+1):]...)
	}

	return array
}

// InArray check is in array
func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

// Remove member array
func Remove(array interface{}, value interface{}) interface{} {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		isExist, index := InArray(value, array)
		if isExist {
			switch reflect.TypeOf(reflect.ValueOf(array).Index(0).Interface()).Kind() {
			case reflect.Bool:
				array = append(array.([]bool)[:index], array.([]bool)[(index+1):]...)
			case reflect.Int:
				array = append(array.([]int)[:index], array.([]int)[(index+1):]...)
			case reflect.Int8:
				array = append(array.([]int8)[:index], array.([]int8)[(index+1):]...)
			case reflect.Int16:
				array = append(array.([]int16)[:index], array.([]int16)[(index+1):]...)
			case reflect.Int32:
				array = append(array.([]int32)[:index], array.([]int32)[(index+1):]...)
			case reflect.Int64:
				array = append(array.([]int64)[:index], array.([]int64)[(index+1):]...)
			case reflect.Uint:
				array = append(array.([]uint)[:index], array.([]uint)[(index+1):]...)
			case reflect.Uint8:
				array = append(array.([]uint8)[:index], array.([]uint8)[(index+1):]...)
			case reflect.Uint16:
				array = append(array.([]uint16)[:index], array.([]uint16)[(index+1):]...)
			case reflect.Uint32:
				array = append(array.([]uint32)[:index], array.([]uint32)[(index+1):]...)
			case reflect.Uint64:
				array = append(array.([]uint64)[:index], array.([]uint64)[(index+1):]...)
			case reflect.Float32:
				array = append(array.([]float32)[:index], array.([]float32)[(index+1):]...)
			case reflect.Float64:
				array = append(array.([]float64)[:index], array.([]float64)[(index+1):]...)
			case reflect.String:
				array = append(array.([]string)[:index], array.([]string)[(index+1):]...)

			}

		}
	}

	return array
}

// Array Unique filtered
func (s ArrInt32) Unique(intSlice []int32) []int32 {
	keys := make(map[int32]bool)
	list := []int32{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}
