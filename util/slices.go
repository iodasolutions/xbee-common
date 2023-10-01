package util

func InSlice(list []interface{}, elt interface{}) bool {
	for _, elt2 := range list {
		if elt2 == elt {
			return true
		}
	}
	return false
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type StringSet struct {
	aMap map[string]bool
}

func NewEmptyStringSet() *StringSet {
	return &StringSet{
		aMap: map[string]bool{},
	}
}
func SetFromStringSlice(list []string) *StringSet {
	result := map[string]bool{}
	for _, s := range list {
		result[s] = true
	}
	return &StringSet{
		aMap: result,
	}
}
func (set *StringSet) HasElement(key string) bool {
	_, ok := set.aMap[key]
	return ok
}
func (set *StringSet) Remove(elts ...string) *StringSet {
	for _, elt := range elts {
		ok := set.HasElement(elt)
		if ok {
			delete(set.aMap, elt)
		}
	}
	return set
}
func (set *StringSet) Add(elts ...string) {
	for _, elt := range elts {
		set.aMap[elt] = true
	}
}

func (set *StringSet) Slice() (result []string) {
	for k := range set.aMap {
		result = append(result, k)
	}
	return
}
func (set *StringSet) Size() int {
	return len(set.aMap)
}
