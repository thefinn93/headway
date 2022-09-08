package main

const base = "cmd/headway-build"

func main() {
	if err := downloadMetros(); err != nil {
		panic(err)
	}

}
