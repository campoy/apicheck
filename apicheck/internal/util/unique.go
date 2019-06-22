package util

import (
	"reflect"
	"sort"
)

// SortUnique concatenats and sorts the given lists and removes all repeated elements.
func SortUnique(s ...[]string) []string {
	if len(s) == 0 {
		return nil
	}

	var res []string
	for _, t := range s {
		res = append(res, t...)
	}
	sort.Strings(res)
	i := 1
	for i < len(res) {
		if res[i-1] != res[i] {
			i++
			continue
		}
		copy(res[i:], res[i+1:])
		res = res[:len(res)-1]
	}
	return res
}

// KeysFromMap returns a slice with the keys in the given map as long as they're strings.
func KeysFromMap(m interface{},
	toString func(v interface{}) string,
	filter func(v interface{}) bool) []string {
	v := reflect.ValueOf(m)
	keys := make([]string, v.Len())
	for i, k := range v.MapKeys() {
		if filter != nil && !filter(v.MapIndex(k).Interface()) {
			continue
		}
		s := k.Interface()
		if toString != nil {
			keys[i] = toString(s)
		} else {
			keys[i] = s.(string)
		}
	}
	return keys
}
