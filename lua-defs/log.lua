---@meta

--- Structured logging module. Logs plugin activity to the CMS application logger.
--- Plugin name is automatically included as a structured field.
---@class log
log = {}

--- Log an informational message.
---@param message string
---@param context? table<string, any> Structured key-value fields.
function log.info(message, context) end

--- Log a warning message.
---@param message string
---@param context? table<string, any> Structured key-value fields.
function log.warn(message, context) end

--- Log an error message.
---@param message string
---@param context? table<string, any> Structured key-value fields.
function log.error(message, context) end

--- Log a debug message.
---@param message string
---@param context? table<string, any> Structured key-value fields.
function log.debug(message, context) end
