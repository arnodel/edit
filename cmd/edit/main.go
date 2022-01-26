package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/arnodel/edit"
)

func main() {
	flag.Parse()
	var buf edit.Buffer
	if flag.NArg() == 0 {
		buf = edit.NewEmptyFileBuffer()
	} else {
		buf = edit.NewBufferFromFile(flag.Arg(0))
	}

	// Log to a file because the terminal is used
	// f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	f, err := ioutil.TempFile("", "edit.logs.")
	if err != nil {
		log.Fatalf("error opening file: %s", err)
	}
	defer f.Close()
	log.Printf("Logging to %s", f.Name())
	log.SetOutput(f)

	// Initialise the app
	win := edit.NewWindow(buf)
	app := edit.NewApp(win)

	// Initialise the screen
	screen, err := edit.NewScreen()
	if err != nil {
		log.Fatalf("error getting screen: %s", err)
	}
	defer screen.Cleanup()

	// Event loop
	for app.Running() {
		app.Draw(screen)
		app.HandleEvent(screen.PollEvent())
	}
}
