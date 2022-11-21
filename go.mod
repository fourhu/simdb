module github.com/fourhu/simdb

go 1.19

require github.com/fourhu/elemental v0.0.0-20221028093111-84dd72293f84

require (
	github.com/araddon/dateparse v0.0.0-20200409225146-d820a6159ab1 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
)

replace (
	github.com/sonyarouje/simdb => github.com/fourhu/simdb v0.0.0-20221121041954-9c5818c56e85
	go.aporeto.io/elemental => github.com/fourhu/elemental v0.0.0-20220831102436-5e4f85e59198
)
