package main

// checking how defer works
func doWork(log *[]string) {
	*log = append(*log, "start")
	defer func() { *log = append(*log, "cleanup A") }()
	defer func() { *log = append(*log, "cleanup B") }()
	*log = append(*log, "end")
}
