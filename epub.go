package url2epub

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"text/template"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/html"

	"go.yhsif.com/url2epub/ziputil"
)

// EpubMimeType is the mime type for epub.
const EpubMimeType = `application/epub+zip`

const (
	contentTypePeekSize = 512

	epubMimetypeFilename = `mimetype`

	epubContainerFilename = `META-INF/container.xml`
	epubContainerContent  = `<?xml version="1.0"?>
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container" version="1.0">
 <rootfiles>
  <rootfile full-path="` + epubOpfFullpath + `" media-type="application/oebps-package+xml"/>
 </rootfiles>
</container>
`

	epubContentDir      = "content"
	epubArticleFilename = "article.xhtml"
	epubNavFilename     = "nav.xhtml"
	epubOpfFullpath     = epubContentDir + "/content.opf"
)

var (
	epubOpfTmpl = template.Must(template.New("opf").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" xmlns:opf="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="BookID">
 <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
  <dc:identifier id="BookID">{{.ID}}</dc:identifier>
  <dc:title>{{.Title}}</dc:title>
  {{if .Lang -}}
	<dc:language>{{.Lang}}</dc:language>
	{{- end}}
  <meta property="dcterms:modified">{{.Time}}</meta>
 </metadata>
 <manifest>
  <item id="nav" href="{{.NavPath}}" media-type="application/xhtml+xml" properties="nav"/>
  <item id="{{.ArticlePath}}" href="{{.ArticlePath}}" media-type="application/xhtml+xml"/>
  {{range $path, $type := .Images}}
  <item id="{{$path}}" href="{{$path}}" media-type="{{$type}}"/>
	{{- end}}
 </manifest>
 <spine>
  <itemref idref="{{.ArticlePath}}"/>
 </spine>
</package>
`))

	epubNavTmpl = template.Must(template.New("nav").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <title>{{.Title}}</title>
  <meta http-equiv="default-style" content="text/html; charset=utf-8"></meta>
 </head>
 <body>
  <nav xmlns:epub="http://www.idpf.org/2007/ops" epub:type="toc">
   <h2>Contents</h2>
   <ol epub:type="list">
    <li><a href="{{.ArticlePath}}">Content</a></li>
   </ol>
  </nav>
 </body>
</html>
`))
)

type epubOpfData struct {
	ID          string
	Title       string
	Lang        string
	Time        string
	ArticlePath string
	NavPath     string
	Images      map[string]string
}

// EpubArgs defines the args used by Epub function.
type EpubArgs struct {
	// The destination to write the epub content to.
	Dest io.Writer

	// The title of the epub.
	Title string

	// The node pointing to the html tag.
	Node *html.Node

	// Images map:
	// key: image local filename
	// value: image content
	Images map[string]io.Reader
}

// Epub creates an Epub 3.0 file from given content.
func Epub(args EpubArgs) (id string, err error) {
	randomID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("epub: unable to generate uuid: %w", err)
	}

	z := zip.NewWriter(args.Dest)
	defer func() {
		if closeErr := z.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close error: %w", closeErr))
		}
	}()

	// mimetype must be the first file in the zip,
	// and must use Store instead of Deflate.
	if err := ziputil.StoreFile(z, epubMimetypeFilename, ziputil.StringWriterTo(EpubMimeType)); err != nil {
		return "", err
	}

	if err := ziputil.WriteFile(z, epubContainerFilename, ziputil.StringWriterTo(epubContainerContent)); err != nil {
		return "", err
	}

	if err := ziputil.WriteFile(
		z,
		path.Join(epubContentDir, epubArticleFilename),
		ziputil.WriterToWrapper(func(w io.Writer) (int64, error) {
			// NOTE: this does not return the correct n, but it's good enough for our
			// use case.
			return 0, html.Render(w, args.Node)
		}),
	); err != nil {
		return "", err
	}

	imageContentTypes := make(map[string]string, len(args.Images))
	for f, reader := range args.Images {
		if err := func() (err error) {
			filename := path.Join(epubContentDir, f)
			if readCloser, ok := reader.(io.ReadCloser); ok {
				defer DrainAndClose(readCloser)
			}
			var buf []byte
			if buffer, ok := reader.(*bytes.Buffer); ok {
				buf = buffer.Bytes()
			} else {
				r := bufio.NewReader(reader)
				var peekErr error
				buf, peekErr = r.Peek(contentTypePeekSize)
				if peekErr != nil && peekErr != io.EOF {
					err = fmt.Errorf("epub: unable to detect content type for %q: %w", filename, peekErr)
					return
				}
				reader = r
			}
			imageContentTypes[f] = http.DetectContentType(buf)

			return ziputil.WriteFile(
				z,
				filename,
				ziputil.WriterToWrapper(func(w io.Writer) (int64, error) {
					return io.Copy(w, reader)
				}),
			)
		}(); err != nil {
			return "", err
		}
	}

	data := epubOpfData{
		ID:          html.EscapeString(id),
		Title:       html.EscapeString(args.Title),
		Lang:        html.EscapeString(FromNode(args.Node).GetLang()),
		Time:        time.Now().UTC().Format(time.RFC3339),
		ArticlePath: epubArticleFilename,
		NavPath:     epubNavFilename,
		Images:      imageContentTypes,
	}
	if err := ziputil.WriteFile(
		z,
		path.Join(epubContentDir, epubNavFilename),
		ziputil.WriterToWrapper(func(w io.Writer) (int64, error) {
			// NOTE: this does not return the correct n, but it's good enough for our
			// use case.
			return 0, epubNavTmpl.Execute(w, data)
		}),
	); err != nil {
		return "", err
	}

	id = randomID.String()
	if err := ziputil.WriteFile(
		z,
		epubOpfFullpath,
		ziputil.WriterToWrapper(func(w io.Writer) (int64, error) {
			// NOTE: this does not return the correct n, but it's good enough for our
			// use case.
			return 0, epubOpfTmpl.Execute(w, data)
		}),
	); err != nil {
		return "", err
	}

	return id, nil
}
