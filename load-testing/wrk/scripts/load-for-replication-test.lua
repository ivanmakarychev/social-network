local urlEncodedCyrillicLetters = {"%D0%B0", "%D0%B1", "%D0%B2", "%D0%B3", "%D0%B4", "%D0%B5", "%D0%B6", "%D0%B7", "%D0%B8", "%D0%BA", "%D0%BB", "%D0%BC", "%D0%BD", "%D0%BE", "%D0%BF", "%D1%80", "%D1%81", "%D1%82", "%D1%83", "%D1%84", "%D1%85"}

local function randomLetter ()
   letter = urlEncodedCyrillicLetters[ math.random( #urlEncodedCyrillicLetters ) ]
   return letter
end

math.randomseed(os.time())

request1 = function()
   name =  randomLetter() .. randomLetter()
   surname = randomLetter() .. randomLetter()
   path = string.format("/profiles?name=%s&surname=%s", name, surname)
   return wrk.format("GET", path)
end

request2 = function()
   path = string.format("/profile?id=%d", math.random(1, 1000000))
   return wrk.format("GET", path)
end

requests = {}
requests[0] = request1
requests[1] = request2

request = function()
    return requests[math.random(0, 1)]()
end