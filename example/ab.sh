cat post_data.txt
ab -n 100 -c 10 -p post_data.txt -T 'application/json' http://localhost:9999/bbb
#ab -n 1000 -c 100  http://localhost:9999/aaa