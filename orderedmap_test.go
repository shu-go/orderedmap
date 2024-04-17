package orderedmap_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/shu-go/gotwant"
	"github.com/shu-go/orderedmap"
)

func Example() {
	type myStruct struct {
		V1 string
		V2 int
	}
	m := orderedmap.New[string, myStruct]()

	m.Set("ichi", myStruct{
		V1: "one",
		V2: 1,
	})
	m.Set("ni", myStruct{
		V1: "two",
		V2: 2,
	})
	m.Set("san", myStruct{
		V1: "three",
		V2: 3,
	})

	v, found := m.Get("ichi")
	fmt.Println(found)
	fmt.Println(v)

	fmt.Println()

	fmt.Println(m.Contains("ni"))
	fmt.Println(m.Contains("go"))
	// Output:
	// true
	// {one 1}
	//
	// true
	// false
}

func Example_marshal() {
	m := orderedmap.New[int, string]()
	m.Set(5, "go")
	m.Set(9, "ku")
	m.Set(6, "ro-")
	m.Set(3, "san")

	b, _ /*err*/ := json.MarshalIndent(m, "", "  ")

	fmt.Println(string(b))
	// Output:
	// {
	//   "5": "go",
	//   "9": "ku",
	//   "6": "ro-",
	//   "3": "san"
	// }

}

func Example_unmarshal() {
	m := orderedmap.New[int, string]()
	m.Set(999, "dummy")

	s := `{
	  "5": "go",
	  "9": "ku",
	  "6": "ro-",
	  "3": "san"
	}`

	_ /*err*/ = json.Unmarshal([]byte(s), m)

	fmt.Println(m.Get(5))
	fmt.Println(m.Get(9))
	fmt.Println(m.Get(6))
	fmt.Println(m.Get(3))
	fmt.Println(m.Get(999))
	fmt.Println(m.Get(9999))
	// Output:
	// go true
	// ku true
	// ro- true
	// san true
	//  false
	//  false
}

func Example_sort() {
	m := orderedmap.New[string, int]()
	m.Set("5", 1)
	m.Set("9", 2)
	m.Set("6", 3)
	m.Set("3", 4)

	m.Sort(func(i, j string) bool {
		return i < j
	})

	b, _ /*err*/ := json.MarshalIndent(m, "", "  ")

	fmt.Println(string(b))
	// Output:
	// {
	//   "3": 4,
	//   "5": 1,
	//   "6": 3,
	//   "9": 2
	// }
}

func FuzzRandomOperations(f *testing.F) {
	m := orderedmap.New[uint, uint]()
	var keys []uint
	f.Fuzz(func(t *testing.T, key uint, op bool) {
		if op {
			key = key % 1000
			m.Set(key, key)

			found := false
			for _, k := range keys {
				if k == key {
					found = true
					break
				}
			}
			if !found {
				keys = append(keys, key)
			}
			gotwant.Test(t, m.Keys(), keys)

			v, found := m.Get(key)
			gotwant.Test(t, found, true)
			gotwant.Test(t, v, key)

		} else {
			m.Delete(key)

			idx := -1
			for i, k := range keys {
				if k == key {
					idx = i
					break
				}
			}
			if idx != -1 {
				keys = append(keys[:idx], keys[idx+1:]...)
			}
			gotwant.Test(t, m.Keys(), keys)

			_, found := m.Get(key)
			gotwant.Test(t, found, false)
		}
	})
}

func FuzzRandomOperationsReordered(f *testing.F) {
	m := orderedmap.New[uint, uint]()
	m.PreserveOrder(false)
	var keys []uint
	f.Fuzz(func(t *testing.T, key uint, op bool) {
		if op {
			key = key % 1000
			m.Set(key, key)

			idx := -1
			for i, k := range keys {
				if k == key {
					idx = i
					break
				}
			}
			if idx != -1 {
				keys = append(keys[:idx], keys[idx+1:]...)
			}
			keys = append(keys, key)
			gotwant.Test(t, m.Keys(), keys)

			v, found := m.Get(key)
			gotwant.Test(t, found, true)
			gotwant.Test(t, v, key)

		} else {
			m.Delete(key)

			idx := -1
			for i, k := range keys {
				if k == key {
					idx = i
					break
				}
			}
			if idx != -1 {
				keys = append(keys[:idx], keys[idx+1:]...)
			}
			gotwant.Test(t, m.Keys(), keys)

			_, found := m.Get(key)
			gotwant.Test(t, found, false)
		}
	})
}

func TestPrint(t *testing.T) {
	type mystruct struct {
		A, B string
	}
	std := make(map[string]mystruct)
	std["ichi"] = mystruct{
		A: "a1",
		B: "b1",
	}
	std["ni"] = mystruct{
		A: "a2",
		B: "b2",
	}
	std["san"] = mystruct{
		A: "a3",
		B: "b3",
	}

	m := orderedmap.New[string, mystruct]()
	m.Set("ichi", mystruct{
		A: "a1",
		B: "b1",
	})
	m.Set("ni", mystruct{
		A: "a2",
		B: "b2",
	})
	m.Set("san", mystruct{
		A: "a3",
		B: "b3",
	})

	stdstr := fmt.Sprint(std)
	omstr := fmt.Sprint(m)
	omstr = strings.ReplaceAll(omstr, "OrderedMap", "map")
	gotwant.Test(t, omstr, stdstr)

	stdstr = fmt.Sprintf("%+v", std)
	omstr = fmt.Sprintf("%+v", m)
	omstr = strings.ReplaceAll(omstr, "OrderedMap", "map")
	gotwant.Test(t, omstr, stdstr)

	stdstr = fmt.Sprintf("%#v", std)
	omstr = fmt.Sprintf("%#v", m)
	omstr = strings.ReplaceAll(omstr, "OrderedMap", "map")
	gotwant.Test(t, omstr, stdstr)

	stdstr = fmt.Sprintf("%s", std)
	omstr = fmt.Sprintf("%s", m)
	omstr = strings.ReplaceAll(omstr, "OrderedMap", "map")
	gotwant.Test(t, omstr, stdstr)

	m2 := orderedmap.New[string, any]()
	err := m2.UnmarshalJSON([]byte(`{"a":1,"z":999,"b":2}`))
	gotwant.TestError(t, err, nil)
	gotwant.Test(t, m2.Keys(), []string{"a", "z", "b"})
	omstr = fmt.Sprintf("%#v", m2)
	gotwant.Test(t, omstr, `OrderedMap[string]interface {}{"a":1, "z":999, "b":2}`)

}

func TestSort(t *testing.T) {
	m := orderedmap.New[string, int]()
	m.Set("5", 0)
	m.Set("9", 0)
	m.Set("6", 0)
	m.Set("3", 0)

	m.Sort(func(i, j string) bool {
		return i < j
	})
	gotwant.Test(t, m.Keys(), []string{"3", "5", "6", "9"})

	b, err := json.Marshal(m)
	gotwant.TestError(t, err, nil)
	gotwant.Test(t, string(b), `{"3":0,"5":0,"6":0,"9":0}`)

	//

	m.Sort(func(i, j string) bool {
		return i > j
	})
	gotwant.Test(t, m.Keys(), []string{"9", "6", "5", "3"})

	b, err = json.Marshal(m)
	gotwant.TestError(t, err, nil)
	gotwant.Test(t, string(b), `{"9":0,"6":0,"5":0,"3":0}`)
}

func TestMarshal(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		m := orderedmap.New[string, string]()
		j, err := m.MarshalJSON()
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, string(j), `{}`)
	})

	t.Run("String2Int", func(t *testing.T) {
		m := orderedmap.New[string, int]()
		m.Set("b", 2)
		m.Set("a", 1)
		m.Set("c", 3)

		j, err := json.Marshal(m)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, string(j), `{"b":2,"a":1,"c":3}`)
	})

	t.Run("Int2String", func(t *testing.T) {
		m := orderedmap.New[int, string]()
		m.Set(2, "b")
		m.Set(1, "a")
		m.Set(3, "c")

		j, err := json.Marshal(m)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, string(j), `{"2":"b","1":"a","3":"c"}`)
	})

	t.Run("MyStruct", func(t *testing.T) {
		type myStruct struct {
			V1 string `json:"v"`
			V2 string `json:"w"`
		}
		m := orderedmap.New[string, myStruct]()

		m.Set("a", myStruct{V1: "a-v1", V2: "a-v2"})
		m.Set("b", myStruct{V1: "b-v1", V2: "b-v2"})

		j, err := json.Marshal(m)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, string(j), `{"a":{"v":"a-v1","w":"a-v2"},"b":{"v":"b-v1","w":"b-v2"}}`)

		j, err = json.MarshalIndent(m, "", "  ")
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, string(j), `{
  "a": {
    "v": "a-v1",
    "w": "a-v2"
  },
  "b": {
    "v": "b-v1",
    "w": "b-v2"
  }
}`)
	})

	t.Run("Nest", func(t *testing.T) {
		m := orderedmap.New[string, *orderedmap.OrderedMap[string, string]]()
		subm := orderedmap.New[string, string]()
		subm.Set("sub1", "ichi")
		subm.Set("sub2", "ni")
		m.Set("1", subm)
		subm = orderedmap.New[string, string]()
		subm.Set("sub3", "san")
		m.Set("2", subm)

		j, err := json.MarshalIndent(m, "", "  ")
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, string(j), `{
  "1": {
    "sub1": "ichi",
    "sub2": "ni"
  },
  "2": {
    "sub3": "san"
  }
}`)
	})
}

func TestUnmarshal(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		m := orderedmap.New[string, string]()
		err := json.Unmarshal([]byte(`{}`), &m)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, len(m.Keys()), 0)
	})

	t.Run("String2Int", func(t *testing.T) {
		m := orderedmap.New[string, int]()
		err := json.Unmarshal([]byte(`{"z":1, "b": 2}`), &m)
		gotwant.TestError(t, err, nil)

		v, found := m.Get("b")
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, 2)

		gotwant.Test(t, m.Keys(), []string{"z", "b"})
	})

	t.Run("Int2String", func(t *testing.T) {
		m := orderedmap.New[int, string]()
		err := json.Unmarshal([]byte(`{"9":"a", "2": "b"}`), &m)
		gotwant.TestError(t, err, nil)

		v, found := m.Get(2)
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, "b")

		gotwant.Test(t, m.Keys(), []int{9, 2})
	})

	t.Run("MyStruct", func(t *testing.T) {
		type myStruct struct {
			V1 string `json:"v"`
			V2 string `json:"w"`
		}
		m := orderedmap.New[string, myStruct]()
		err := json.Unmarshal([]byte(`{
  "b": {
    "v": "b-v1",
    "w": "b-v2"
  },
  "1": {
    "v": "a-v1",
    "w": "a-v2"
  }
}`), &m)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, m.Keys(), []string{"b", "1"})

		v, found := m.Get("b")
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, myStruct{V1: "b-v1", V2: "b-v2"})

		v, found = m.Get("1")
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, myStruct{V1: "a-v1", V2: "a-v2"})
	})

	t.Run("MyStruct", func(t *testing.T) {
		type myStruct struct {
			V1 string `json:"v"`
			V2 string `json:"w"`
		}
		m := orderedmap.New[string, myStruct]()
		err := m.UnmarshalJSON([]byte(`{
  "b": {
    "v": "b-v1",
    "w": "b-v2"
  },
  "1": {
    "v": "a-v1",
    "w": "a-v2"
  }
}`))
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, m.Keys(), []string{"b", "1"})

		v, found := m.Get("b")
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, myStruct{V1: "b-v1", V2: "b-v2"})

		v, found = m.Get("1")
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, myStruct{V1: "a-v1", V2: "a-v2"})
	})

	t.Run("Nest", func(t *testing.T) {
		m := orderedmap.New[string, *orderedmap.OrderedMap[string, string]]()

		err := json.Unmarshal([]byte(`{
  "2": {
    "sub3": "san"
  },
  "1": {
    "sub1": "ichi",
    "sub2": "ni"
  }
}`), &m)
		gotwant.TestError(t, err, nil)

		subm, found := m.Get("2")
		gotwant.Test(t, found, true)

		v, found := subm.Get("sub2")
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, "ni")

		gotwant.Test(t, m.Keys(), []string{"2", "1"})

	})
}

func TestPreserveOrder(t *testing.T) {
	t.Run("Preserved", func(t *testing.T) {
		m := orderedmap.New[int, int]()
		m.PreserveOrder(true) // default

		m.Set(1, 0)
		m.Set(2, 0)
		m.Set(3, 0)
		gotwant.Test(t, m.Keys(), []int{1, 2, 3})

		m.Set(1, 100)
		gotwant.Test(t, m.Keys(), []int{1, 2, 3})
		v, found := m.Get(1)
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, 100)

		m.Set(1, 200)
		gotwant.Test(t, m.Keys(), []int{1, 2, 3})
		v, found = m.Get(1)
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, 200)

	})

	t.Run("Reordered", func(t *testing.T) {
		m := orderedmap.New[int, int]()
		m.PreserveOrder(false)

		m.Set(1, 0)
		m.Set(2, 0)
		m.Set(3, 0)
		gotwant.Test(t, m.Keys(), []int{1, 2, 3})

		m.Set(1, 100)
		gotwant.Test(t, m.Keys(), []int{2, 3, 1})
		v, found := m.Get(1)
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, 100)

		m.Set(1, 200)
		gotwant.Test(t, m.Keys(), []int{2, 3, 1})
		v, found = m.Get(1)
		gotwant.Test(t, found, true)
		gotwant.Test(t, v, 200)
	})
}

func TestDelete(t *testing.T) {
	t.Run("Last", func(t *testing.T) {
		m := orderedmap.New[int, int]()
		m.Set(1, 0)
		m.Set(2, 0)
		m.Set(3, 0)
		m.Delete(3)
		gotwant.Test(t, m.Keys(), []int{1, 2})
	})

	t.Run("DeleteAndSet", func(t *testing.T) {
		m := orderedmap.New[int, int]()
		m.Set(1, 0)
		m.Set(2, 0)
		m.Set(3, 0)
		m.Delete(2)
		m.Set(3, 0)
		gotwant.Test(t, m.Keys(), []int{1, 3})

		m = orderedmap.New[int, int]()
		m.PreserveOrder(false) //
		m.Set(1, 0)
		m.Set(2, 0)
		m.Set(3, 0)
		m.Delete(2)
		m.Set(3, 0)
		gotwant.Test(t, m.Keys(), []int{1, 3})
	})
}

func TestUnorderedMap(t *testing.T) {
	m := orderedmap.New[int, int]()
	for i := 0; i < 10000; i++ {
		r := rand.Int() % 1000
		m.Set(r, r)
	}

	u := m.UnorderedMap()

	// keys
	ukeys := []int{}
	for k := range u {
		ukeys = append(ukeys, k)
	}
	keys := m.Keys()

	sort.Slice(ukeys, func(i, j int) bool {
		return ukeys[i] < ukeys[j]
	})
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	gotwant.Test(t, ukeys, keys)

	for _, k := range ukeys {
		uv, ufound := u[k]
		ov, ofound := m.Get(k)
		gotwant.Test(t, ufound, ofound)
		gotwant.Test(t, uv, ov)
	}
}

func TestLen(t *testing.T) {
	m := orderedmap.New[string, int]()
	gotwant.Test(t, m.Len(), 0)

	m.Set("ichi", 1)
	gotwant.Test(t, m.Len(), 1)
	m.Set("ichi", 11)
	gotwant.Test(t, m.Len(), 1)
	m.Set("ni", 2)
	gotwant.Test(t, m.Len(), 2)
	m.Delete("ichi")
	gotwant.Test(t, m.Len(), 1)
	m.Delete("ni")
	gotwant.Test(t, m.Len(), 0)
}

func TestNil(t *testing.T) {
	var std map[string]int
	var m *orderedmap.OrderedMap[string, int]

	t.Run("Set", func(t *testing.T) {
		gotwant.TestPanic(t, func() {
			std["ichi"] = 1
		}, "assignment to entry in nil map")
		gotwant.TestPanic(t, func() {
			m.Set("ichi", 1)
		}, "assignment to entry in nil map")
	})

	t.Run("Get", func(t *testing.T) {
		gotwant.TestPanic(t, func() {
			_, found := std["ichi"]
			gotwant.Test(t, found, false)
		}, nil)
		gotwant.TestPanic(t, func() {
			_, found := m.Get("ichi")
			gotwant.Test(t, found, false)
		}, nil)
	})

	t.Run("GetDefault", func(t *testing.T) {
		gotwant.TestPanic(t, func() {
			_, found := std["ichi"]
			gotwant.Test(t, found, false)
		}, nil)
		gotwant.TestPanic(t, func() {
			ichi := m.GetDefault("ichi", 1)
			gotwant.Test(t, ichi, 1)
		}, nil)
	})

	t.Run("Delete", func(t *testing.T) {
		gotwant.TestPanic(t, func() {
			delete(std, "ichi")
		}, nil)
		gotwant.TestPanic(t, func() {
			m.Delete("ichi")
		}, nil)
	})

	t.Run("Keys", func(t *testing.T) {
		gotwant.TestPanic(t, func() {
			for k := range std {
				k = k + ""
			}
		}, nil)
		gotwant.TestPanic(t, func() {
			_ = m.Keys()
		}, nil)
	})

	t.Run("Len", func(t *testing.T) {
		gotwant.TestPanic(t, func() {
			gotwant.Test(t, len(std), 0)
		}, nil)
		gotwant.TestPanic(t, func() {
			gotwant.Test(t, m.Len(), 0)
		}, nil)
	})

	t.Run("Contains", func(t *testing.T) {
		gotwant.TestPanic(t, func() {
			_, found := std["ichi"]
			gotwant.Test(t, found, false)
		}, nil)
		gotwant.TestPanic(t, func() {
			found := m.Contains("ichi")
			gotwant.Test(t, found, false)
		}, nil)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		gotwant.TestPanic(t, func() {
			b, err := json.Marshal(std)
			gotwant.TestError(t, err, nil)
			gotwant.Test(t, string(b), "null")
		}, nil)
		gotwant.TestPanic(t, func() {
			b, err := json.Marshal(m)
			gotwant.TestError(t, err, nil)
			gotwant.Test(t, string(b), "null")

			var mm orderedmap.OrderedMap[string, int]
			b, err = json.Marshal(mm)
			gotwant.TestError(t, err, nil)
			gotwant.Test(t, string(b), "{}")
		}, nil)
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		gotwant.TestPanic(t, func() {
			var std map[string]int

			err := json.Unmarshal([]byte(`{"ichi":1}`), &std)
			gotwant.TestError(t, err, nil)
			gotwant.Test(t, std["ichi"], 1)
		}, nil)
		gotwant.TestPanic(t, func() {
			var m *orderedmap.OrderedMap[string, int]

			err := json.Unmarshal([]byte(`{"ichi":1}`), &m)
			gotwant.TestError(t, err, nil)
			gotwant.Test(t, m.GetDefault("ichi", 999), 1)
		}, nil)
	})
}

func BenchmarkSet(b *testing.B) {
	std := make(map[string]int)
	m := orderedmap.New[string, int]()

	const kcount = 1000
	keys := []string{}
	values := []int{}
	for i := 0; i < kcount; i++ {
		r := rand.Int()
		keys = append(keys, strconv.Itoa(r))
		values = append(values, r)
	}

	b.Run("StdMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k := keys[i%kcount]
			v := values[i%kcount]
			std[k] = v
		}
	})
	b.Run("OM", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k := keys[i%kcount]
			v := values[i%kcount]
			m.Set(k, v)
		}
	})
}

func BenchmarkGet(b *testing.B) {
	std := make(map[string]int)
	m := orderedmap.New[string, int]()

	const count = 1000
	for i := 0; i < count; i++ {
		v := rand.Int()
		k := strconv.Itoa(v)

		std[k] = v
		m.Set(k, v)
	}

	const kcount = 100
	keys := []string{}
	for i := 0; i < kcount; i++ {
		keys = append(keys, strconv.Itoa(rand.Int()))
	}

	b.Run("StdMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k := keys[i%kcount]
			_ = std[k]
		}
	})
	b.Run("OM", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k := keys[i%kcount]
			_, _ = m.Get(k)
		}
	})
}

func BenchmarkDelete(b *testing.B) {
	std := make(map[string]int)
	m := orderedmap.New[string, int]()

	const count = 1000
	for i := 0; i < count; i++ {
		v := rand.Int()
		k := strconv.Itoa(v)

		std[k] = v
		m.Set(k, v)
	}

	const kcount = 100
	keys := []string{}
	for i := 0; i < kcount; i++ {
		keys = append(keys, strconv.Itoa(rand.Int()))
	}

	b.Run("StdMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k := keys[i%kcount]
			delete(std, k)
		}
	})
	b.Run("OM", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k := keys[i%kcount]
			m.Delete(k)
		}
	})
}

func BenchmarkKeys(b *testing.B) {
	std := make(map[string]int)
	m := orderedmap.New[string, int]()

	for i := 0; i < 1000; i++ {
		v := rand.Int()
		k := strconv.Itoa(v)

		std[k] = v
		m.Set(k, v)
	}

	b.Run("StdMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			keys := make([]string, 0, len(std))
			for k := range std {
				keys = append(keys, k)
			}
		}
	})
	b.Run("OM", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = m.Keys()
		}
	})
}

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte(`{
  "number": 4,
  "string": "x",
  "z": 1,
  "a": "should not break with unclosed { character in value",
  "b": 3,
  "slice": [
    "1",
    1
  ],
  "orderedmap": {
    "e": 1,
    "a { nested key with brace": "with a }}}} }} {{{ brace value",
	"after": {
		"link": "test {{{ with even deeper nested braces }"
	}
  },
  "test\"ing": 9,
  "after": 1,
  "multitype_array": [
    "test",
	1,
	{ "map": "obj", "it" : 5, ":colon in key": "colon: in value" },
	[{"inner": "map"}]
  ],
  "should not break with { character in key": 1
}`)
	b.Run("StdMap", func(b *testing.B) {
		std := make(map[string]any)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			json.Unmarshal(data, &std)
		}
	})
	b.Run("OM", func(b *testing.B) {
		m := orderedmap.New[string, any]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			json.Unmarshal(data, &m)
		}
	})
	b.Run("OM#Direct", func(b *testing.B) {
		m := orderedmap.New[string, any]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.UnmarshalJSON(data)
		}
	})

	data = data[:0]
	data = append(data, '{')
	for i := 0; i < 10000; i++ {
		if i > 0 {
			data = append(data, ',')
		}
		data = append(data, []byte(fmt.Sprintf(`"key%v": %d`, i, i))...)
	}
	data = append(data, '}')
	b.Run("StdMap", func(b *testing.B) {
		std := make(map[string]any)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			json.Unmarshal(data, &std)
		}
	})
	b.Run("OM", func(b *testing.B) {
		m := orderedmap.New[string, any]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			json.Unmarshal(data, &m)
		}
	})
	b.Run("OM#Direct", func(b *testing.B) {
		m := orderedmap.New[string, any]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.UnmarshalJSON(data)
		}
	})
}

func BenchmarkMarshal(b *testing.B) {
	std := make(map[string]any)
	std["number"] = 3
	std["string"] = "x"
	std["specialstring"] = "\\.<>[]{}_-"
	std["number"] = 4
	std["z"] = 1
	std["a"] = 2
	std["b"] = 3
	std["slice"] = []interface{}{
		"1",
		1,
	}
	substd := make(map[string]int)
	substd["e"] = 1
	substd["a"] = 2
	std["orderedmap"] = substd
	std["test\n\r\t\\\"ing"] = 9

	m := orderedmap.New[string, any]()
	m.Set("number", 3)
	m.Set("string", "x")
	m.Set("specialstring", "\\.<>[]{}_-")
	m.Set("number", 4)
	m.Set("z", 1)
	m.Set("a", 2)
	m.Set("b", 3)
	m.Set("slice", []interface{}{
		"1",
		1,
	})
	subm := orderedmap.New[string, int]()
	subm.Set("e", 1)
	subm.Set("a", 2)
	m.Set("orderedmap", subm)
	m.Set("test\n\r\t\\\"ing", 9)

	b.Run("StdMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			json.Marshal(std)
		}
	})
	b.Run("OM", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			json.Marshal(m)
		}
	})
	b.Run("OM#Direct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.MarshalJSON()
		}
	})
}

func BenchmarkSort(b *testing.B) {
	b.Run("OM", func(b *testing.B) {
		m := orderedmap.New[string, int]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			for i := 0; i < 1000; i++ {
				v := rand.Int()
				k := strconv.Itoa(v)
				m.Set(k, v)
			}
			b.StartTimer()

			m.Sort(func(i, j string) bool {
				return i < j
			})
		}
	})
}
