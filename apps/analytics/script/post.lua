wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"

wrk.body = '{"type":"page_view","source":"web","properties":{"url":"/home"}}'