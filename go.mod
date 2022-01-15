module github.com/splitsh/lite

go 1.17

require (
	github.com/boltdb/bolt v1.3.1
	github.com/libgit2/git2go v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/crypto v0.0.0-20201203163018-be400aefbc4c // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
)

replace github.com/libgit2/git2go => ../../libgit2/git2go
