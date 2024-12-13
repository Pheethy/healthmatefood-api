package file

import (
	"context"
	"healthmatefood-api/models"
)

type IFileUsecase interface {
	UploadToGCP(ctx context.Context, fileReq []*models.FileReq) ([]*models.FileResp, error)
	DeleteOnGCP(req []*models.DeleteFileReq) error
}
