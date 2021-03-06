module github.com/brigadecore/brigade/v2

go 1.15

// replace github.com/brigadecore/brigade/sdk/v2 => github.com/krancour/brigade/sdk/v2 v2.0.0-20200923022431-8fd92cf0f5f6

require (
	github.com/AlecAivazis/survey/v2 v2.0.7
	github.com/Azure/go-amqp v0.12.7
	github.com/brigadecore/brigade/sdk/v2 v2.0.0-20200923171232-9f56c474d8bf
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/fatih/color v1.9.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/mux v1.7.4
	github.com/gosuri/uitable v0.0.4
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/rs/cors v1.7.0
	github.com/satori/go.uuid v1.2.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.2.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.mongodb.org/mongo-driver v1.3.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	golang.org/x/net v0.0.0-20200219183655-46282727080f
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
)
