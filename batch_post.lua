-- Initialize random seed
math.randomseed(os.time())

-- Configuration variables
local BATCH_SIZE = 500 -- Number of events sent per single HTTP request

-- Arrays of mock data to cycle through dynamically
local event_types = { "page_view", "click", "add_to_cart", "purchase", "search", "error_log" }
local sources     = { "web", "ios", "android", "backend_worker" }
local user_ids    = { "usr_1001", "usr_2002", "usr_3003", "usr_4004", "usr_5005", "usr_6006" }

local urls        = { "/home", "/pricing", "/features", "/checkout", "/dashboard" }
local buttons     = { "signup_btn", "submit_payment", "close_modal", "apply_coupon" }
local items       = { "premium_subscription", "enterprise_tier", "addon_storage" }

-- Function to generate a single randomized event structure as a string
local function generate_single_event()
    local ev_type = event_types[math.random(#event_types)]
    local src     = sources[math.random(#sources)]
    local usr     = user_ids[math.random(#user_ids)]
    
    -- Dynamically build custom properties based on the event type
    local props = "{}"
    if ev_type == "page_view" then
        props = string.format('{"url":"%s","referrer":"direct"}', urls[math.random(#urls)])
    elseif ev_type == "click" then
        props = string.format('{"element_id":"%s","page":"%s"}', buttons[math.random(#buttons)], urls[math.random(#urls)])
    elseif ev_type == "add_to_cart" or ev_type == "purchase" then
        props = string.format('{"sku":"%s","price":%d,"currency":"USD"}', items[math.random(#items)], math.random(19, 499))
    elseif ev_type == "search" then
        props = '{"query":"bi analytics engine","results_returned":12}'
    elseif ev_type == "error_log" then
        props = '{"code":500,"message":"upstream database timeout connection reset"}'
    end

    -- Construct the JSON object snippet
    return string.format(
        '{"type":"%s","source":"%s","user_id":"%s","properties":%s}',
        ev_type, src, usr, props
    )
end

-- Setup request method and headers once globally
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"

-- The request function runs on every single iteration of the load test loop
request = function()
    -- Create an array buffer of event strings
    local events = {}
    for i = 1, BATCH_SIZE do
        events[i] = generate_single_event()
    end -- <-- FIXED: Changed from '}' to 'end'
    
    -- Join the array components into a single valid JSON root array: [{},{},{}]
    wrk.body = "[" .. table.concat(events, ",") .. "]"
    
    return wrk.format(nil, nil, nil, wrk.body)
end