package main

import "os"

func main1() {
	os.Remove("./genesis")
	os.Remove("./block")
	os.Remove("./block.lock")
	os.Remove("./draft.db")
	os.Remove("./wallet.dat")
}
