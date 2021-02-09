#cat post_data.txt
#ab -n 100 -c 10 -p post_data.txt -T 'application/json' http://localhost:9999/bbb
#ab -n 1000 -c 100  http://localhost:9999/v1/v2/v3/eee


#wrk -c 400 -t 8 -d 3m http://localhost:9999/test

#brew install graphviz
go tool pprof ./output/bin/test ./profile