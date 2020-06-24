package main

import "log"

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func checkErrExit(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
