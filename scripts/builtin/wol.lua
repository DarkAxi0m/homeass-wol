local mac = arg[1]
if not mac then
  print("Usage: lua wake.lua <MAC_ADDRESS>")
  os.exit(1)
end

local handle = io.popen("wakeonlan " .. mac .. " 2>&1")
local result = handle:read("*a")
local success, _, code = handle:close()

if success then
  print("success")
  os.exit(0)
else
  print("fail: " .. result)
  os.exit(code or 1)
end

