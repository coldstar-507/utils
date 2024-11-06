module github.com/coldstar-507/utils

go 1.22.1

require (
	github.com/btcsuite/btcd/btcutil v1.1.5
	github.com/coldstar-507/flatgen v0.0.0-20240721154545-7f7a3c686f6f
	go.mongodb.org/mongo-driver v1.16.1
)

replace github.com/coldstar-507/flatgen => /home/scott/dev/down4/flatbufs/flatgen

require github.com/google/flatbuffers v24.3.25+incompatible // indirect
