package consistenthash

import (
	"fmt"
	"strconv"
	"testing"
)

func TestAddKeys(t *testing.T) {
	m := NewMap(3, func(data []byte) uint32 {
		hashValue, _ := strconv.Atoi(string(data))
		return uint32(hashValue)
	})

	m.AddNodes("2", "4", "6")
	fmt.Println(m.Get(""))

	//testCase := map[string]string{
	//	"2":  "2",
	//	"11": "2",
	//	"23": "4",
	//	"27": "2",
	//}
	//for k, v := range testCase {
	//	if m.Get(k) != v {
	//		t.Errorf("Asking for %s, should have yielded %s", k, v)
	//	}
	//}
	//
	//m.AddNodes("8")
	//testCase["27"] = "8"
	//for k, v := range testCase {
	//	if m.Get(k) != v {
	//		t.Errorf("Asking for %s, should have yielded %s", k, v)
	//	}
	//}
}
