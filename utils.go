package main

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func unique(e []string) []string {
	r := []string{}

	for _, s := range e {
		if !contains(r[:], s) {
			r = append(r, s)
		}
	}
	return r
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func counter(arr []string) map[string]int {
	var count = make(map[string]int)
	for _, x := range arr {
		count[x]++
	}
	return count
}
