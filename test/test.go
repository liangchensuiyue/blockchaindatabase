package test

type Test struct {
	PrintTxInfo func()
	CleanTxInfo func()

	PrintQueryInfo func()
	CleanQueryInfo func()

	PrintBlockInfo func()
	CleanBlockInfo func()
}

var SystemInfo *Test = &Test{}
