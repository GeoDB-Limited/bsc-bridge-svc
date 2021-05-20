module github.com/bsc-bridge-svc

go 1.16

require (
	github.com/GeoDB-Limited/odin-core v1.0.1-0.20210517140735-157e64d7c1ff
	github.com/Masterminds/squirrel v1.5.0
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/cosmos/go-bip39 v1.0.0
	github.com/ethereum/go-ethereum v1.10.3
	github.com/go-chi/chi v1.5.4
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/lib/pq v1.10.0
	github.com/pkg/errors v0.9.1
	github.com/rubenv/sql-migrate v0.0.0-20210408115534-a32ed26c37ea
	github.com/sirupsen/logrus v1.8.1
	google.golang.org/grpc v1.36.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
