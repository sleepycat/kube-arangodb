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

package deployment

import (
	"fmt"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

func resourceLimitAsEnv(t *testing.T, requirements core.ResourceRequirements) core.EnvVar {
	value, ok := requirements.Limits[core.ResourceMemory]
	require.True(t, ok)

	return core.EnvVar{
		Name:  resources.ArangoDBOverrideDetectedTotalMemoryEnv,
		Value: fmt.Sprintf("%d", value.Value()),
	}
}

func TestEnsurePod_ArangoDB_Resources(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "DBserver POD with resource requirements",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Resources: resourcesUnfiltered,
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
		{
			Name: "DBserver POD with resource requirements, with override flag",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Resources:                   resourcesUnfiltered,
						OverrideDetectedTotalMemory: util.NewBool(true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Env: []core.EnvVar{
								resourceLimitAsEnv(t, resourcesUnfiltered),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
		{
			Name: "DBserver POD without resource requirements, with override flag",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						OverrideDetectedTotalMemory: util.NewBool(true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
	}

	runTestCases(t, testCases...)
}
