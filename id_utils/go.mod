module github.com/coldstar-507/utils/id_utils

go 1.23.2

require (
	github.com/coldstar-507/flatgen v0.0.0-20240830172816-703a5c6098f5
	github.com/coldstar-507/utils/utils v0.0.0
)

require (
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	go.mongodb.org/mongo-driver v1.17.1 // indirect
)

replace (
	github.com/coldstar-507/flatgen => ../../../flatbufs/flatgen
	github.com/coldstar-507/utils/utils => ../utils
)
