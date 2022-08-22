module gworld/git/GoFrameWork/services/AASManager

go 1.14

require (
	github.com/buaazp/fasthttprouter v0.1.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/consul/api v1.5.0 // indirect
	github.com/inconshreveable/log15 v0.0.0-20201112154412-8562bdadbbac // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/uber/jaeger-client-go v2.29.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/valyala/fasthttp v1.26.0 // indirect
	github.com/xiaomi-tc/log15 v0.0.0-20191113113727-7f66c2abf493 // indirect
	golang.org/x/net v0.0.0-20210510120150-4163338589ed // indirect
	google.golang.org/grpc v1.38.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gworld/git/GoFrameWork/common/bjwt v0.0.0-00010101000000-000000000000 // indirect
	gworld/git/GoFrameWork/common/new_log15 v0.0.0-00010101000000-000000000000 // indirect
	gworld/git/GoFrameWork/common/service_util v0.0.0-00010101000000-000000000000
	gworld/git/GoFrameWork/common/utils/grpcTracing v0.0.0-00010101000000-000000000000
	gworld/git/GoFrameWork/pb/AASMessage v0.0.0-00010101000000-000000000000
	gworld/git/GoFrameWork/pb/ComMessage v0.0.0-00010101000000-000000000000 // indirect
	gworld/git/GoFrameWork/pb/NodeAgent v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	gworld/git/GoFrameWork/common/bjwt => ../../../GoFrameWork/common/bjwt
	gworld/git/GoFrameWork/common/new_log15 => ../../../GoFrameWork/common/new_log15
	gworld/git/GoFrameWork/common/service_util => ../../../GoFrameWork/common/service_util
	gworld/git/GoFrameWork/common/utils/grpcTracing => ../../../GoFrameWork/common/utils/grpcTracing
	gworld/git/GoFrameWork/pb/AASMessage => ../../../GoFrameWork/pb/AASMessage
	gworld/git/GoFrameWork/pb/ComMessage => ../../../GoFrameWork/pb/ComMessage
	gworld/git/GoFrameWork/pb/NodeAgent => ../../../GoFrameWork/pb/NodeAgent
	gworld/git/third-package/satori/go.uuid => ../../../third-package/satori/go.uuid
)
