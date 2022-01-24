package main

import (
	"flag"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	// "github.com/sourcegraph/go-lsp"
	// "github.com/sourcegraph/jsonrpc2"
	// "golang.org/x/tools/internal/lsp/protocol"
)

// That was messing around with gopls.  It's here in case I want to try that
// again
// func tryGopls() {
// 	gls := exec.Command("gopls", "-v", "-logfile", "gopls.log", "-listen", "localhost:7463")
// 	if err := gls.Start(); err != nil {
// 		log.Fatal("Could not start language server: ", err)
// 	}
// 	time.Sleep(time.Second / 10)
// 	conn, err := net.Dial("tcp", "localhost:7463")
// 	if err != nil {
// 		log.Fatal("Could not connect to language server: ", err)
// 	}
// 	codec := jsonrpc2.VSCodeObjectCodec{}
// 	stream := jsonrpc2.NewBufferedStream(conn, codec)
// 	handler := jsonrpc2.HandlerWithError(func(context.Context, *jsonrpc2.Conn, *jsonrpc2.Request) (result interface{}, err error) {
// 		return nil, nil
// 	})
// 	ctx := context.Background()
// 	rpcConn := jsonrpc2.NewConn(ctx, stream, handler, jsonrpc2.LogMessages(log.New(os.Stderr, "", log.LstdFlags)))
// 	initResult := lsp.InitializeResult{}
// 	log.Print("Calling initialize")
// 	err = rpcConn.Call(ctx, "initialize", &lsp.InitializeParams{
// 		RootURI: "file:///Users/adelobelle/personal/codenav",
// 	}, &initResult)
// 	if err != nil {
// 		log.Fatal("Error: ", err)
// 	}
// 	log.Print("Result: ", initResult)
// }

func main() {
	flag.Parse()
	filename := flag.Arg(0)

	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("This is a test log entry")

	tcellScreen, err := tcell.NewScreen()
	if err != nil {
		log.Print("cannot get tcellScreen: ", err)
		return
	}
	if err := tcellScreen.Init(); err != nil {
		log.Fatal("cannot init tcellScreen: ", err)
	}
	defer tcellScreen.Fini()
	tcellScreen.EnableMouse()

	screen := &Screen{tcellScreen: tcellScreen}

	evtConverter := TcellEventConverter{}

	app := NewApp(filename)
	app.Init()

	for {
		screen.Fill(' ')
		app.Draw(screen)
		tcellScreen.Show()
		tcevt := tcellScreen.PollEvent()
		keyEvt, ok := tcevt.(*tcell.EventKey)
		if ok && keyEvt.Key() == tcell.KeyCtrlC {
			return
		}
		app.HandleEvent(evtConverter.EventFromTcell(tcevt))
	}
}
