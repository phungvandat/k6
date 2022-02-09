dev:
	go run *.go

test_get:
	METHOD=get k6 run script.js

test_post:
	METHOD=post k6 run script.js

test_post_sync:
	METHOD=post ENDPOINT=sync k6 run script.js

test_post_async: 
	METHOD=post ENDPOINT=async k6 run script.js

test_post_batch:
	METHOD=post ENDPOINT=batch k6 run script.js