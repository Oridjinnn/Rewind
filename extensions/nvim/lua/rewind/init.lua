-- Rewind Neovim Plugin
-- Records IDE activity and sends to Rewind server (localhost:9876)

local M = {}

local config = {
  server_port = 9876,
  enabled = false, -- Opt-in: off by default
  record_files = true,
  record_commands = true,
  record_git = true,
  batch = true,
  batch_interval_ms = 5000,
}

local batch = {}
local batch_timer = nil

-- Configuration
function M.setup(opts)
  opts = opts or {}
  for k, v in pairs(opts) do
    config[k] = v
  end
  if config.enabled then
    M.enable()
  end
end

function M.enable()
  config.enabled = true
  vim.notify("Rewind: Recording enabled", vim.log.levels.INFO)
end

function M.disable()
  config.enabled = false
  flush_batch()
  vim.notify("Rewind: Recording paused", vim.log.levels.WARN)
end

function M.toggle()
  if config.enabled then
    M.disable()
  else
    M.enable()
  end
  return config.enabled
end

-- Send event to Rewind server
local function send_event(event_type, extra)
  if not config.enabled then return end
  extra = extra or {}

  local project = vim.fn.getcwd()
  local project_name = vim.fn.fnamemodify(project, ":t")

  local event = {
    protocol = "rewind-ide-v1",
    ide = "nvim",
    version = vim.version() and string.format("%d.%d.%d", vim.version().major, vim.version().minor, vim.version().patch) or "0.0.0",
    project = project_name,
    project_path = project,
    event = {
      type = event_type,
      timestamp = os.date("!%Y-%m-%dT%H:%M:%S"),
    }
  }

  for k, v in pairs(extra) do
    event.event[k] = v
  end

  if config.batch then
    table.insert(batch, event)
    schedule_flush()
  else
    send_immediate(event)
  end
end

local function schedule_flush()
  if batch_timer then return end
  batch_timer = vim.defer_fn(function()
    flush_batch()
  end, config.batch_interval_ms)
end

local function flush_batch()
  if batch_timer then
    pcall(vim.fn.timer_stop, batch_timer)
    batch_timer = nil
  end
  if #batch == 0 then return end
  local events = batch
  batch = {}
  send_immediate(events)
end

local function send_immediate(payload)
  local body = vim.json.encode(payload)
  local curl_cmd = string.format(
    [[curl -s -X POST http://localhost:%d/ -H 'Content-Type: application/json' -d '%s' 2>/dev/null &]],
    config.server_port, body:gsub("'", "'\\''")
  )
  vim.fn.jobstart({"sh", "-c", curl_cmd}, {detach = true})
end

-- Detect file language
local function detect_language(filepath)
  local ext = vim.fn.fnamemodify(filepath, ":e"):lower()
  local map = {
    lua = "lua", py = "python", rb = "ruby", js = "javascript",
    ts = "typescript", go = "go", rs = "rust", java = "java",
    c = "c", cpp = "cpp", h = "c", cs = "csharp",
    php = "php", sql = "sql", html = "html", css = "css",
    json = "json", yaml = "yaml", yml = "yaml", md = "markdown",
    sh = "shell", bash = "shell", toml = "toml"
  }
  return map[ext] or ext
end

-- Autocommands for file events
vim.api.nvim_create_autocmd("BufReadPost", {
  callback = function(args)
    send_event("file_open", {
      file = args.file,
      language = detect_language(args.file),
    })
  end,
})

vim.api.nvim_create_autocmd("BufWritePost", {
  callback = function(args)
    send_event("file_save", {
      file = args.file,
      language = detect_language(args.file),
    })
  end,
})

vim.api.nvim_create_autocmd("BufWipeout", {
  callback = function(args)
    send_event("file_close", {
      file = args.file,
      language = detect_language(args.file),
    })
  end,
})

-- Command-line tracking
vim.api.nvim_create_autocmd("CmdlineLeave", {
  callback = function()
    local cmd = vim.fn.getcmdline()
    if cmd and #cmd > 0 then
      send_event("run_config", {
        message = cmd,
        metadata = {type = "nvim_command"},
      })
    end
  end,
})

-- Git events (basic: on write to git files)
vim.api.nvim_create_autocmd("BufWritePost", {
  pattern = {"gitcommit", "gitrebase"},
  callback = function(args)
    send_event("git_commit", {
      file = args.file,
      message = "Git operation triggered",
    })
  end,
})

-- Vim resize/refocus (lightweight activity pulse)
vim.api.nvim_create_autocmd("VimResized", {
  callback = function()
    send_event("search", {
      message = "Window layout changed",
      metadata = {columns = vim.o.columns, lines = vim.o.lines},
    })
  end,
})

-- User commands for manual control
vim.api.nvim_create_user_command("RewindEnable", M.enable, {desc = "Enable Rewind recording"})
vim.api.nvim_create_user_command("RewindDisable", M.disable, {desc = "Disable Rewind recording"})
vim.api.nvim_create_user_command("RewindToggle", M.toggle, {desc = "Toggle Rewind recording"})

-- Auto-enable if configured
if config.enabled then
  M.enable()
end

return M