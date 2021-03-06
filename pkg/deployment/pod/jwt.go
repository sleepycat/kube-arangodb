//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Adam Janikowski
//

package pod

import (
	"fmt"
	"path/filepath"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
)

const ActiveJWTKey = "-"

func IsAuthenticated(i Input) bool {
	return i.Deployment.IsAuthenticated()
}

func VersionHasJWTSecretKeyfile(v driver.Version) bool {
	if v.CompareTo("3.3.22") >= 0 && v.CompareTo("3.4.0") < 0 {
		return true
	}
	if v.CompareTo("3.4.2") >= 0 {
		return true
	}

	return false
}

func JWTSecretFolder(name string) string {
	return fmt.Sprintf("%s-jwt-folder", name)
}

func VersionHasJWTSecretKeyfolder(v driver.Version, enterprise bool) bool {
	return features.JWTRotation().Supported(v, enterprise)
}

func JWT() Builder {
	return jwt{}
}

type jwt struct{}

func (e jwt) Envs(i Input) []core.EnvVar {
	if !IsAuthenticated(i) {
		return nil
	}

	if !VersionHasJWTSecretKeyfile(i.Version) {
		return []core.EnvVar{k8sutil.CreateEnvSecretKeySelector(constants.EnvArangodJWTSecret,
			i.Deployment.Authentication.GetJWTSecretName(), constants.SecretKeyToken)}
	}

	return nil
}

func (e jwt) Args(i Input) k8sutil.OptionPairs {
	if !IsAuthenticated(i) {
		// Without authentication
		return k8sutil.NewOptionPair(k8sutil.OptionPair{Key: "--server.authentication", Value: "false"})
	}

	options := k8sutil.CreateOptionPairs(2)

	options.Add("--server.authentication", "true")

	if VersionHasJWTSecretKeyfolder(i.Version, i.Enterprise) {
		options.Add("--server.jwt-secret-folder", k8sutil.ClusterJWTSecretVolumeMountDir)
	} else if VersionHasJWTSecretKeyfile(i.Version) {
		keyPath := filepath.Join(k8sutil.ClusterJWTSecretVolumeMountDir, constants.SecretKeyToken)
		options.Add("--server.jwt-secret-keyfile", keyPath)
	} else {
		options.Addf("--server.jwt-secret", "$(%s)", constants.EnvArangodJWTSecret)
	}

	return options
}

func (e jwt) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	if !IsAuthenticated(i) {
		return nil, nil
	}

	var vol core.Volume
	if VersionHasJWTSecretKeyfolder(i.Version, i.Enterprise) {
		vol = k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, JWTSecretFolder(i.ApiObject.GetName()))
	} else {
		vol = k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, i.Deployment.Authentication.GetJWTSecretName())
	}
	return []core.Volume{vol}, []core.VolumeMount{k8sutil.ClusterJWTVolumeMount()}
}

func (e jwt) Verify(i Input, cachedStatus inspector.Inspector) error {
	if !IsAuthenticated(i) {
		return nil
	}

	if !VersionHasJWTSecretKeyfolder(i.Version, i.Enterprise) {
		secret, exists := cachedStatus.Secret(i.Deployment.Authentication.GetJWTSecretName())
		if !exists {
			return errors.Errorf("Secret for JWT token is missing %s", i.Deployment.Authentication.GetJWTSecretName())
		}

		if err := k8sutil.ValidateTokenFromSecret(secret); err != nil {
			return errors.Wrapf(err, "Cluster JWT secret validation failed")
		}
	}

	return nil
}
