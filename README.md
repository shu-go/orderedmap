Generic orderedmap with (Un)MarshalJSON

[![Go Report Card](https://goreportcard.com/badge/github.com/shu-go/orderedmap)](https://goreportcard.com/report/github.com/shu-go/orderedmap)
![MIT License](https://img.shields.io/badge/License-MIT-blue)

# go get

```
go get github.com/shu-go/orderedmap
```

# Codes

## New

```
var m *orderedmap.OrderedMap[int, int]
m = orderedmap.New[int, int]()

// or simply,
m := orderedmap.New[int, int]()
```

## Set

```
m.Set(1, 100)
m.Set(9, 900)
m.Set(2, 200)

m.Keys() //=> []int{1, 9, 2}
```

### Re-Order

```
m.PreserveOrder(false)
m.Set(1, 100)

m.Keys() //=> []int{9, 2, 1}
```

## Get

```
v, found := m.Get(1)
```

## GetDefault

```
v := m.Get(8, 1234)
```

## Keys

```
keys := m.Keys() // [1, 9, 2]
```

## Delete

```
m.Delete(2)
m.Delete(10) // no error

m.Keys() //=> [1, 9]
```

## Contains

```
m.Contains(8) //=> false
m.Contains(1) //=> true
```

## JSON

### MarshalJSON

```
data, err := json.Marshal(m) //=> `{"1":100,"9":900}`
```

### UnmarshalJSON

```
data := []byte(`{"100":1000,"200":2000}`)
err := json.Unmarshal(data, &m) // CLEARED and Unmarshalled

// m.UnmarshalJSON(data) is faster.
```

## Sort

```
m := orderedmap.New[string, any]()
m.UnmarshalJSON([]byte(`{"a":1,"z":999,"b":2}`))

m.Keys() //=> ["a", "z", "b"]

m.Sort(func(i, j string) bool {
    // i,j are keys
    return i < j
})

m.Keys() //=> ["a", "b", "z"]
```

## Format

```
m := orderedmap.New[string, any]()
m.UnmarshalJSON([]byte(`{"a":1,"z":999,"b":2}`))

fmt.Sprint(m) //=> OrderedMap[a:1 z:999 b:2]

fmt.Sprintf("%#v", m) //=> OrderedMap[string]interface {}{"a":1, "z":999, "b":2}
```


<!-- vim: set et ft=markdown sts=4 sw=4 ts=4 tw=0 : -->
