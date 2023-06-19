# protopool

[![GoDoc](https://godoc.org/github.com/jussi-kalliokoski/protopool?status.svg)](https://godoc.org/github.com/jussi-kalliokoski/protopool)
[![CI status](https://github.com/jussi-kalliokoski/protopool/workflows/CI/badge.svg)](https://github.com/jussi-kalliokoski/protopool/actions)

A go library for pooling allocations that are only used for marshaling right after creation, e.g. protobuf.

## Usage

### Struct Pointers

Pointers to structs can be recycled:

```go
type struct MyStruct{ Value: int }
var v protopool.Value
s1 := protopool.New[MyStruct](&v) // *MyStruct
s1.Value = 123
s2 := protopool.New[MyStruct](&v) // *MyStruct
s2.Value = 234
v.Reset()
s3 := protopool.New[MyStruct](&v) // points to the same address as s1, but values are zeroed
s4 := protopool.New[MyStruct](&v) // points to the same address as s2, but values are zeroed
```

### Slices

Slices can be recycled:

```go
var v protopool.Value
var s1 []int
l1 := protopool.NewList(&v, &s1)
l1.Append(1, 2) // s1 will be []int{ 1, 2 }
l2 := protopool.NewList(&v, &s2)
l2.Append(3, 4) // s2 will be []int{ 3, 4 }
v.Reset()
l3 := protopool.NewList(&v, &s1)
l3.Append(5, 6) // s3 will be []int{ 5, 6 }, but using the same underlying allocation as s1
l4 := protopool.NewList(&v, &s2)
l4.Append(7, 8) // s4 will be []int{ 7, 8 }, but using the same underlying allocation as s2
```

### Maps

Maps can be recycled:

```go
var v protopool.Value
var m1 map[int32]int64
protopool.Put(&v, &m1, 1, 2) // m1 will be { 1: 2 }
var m2 map[int32]int64
protopool.Put(&v, &m2, 3, 4) // m2 will be { 3, 4 }
v.Reset()
var m3 map[int32]int64
protopool.Put(&v, &m3, 5, 6) // m3 will be { 5: 6 }, but using the same underlying map as m1
var m4 map[int32]int64
protopool.Put(&v, &m4, 7, 8) // m4 will be { 7: 8 }, but using the same underlying map as m2
```


## Memory Use Considerations

Given that the structures can often contain slices or strings that are sub-sliced from other, larger, temporarily allocated values, and that otherwise the memory use of any single usage point will be climbing towards the P100 memory usage of a single usage point, it's a good idea to limit the reuse of the values with staggering, e.g.

```go
struct MyPoolValue struct {
    ReuseRemaining int
    protos protopool.Value
}

const minUses = 5000
const maxUses = minUses + 1000

var p sync.Pool

doSomething := func() {
    pv, _ := p.Get().(*MyPoolValue)
    if pv == nil || pv.ReuseRemaining <= 0 {
        pv = &MyPoolValue{
            ReuseRemaining: minUses + rand.Intn(maxUses - minUses),
        }
    } else {
        pv.ReuseRemaining--
        pv.protos.Reset()
    }
    defer p.Put(pv)

    // ... construct objects ...
}
```

## Performance

As with all things performance, measure before and after. As a general rule of thumb, if your struct is 4 words or less, go's default allocation will probably work better than protopool, and will also improve in the future. You can see this in effect by adding back the commented out lines that call to the `buildUptimeCounters` in the benchmark and see the effect for yourself. With that in mind, `protopool` can reduce the allocations in construction of large data structures to near zero, while improving the CPU usage significantly:

```
goos: darwin
goarch: amd64
pkg: github.com/jussi-kalliokoski/protopool
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
Benchmark/without_pool-16                   8732           4187300 ns/op         3273642 B/op      21112 allocs/op
Benchmark/with_pool-16                     10000           3576405 ns/op             518 B/op          6 allocs/op
```
