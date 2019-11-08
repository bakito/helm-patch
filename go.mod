module github.com/bakito/helm-patch

go 1.13

require (
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.2.4
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.0.0-rc.3
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
)

replace github.com/docker/docker => github.com/docker/docker v0.0.0-20190731150326-928381b2215c
