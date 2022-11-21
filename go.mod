module github.com/fourhu/simdb

go 1.19

require go.aporeto.io/elemental v1.100.1-0.20210910205400-851949ed821f

replace (
	github.com/sonyarouje/simdb => github.com/fourhu/simdb v0.0.0-20221121050850-5b373537adec
	go.aporeto.io/elemental => github.com/fourhu/elemental v0.0.0-20220831102436-5e4f85e59198
)

require (
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
)

require (
	github.com/araddon/dateparse v0.0.0-20200409225146-d820a6159ab1 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
)
