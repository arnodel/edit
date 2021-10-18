return function(app)

    app.NewCommand{
        Name: "insert-rune",
        Args: {
            {
                Name = "Rune",
                Type = "rune"
            }
        }
        Action: function(win, rune) win.InsertRune(rune) end
    }
    app.BindEvents{
        Seq = "Ctrl-X Ctrl-X",
        Action = function() app.Log("Hello from Lua!!!") end
    }
    app.BindEvents{
        Seq = "Ctrl-B",
        Action = function(win) win.MoveCursor(-1, 0) end
    }
    app.BindEvents{
        Seq = "Ctrl-X 2 Rune.Rune",
        Action = function(win, rune) win.InsertRune(rune) win.InsertRune(rune) end
    }
    app.BindEvents{
        Seq = "Ctrl-X Ctrl-N", 
        Action = function() app.SwitchWindow() end
    }
end
