package orderedmap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"reflect"
	"sort"
	"strings"

	"github.com/shu-go/jbdec"
	"gopkg.in/yaml.v3"
)

type elem[V any] struct {
	v V

	maxIdx int
}

type OrderedMap[K comparable, V any] struct {
	m map[K]*elem[V]

	keys []K

	overwriteSeq bool

	work bytes.Buffer
}

func New[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		m:    make(map[K]*elem[V]),
		keys: nil,

		overwriteSeq: false,
	}
}

func (m *OrderedMap[K, V]) PreserveOrder(b bool) {
	m.overwriteSeq = !b
}

func (m *OrderedMap[K, V]) Set(key K, value V) {
	if m == nil {
		panic("assignment to entry in nil map")
	}

	if e, found := m.m[key]; !found {
		m.keys = append(m.keys, key)
		m.m[key] = &elem[V]{
			maxIdx: len(m.keys),
			v:      value,
		}

	} else {
		e.v = value
		if m.overwriteSeq {
			idx := m.indexOfKey(key, e.maxIdx)
			e.maxIdx = idx
			m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
			m.keys = append(m.keys, key)
		}
	}
}

func (m *OrderedMap[K, V]) Delete(key K) {
	if m == nil {
		return
	}

	e, found := m.m[key]

	if !found {
		return
	}
	delete(m.m, key)

	idx := m.indexOfKey(key, e.maxIdx)
	e.maxIdx = idx
	m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
}

func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	if m == nil {
		var gnil V
		return gnil, false
	}

	if e, found := m.m[key]; found {
		return e.v, true
	}

	var gnil V
	return gnil, false
}

func (m *OrderedMap[K, V]) GetDefault(key K, defvalue V) V {
	if m == nil {
		return defvalue
	}

	value, found := m.Get(key)
	if found {
		return value
	}
	return defvalue
}

func (m *OrderedMap[K, V]) Len() int {
	if m == nil {
		return 0
	}
	return len(m.keys)
}

func (m *OrderedMap[K, V]) Keys() []K {
	if m == nil {
		return nil
	}

	return m.keys
}

func (m *OrderedMap[K, V]) Contains(key K) bool {
	if m == nil {
		return false
	}

	_, found := m.m[key]
	return found
}

func (m *OrderedMap[K, V]) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}

	buf := &m.work
	buf.Reset()
	buf.Grow(len(m.m) * 8)

	buf.WriteByte('{')

	for i, k := range m.keys {
		e, found := m.m[k]
		if !found {
			continue
		}

		if i > 0 {
			buf.WriteByte(',')
		}

		if s, ok := any(k).(string); ok {
			buf.WriteByte('"')
			buf.WriteString(s)
			buf.WriteByte('"')

		} else {
			b, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			buf.WriteByte('"')
			buf.Write(b)
			buf.WriteByte('"')
		}

		buf.WriteByte(':')

		b, err := json.Marshal(e.v)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

func (m *OrderedMap[K, V]) UnmarshalJSON(b []byte) error {
	*m = *New[K, V]() // clear

	if len(b) == 0 {
		return nil
	}

	m.m = make(map[K]*elem[V])
	m.keys = m.keys[:0]

	var key K
	var value V

	parsingKey := true
	valueBuf := &m.work
	valueBuf.Reset()
	stack := []byte{}

	dec := jbdec.New(b)

	tok := dec.Next()
	if tok.Type == jbdec.EOF {
		return nil
	}
	if tok.Type != jbdec.BeginObject {
		return errors.New("must begin with {")
	}

	for {
		tok := dec.Next()
		if tok.Type == jbdec.EOF {
			if !parsingKey {
				return errors.New("sudden EOF")
			}
			break
		}

		if tok.Type == jbdec.EndObject && parsingKey {
			break
		}

		if tok.Type == jbdec.ValueSeparator && parsingKey {
			continue
		}

		if parsingKey {
			s := tok.Bytes()
			err := json.Unmarshal(s, &key)
			if err != nil {
				if len(s) > 2 {
					s = s[1 : len(s)-1] // trim "
					err2 := json.Unmarshal(s, &key)
					if err2 != nil {
						return err2
					}
				} else {
					return err
				}
			}

			colTok := dec.Next() // :
			if colTok.Type != jbdec.NameSeparator {
				return errors.New(": is required")
			}

			parsingKey = false
			valueBuf.Reset()

			continue
		}

		// not parsingKey

		if tok.Type == jbdec.BeginObject {
			valueBuf.WriteByte('{')
			stack = append(stack, '{')

			//log.Print("BeginObject", "=>", valueBuf.String())

		} else if tok.Type == jbdec.EndObject {
			if len(stack) == 0 {
				return errors.New("} when stack is empty")
			} else if stack[len(stack)-1] != '{' {
				return errors.New("} is mismatch")
			}

			valueBuf.WriteByte('}')
			stack = stack[:len(stack)-1]

			//log.Print("EndObject", "=>", valueBuf.String())

		} else if tok.Type == jbdec.BeginArray {
			valueBuf.WriteByte('[')
			stack = append(stack, '[')

			//log.Print("BeginArray", "=>", valueBuf.String())

		} else if tok.Type == jbdec.EndArray {
			if len(stack) == 0 {
				return errors.New("] when stack is empty")
			} else if stack[len(stack)-1] != '[' {
				return errors.New("] is mismatch")
			}

			//log.Print("EndArray", "=>", valueBuf.String())

			valueBuf.WriteByte(']')
			stack = stack[:len(stack)-1]

		} else if len(stack) == 0 && tok.Type == jbdec.ValueSeparator {
			//log.Print("IGNORED")
			//nop
		} else if tok.Type == jbdec.NameSeparator {
			valueBuf.WriteByte(':')

			//log.Print("NameSeparator", "=>", valueBuf.String())

		} else if tok.Type == jbdec.ValueSeparator {
			valueBuf.WriteByte(',')

			//log.Print("ValueSeparator", "=>", valueBuf.String())

		} else {
			s := tok.Bytes()
			valueBuf.Write(s)

			//log.Print("else ", string(s), "=>", valueBuf.String())
		}

		if len(stack) == 0 {
			//log.Printf("%v: %q", key, valueBuf.String())

			err := json.Unmarshal(valueBuf.Bytes(), &value)
			if err != nil {
				return err
			}
			//log.Printf("%v: %#v", key, value)

			m.Set(key, value)

			parsingKey = true
		}
	}

	return nil
}

func (m *OrderedMap[K, V]) MarshalYAML() (any, error) {
	if m == nil || len(m.keys) == 0 {
		return nil, nil
	}

	var v V
	vtype := reflect.TypeOf(v)
	fields := make([]reflect.StructField, 0, len(m.keys))
	for i := 0; i < len(m.keys); i++ {
		fields = append(fields, reflect.StructField{
			Name: fmt.Sprintf("F%05d", i),
			Type: vtype,
			Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%v"`, m.keys[i])),
		})
	}

	stType := reflect.StructOf(fields)
	st := reflect.New(stType).Elem()

	for i := 0; i < len(m.keys); i++ {
		st.FieldByIndex([]int{i}).Set(reflect.ValueOf(m.GetDefault(m.keys[i], v)))
	}

	return st.Interface(), nil
}

// NOT SUPPORTED: number key, nested OrderedMap
func (m *OrderedMap[K, V]) UnmarshalYAML(value *yaml.Node) error {
	*m = *New[K, V]() // clear

	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i]
		val := value.Content[i+1]

		var k K
		var v V

		//if err := key.Decode(&k); err != nil {
		//	return err
		//}
		if err := key.Decode(&k); err != nil {
			if len(key.Value) >= 2 && key.Value[0] == '"' && key.Value[len(key.Value)-1] == '"' {
				key.Value = key.Value[1 : len(key.Value)-1]
				err = key.Decode(&k)
			}
			return err
		}
		if err := val.Decode(&v); err != nil {
			return err
		}

		m.Set(k, v)
	}

	return nil
}

func (m *OrderedMap[K, V]) UnorderedMap() map[K]V {
	u := make(map[K]V)

	for k, e := range m.m {
		u[k] = e.v
	}

	return u
}

func (m *OrderedMap[K, V]) Sort(less func(K, K) bool) {
	handler := SliceHandler{
		len: func() int {
			return len(m.keys)
		},
		less: func(i, j int) bool {
			return less(m.keys[i], m.keys[j])
		},
		swap: func(i, j int) {
			ikey := m.keys[i]
			jkey := m.keys[j]
			ielem := m.m[ikey]
			jelem := m.m[jkey]

			m.keys[i], m.keys[j] = m.keys[j], m.keys[i]

			// swap elems to swap maxIdx
			ielem.maxIdx = j
			jelem.maxIdx = i
			m.m[ikey] = ielem
			m.m[jkey] = jelem
		},
	}

	sort.Sort(handler)
}

func (m OrderedMap[K, V]) Format(s fmt.State, verb rune) {
	sb := &strings.Builder{}

	switch true {
	case s.Flag('#'):
		var k K
		kname := reflect.TypeOf(k).Name()
		var v V
		var vname string
		vt := reflect.TypeOf(v)
		if vt == nil {
			vname = "interface {}"
		} else {
			vname = path.Base(vt.PkgPath()) + "." + vt.Name()
		}

		sb.WriteString("OrderedMap[")
		sb.WriteString(kname)
		sb.WriteByte(']')
		sb.WriteString(vname)
		sb.WriteByte('{')
		for i, k := range m.keys {
			if i != 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(sb, "%#v:%#v", k, m.m[k].v)
		}
		sb.WriteByte('}')

	case s.Flag('+'):
		sb.WriteString("OrderedMap[")
		for i, k := range m.keys {
			if i != 0 {
				sb.WriteByte(' ')
			}
			fmt.Fprintf(sb, "%+v:%+v", k, m.m[k].v)
		}
		sb.WriteByte(']')

	default:
		sb.WriteString("OrderedMap[")
		for i, k := range m.keys {
			if i != 0 {
				sb.WriteByte(' ')
			}
			fmt.Fprint(sb, k, ":", m.m[k].v)
		}
		sb.WriteByte(']')
	}

	fmt.Fprint(s, sb.String())
}

type SliceHandler struct {
	len  func() int
	less func(i, j int) bool
	swap func(i, j int)
}

func (h SliceHandler) Len() int {
	return h.len()
}

func (h SliceHandler) Less(i, j int) bool {
	return h.less(i, j)
}

func (h SliceHandler) Swap(i, j int) {
	h.swap(i, j)
}

func (m *OrderedMap[K, V]) indexOfKey(key K, maxIdx int) int {
	start := maxIdx
	if len(m.keys)-1 < start {
		start = len(m.keys) - 1
	}

	idx := -1
	for i := start; i >= 0; i-- {
		if m.keys[i] == key {
			idx = i
			break
		}
	}
	if idx == -1 {
		for i := len(m.keys) - 1; i >= 0; i-- {
			if m.keys[i] == key {
				idx = i
				break
			}
		}
	}
	return idx
}
