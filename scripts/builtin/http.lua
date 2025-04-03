local host = arg[1]
if not host then
  print("Usage: lua http.lua <host>")
  os.exit(0)
end
local http = require("socket.http")
local res, code = http.request(host)
if code == 200 then
  print("success")
  os.exit(0)
else
  print("fail")
  os.exit(1)
end


