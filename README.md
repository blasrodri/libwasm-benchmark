# CosmWasmVM benchmark 1VM vs 2VMs

How to run it?
```bash
make benchmark
```

## Results

```
RUST_BACKTRACE=1 go test -v -bench=.
goos: linux
goarch: amd64
pkg: github.com/CosmWasm/wasmvm
cpu: AMD Ryzen 9 5900X 12-Core Processor            
BenchmarkHappyPathOneVM
BenchmarkHappyPathOneVM-24             1        2209188026 ns/op
BenchmarkHappyPathTwoVMs
BenchmarkHappyPathTwoVMs-24            1        2295065363 ns/op
PASS
ok      github.com/CosmWasm/wasmvm      4.513s
```