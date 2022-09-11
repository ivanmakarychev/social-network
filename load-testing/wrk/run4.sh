wrk -c 10 -t 2 -d 10m --timeout 1m -s ./scripts/monitoring-test.lua -H "Authorization: Basic MzoxMjM0NTY3OA==" --latency http://localhost:8081 > 10
