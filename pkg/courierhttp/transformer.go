package courierhttp

import (
	"context"
	"io"
	"net/textproto"

	typesutil "github.com/octohelm/x/types"
)

type Transformer interface {
	// Names name or alias of transformer
	// prefer using some keyword about content-type
	Names() []string

	// New transformer instance by type
	// in this step will to check transformer is valid for type
	New(context.Context, typesutil.Type) (Transformer, error)

	// EncodeTo writer
	EncodeTo(w io.Writer, v any) (mediaType string, err error)

	// DecodeFrom reader
	DecodeFrom(r io.Reader, v any, headers ...textproto.MIMEHeader) error
}
