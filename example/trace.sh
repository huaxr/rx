curl http://localhost:9998/debug/pprof/trace\?seconds\=20 > trace.out
go tool trace trace.out