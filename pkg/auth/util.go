package auth

import (
	harvesterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/harvester/harvester/pkg/settings"
)

func IsRancherAuthMode() bool {
	return settings.AuthenticationMode.Get() == string(harvesterv1.Rancher)
}
