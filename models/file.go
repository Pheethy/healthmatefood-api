package models

import (
	"context"
	"fmt"
	"healthmatefood-api/constants"
	"mime/multipart"

	"cloud.google.com/go/storage"
)

type FileReq struct {
	File        *multipart.FileHeader `form:"file"`
	Destination string                `form:"destination"` /* ที่อยู่ File (ปลายทาง)*/
	Extension   string                `form:"extension"`   /* นามสกุล File */
	FileName    string                `form:"file_name"`   /* ชื่อ File */
}

type FileResp struct {
	FileName string `json:"file_name"` /* ชื่อ File */
	Url      string `json:"url"`       /* url ของรูปภาพ */
}

type FilesResp []*FileResp

func (f FilesResp) GetImagesFromFilesResp(user *User) []*Image {
	images := make([]*Image, 0)
	for index := range f {
		image := new(Image)
		image.NewUUID()
		image.FileName = f[index].FileName
		image.URL = f[index].Url
		image.RefId = user.Id
		image.RefType = constants.REF_TYPE_USER
		image.SetCreatedAt()
		image.SetUpdatedAt()
		images = append(images, image)
	}
	return images
}

type FilePub struct {
	Bucket      string
	Destination string
	File        *FileResp
}

type DeleteFileReq struct {
	Destination string `json:"destination"` /* ที่อยู่สำหรับลบ File */
}

// makePublic gives all users read access to an object.
func (f *FilePub) MakePublic(ctx context.Context, client *storage.Client) error {
	acl := client.Bucket(f.Bucket).Object(f.Destination).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return fmt.Errorf("ACLHandle.Set: %w", err)
	}
	fmt.Printf("Blob %v is now publicly accessible.\n", f.Destination)
	return nil
}
