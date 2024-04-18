package mailtemplates

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/template/mustache/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/utils/fileio"
)

// Subjects
const (
	SubjectVerifyMail = "mytoken notifications - Verify email"
)

// TemplateNames
const (
	TemplateVerifyMail = "verify_mail"
)

//go:embed templates
var _templates embed.FS
var templates fs.FS

var engine *mustache.Engine

func init() {
	var err error
	templates, err = fs.Sub(_templates, "templates")
	if err != nil {
		log.WithError(err).Fatal()
	}
}

// Init initializes the mail templates
func Init() {
	overWriteDir := config.Get().Features.Notifications.Mail.OverwriteDir
	engine = mustache.NewFileSystem(
		fileio.NewLocalAndOtherSearcherFilesystem(overWriteDir, http.FS(templates)),
		".mustache",
	)
	if err := engine.Load(); err != nil {
		log.WithError(err).Fatal()
	}
}

func render(name, suffix string, bindData any) (string, error) {
	var buf bytes.Buffer
	if err := engine.Render(&buf, name+suffix, bindData); err != nil {
		return "", errors.WithStack(err)
	}
	return buf.String(), nil
}

// HTML renders a html-suffix file
func HTML(name string, bindData any) (string, error) {
	return render(name, ".html", bindData)
}

// Text renders a txt-suffix file
func Text(name string, bindData any) (string, error) {
	return render(name, ".txt", bindData)
}
