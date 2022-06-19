wrk -c 1 -t 1 -d 1m --timeout 1m -s ./scripts/find-profiles-by-name-and-surname.lua --latency http://localhost > 1
wrk -c 10 -t 2 -d 1m --timeout 1m -s ./scripts/find-profiles-by-name-and-surname.lua --latency http://localhost > 10
wrk -c 100 -t 2 -d 1m --timeout 1m -s ./scripts/find-profiles-by-name-and-surname.lua --latency http://localhost > 100
wrk -c 1000 -t 2 -d 1m --timeout 1m -s ./scripts/find-profiles-by-name-and-surname.lua --latency http://localhost > 1000