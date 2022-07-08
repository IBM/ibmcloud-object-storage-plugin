module github.com/IBM/ibmcloud-object-storage-plugin

go 1.18

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/IBM/go-sdk-core/v3 v3.3.1
	github.com/IBM/ibm-cos-sdk-go v1.7.0
	github.com/IBM/ibm-cos-sdk-go-config v1.2.0
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/jessevdk/go-flags v1.5.0
	github.com/miekg/dns v1.1.43 // indirect
	github.com/pierrre/gotestcover v0.0.0-20160517101806-924dca7d15f0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.1
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/sig-storage-lib-external-provisioner/v6 v6.3.0
)

replace (
	// workaround to replace dgrijalva/jwt-go with github.com/golang-jwt/jwt/v4 in indirect dependiencies
	// PSIRT PVR0322500 replace github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible => github.com/golang-jwt/jwt/v4 v4.2.0
	sigs.k8s.io/sig-storage-lib-external-provisioner v4.1.0+incompatible => sigs.k8s.io/sig-storage-lib-external-provisioner/v6 v6.0.0
)
