package skiplist

import "github.com/shopspring/decimal"

func NewInt64Map() *SkipList {
	return NewCustomMap(func(l, r interface{}) bool {
		return l.(int64) < r.(int64)
	})
}
func NewInt64MapReverse() *SkipList {
	return NewCustomMap(func(l, r interface{}) bool {
		return l.(int64) > r.(int64)
	})
}
func NewDecimalMap() *SkipList {
	return NewCustomMap(func(l, r interface{}) bool {
		return l.(decimal.Decimal).LessThan(r.(decimal.Decimal))
	})
}
func NewDecimalMapReverse() *SkipList {
	return NewCustomMap(func(l, r interface{}) bool {
		return l.(decimal.Decimal).GreaterThan(r.(decimal.Decimal))
	})
}
