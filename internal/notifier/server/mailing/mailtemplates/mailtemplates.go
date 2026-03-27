package mailtemplates

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	texttemplate "text/template"

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

var htmlTemplates *template.Template
var textTemplates *texttemplate.Template
var templateFS http.FileSystem

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
	templateFS = fileio.NewLocalAndOtherSearcherFilesystem(overWriteDir, http.FS(templates))

	htmlTemplates = template.New("")
	textTemplates = texttemplate.New("")

	if err := loadTemplates(); err != nil {
		log.WithError(err).Fatal()
	}
}

func loadTemplates() error {
	// Get the list of template files from the embedded filesystem
	entries, err := fs.ReadDir(templates, ".")
	if err != nil {
		return errors.WithStack(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".gotmpl") {
			continue
		}

		// Try to read from the filesystem (which checks overwrite dir first)
		content, err := readFile(templateFS, name)
		if err != nil {
			return errors.Wrapf(err, "failed to read template %s", name)
		}

		// Determine the template name (without extension)
		tmplName := strings.TrimSuffix(name, ".gotmpl")

		if strings.HasSuffix(tmplName, ".html") {
			_, err = htmlTemplates.New(tmplName).Parse(string(content))
		} else if strings.HasSuffix(tmplName, ".txt") {
			_, err = textTemplates.New(tmplName).Parse(string(content))
		}
		if err != nil {
			return errors.Wrapf(err, "failed to parse template %s", name)
		}
	}

	return nil
}

func readFile(fs http.FileSystem, name string) ([]byte, error) {
	f, err := fs.Open(filepath.Join("/", name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func render(name, suffix string, bindData any) (string, error) {
	var buf bytes.Buffer
	tmplName := name + suffix
	if suffix == ".html" {
		tmpl := htmlTemplates.Lookup(tmplName)
		if tmpl == nil {
			return "", errors.Errorf("template %s not found", tmplName)
		}
		if err := tmpl.Execute(&buf, bindData); err != nil {
			return "", errors.WithStack(err)
		}
	} else {
		tmpl := textTemplates.Lookup(tmplName)
		if tmpl == nil {
			return "", errors.Errorf("template %s not found", tmplName)
		}
		if err := tmpl.Execute(&buf, bindData); err != nil {
			return "", errors.WithStack(err)
		}
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
