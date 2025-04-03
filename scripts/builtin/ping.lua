local host = arg[1]
if not host then
  print("Usage: lua ping.lua <host>")
  os.exit(1)
end

local handle = io.popen("ping -c 1 " .. host)
local result = handle:read("*a")
handle:close()

if result:match("1 received") then
  print("success")
  os.exit(0)
else
  print("fail")
  os.exit(1)
end

