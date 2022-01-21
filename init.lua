return function(app)
    app.Log("Starting init.lua")
    app.BindEvents("Ctrl-X Ctrl-X", function() app.Log("Hello from Lua!!!") end)
    app.BindEvents("Ctrl-B", function(win) win.MoveCursor(-1, 0) end)
    app.BindEvents("Ctrl-X 2 Rune.Rune", function(win, rune) win.InsertRune(rune) win.InsertRune(rune) end)
    app.BindEvents("Ctrl-X Ctrl-N", function() app.SwitchWindow() end)
end
