function set_time(tag, timestamp, record)
  local timestamp_field = record["log.timestmap.field"]
  if timestamp_field ~= nil then
    record["time"] = record[timestamp_field] or record["time"] or timestamp
  end

  if type(record["time"]) == "number" then
    record["time"] = os.date("!%Y-%m-%dT%H:%M:%SZ", record["time"])
  end

  record["time"] = string.gsub(record["time"], "Z$", "")
  return 1, timestamp, record
end
