package pool

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Value int
	Name  string
	Items []int
}

func (t *TestStruct) Reset() {
	t.Value = 0
	t.Name = ""
	t.Items = t.Items[:0]
}

func NewTestStruct() *TestStruct {
	return &TestStruct{
		Value: 42,
		Name:  "test",
		Items: make([]int, 0, 10),
	}
}

func TestPool(t *testing.T) {
	p := New(NewTestStruct)

	obj1 := p.Get()
	obj1.Value = 100
	obj1.Name = "modified"
	obj1.Items = append(obj1.Items, 1, 2, 3)

	p.Put(obj1)

	obj2 := p.Get()

	assert.Equal(t, 0, obj2.Value)
	assert.Equal(t, "", obj2.Name)
	assert.Equal(t, 0, len(obj2.Items))
	assert.Equal(t, 10, cap(obj2.Items))
}

func TestPoolConcurrent(t *testing.T) {
	p := New(NewTestStruct)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj := p.Get()
			obj.Value = 999
			obj.Name = "concurrent"
			p.Put(obj)
		}()
	}
	wg.Wait()

	obj := p.Get()
	assert.Equal(t, 0, obj.Value)
	assert.Equal(t, "", obj.Name)
}
