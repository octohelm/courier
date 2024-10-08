package internal

import "net/http"

type HeaderGetter interface {
	Header() http.Header
}

type ContentTypeGetter interface {
	ContentType() string
}

type ContentTypeSetter interface {
	SetContentType(contentType string)
}

type FilenameGetter interface {
	Filename() string
}

type FilenameSetter interface {
	SetFilename(f string)
}
