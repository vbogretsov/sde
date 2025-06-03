local token = nil
local num_users = nil

function init(args)
  num_users = args[1] or 100000
end

function request()
  local headers = {
    ["Content-Type"] = "application/json",
    ["Accept"] = "application/json",
  }
  local user_id = string.format("u-%d", math.random(1, num_users))
  local lat = -90 + math.random() * 180
  local lng = -180 + math.random() * 360
  local payload = string.format('{"uid": "%s", "loc": [%s, %s]}', user_id, lat, lng)
  return wrk.format("POST", wrk.path, headers, payload)
end
