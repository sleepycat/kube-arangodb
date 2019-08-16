//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package backup

import (
	"testing"

	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
)

func Test_State_Scheduled_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, database.ArangoBackupStateScheduled)
}

func Test_State_Scheduled_Download(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateScheduled)

	obj.Spec.Download = &database.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: database.ArangoBackupSpecOperation{
			RepositoryURL: "Some URL",
		},
		ID: "id",
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, database.ArangoBackupStateDownload)

	require.False(t, newObj.Status.Available)
}

func Test_State_Scheduled_Create(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateScheduled)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, database.ArangoBackupStateCreate)

	require.False(t, newObj.Status.Available)
}

func Test_State_Scheduled_Upload(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateScheduled)

	obj.Spec.Upload = &database.ArangoBackupSpecOperation{
		RepositoryURL: "test",
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, database.ArangoBackupStateCreate)

	require.False(t, newObj.Status.Available)
}
