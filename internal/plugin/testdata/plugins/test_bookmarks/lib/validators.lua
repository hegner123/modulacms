local M = {}

--- Check whether a string looks like a valid HTTP/HTTPS URL.
--- Only checks the scheme prefix; no network access or full RFC parse.
function M.is_valid_url(url)
    if type(url) ~= "string" then return false end
    if url == "" then return false end
    return url:sub(1, 7) == "http://" or url:sub(1, 8) == "https://"
end

--- Return a trimmed copy of a string (leading/trailing whitespace removed).
function M.trim(s)
    if type(s) ~= "string" then return s end
    return s:match("^%s*(.-)%s*$")
end

return M
