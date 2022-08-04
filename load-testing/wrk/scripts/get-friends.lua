math.randomseed(os.time())

request = function()
   path = string.format("/friends?profile_id=%d", math.random(1, 70000))
   return wrk.format("GET", path)
end
