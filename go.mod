module github.com/ericzzh/mattermost-plugin-oryauth

go 1.16

require (
	github.com/mattermost/mattermost-server/v6 v6.2.1
	github.com/ory/hydra-client-go v2.0.2+incompatible
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)

replace github.com/ory/hydra-client-go v2.0.2+incompatible => ../hydra-client-go
