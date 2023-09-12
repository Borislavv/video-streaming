package dto

import (
	"github.com/Borislavv/video-streaming/internal/domain/vo"
	"mime/multipart"
)

type CreateRequest interface {
	GetName() string
	GetFilepath() string
	GetResourceID() vo.ID
	GetDescription() string
}

type UpdateRequest interface {
	GetId() vo.ID
	GetName() string
	GetDescription() string
}

type DeleteRequest GetRequest
type GetRequest interface {
	GetId() vo.ID
}

type ListRequest interface {
	GetName() string     // path of name
	GetFilepath() string // path of path
	PaginatedRequest
}

type UploadRequest interface {
	GetFile() multipart.File
	GetHeader() *multipart.FileHeader
	GetUploadedFilename() string
	SetUploadedFilename(filename string)
	GetUploadedFilepath() string
	SetUploadedFilepath(filepath string)
}

type PaginatedRequest interface {
	GetPage() int
	GetLimit() int
}

type Chunked interface {
	GetLen() int
	SetLen(len int)
	GetData() []byte
	SetData(data []byte)
	GetError() error
	SetError(err error)
}

type Resource interface {
	GetFilepath() string
	//GetName() string
	//GetFilename() string
	//GetFilesize() int64
	//GetFileMIME() textproto.MIMEHeader
}
