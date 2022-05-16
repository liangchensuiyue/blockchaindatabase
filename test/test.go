package test

type Test struct {
	PrintTxInfo func()
	CleanTxInfo func()

	PrintBlockInfo func()
	CleanBlockInfo func()
}

var SystemInfo *Test = &Test{}
