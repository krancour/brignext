module github.com/krancour/brignext

go 1.13

replace github.com/deis/async => github.com/krancour/async v1.1.1-0.20200219202113-438f9062d901

require (
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/deis/async v0.0.0-00010101000000-000000000000
	github.com/fatih/color v1.9.0 // indirect
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/mux v1.7.4
	github.com/gosuri/uitable v0.0.4
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/stretchr/testify v1.5.0
	github.com/urfave/cli v1.22.2
	github.com/xeipuuv/gojsonschema v1.2.0
	go.mongodb.org/mongo-driver v1.3.0
	golang.org/x/net v0.0.0-20200219183655-46282727080f
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/utils v0.0.0-20200124190032-861946025e34 // indirect
)
