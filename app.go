package main

type App struct {
	windows        []*Window
	visibleWindows []*Window
	currentWindow  *Window
}
