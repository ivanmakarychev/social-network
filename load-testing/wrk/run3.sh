wrk -c 10 -t 2 -d 10m --timeout 1m -s ./scripts/find-profiles-by-name-and-surname.lua --latency http://localhost > 10
