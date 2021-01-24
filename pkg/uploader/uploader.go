package uploader

import "github.com/jecklgamis/gatling-server/pkg/workspace"

type GatlingArtifactUploader interface {
	Upload(taskId string, userFilesDir *workspace.UserFilesDir) error
}
