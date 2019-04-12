package main

func main() {
	new(App).Action(":port", func(c *Context) {
	}).Run()
}
