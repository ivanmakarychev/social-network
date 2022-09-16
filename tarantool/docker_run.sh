docker run \
  --name tarantool \
  --network social-net \
  -p 3301:3301 \
  -v /Users/imakarychev/github.com/ivanmakarychev/social-network/tarantool/data:/var/lib/tarantool \
  -it --rm \
  tarantool/tarantool