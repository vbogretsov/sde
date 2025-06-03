function wrap(tag, timestamp, record)
  return 1, timestamp, { tag = tag, time = record["time"], log = record }
end
