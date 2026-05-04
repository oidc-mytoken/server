// Package spa provides functionality for serving the Svelte Single Page Application
package spa

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/server/apipath"
	"github.com/oidc-mytoken/server/internal/utils/fileio"
)

// The SPA build files are embedded here.
// To use the SPA, build it with: cd frontend && npm run build
// Then copy the build directory here: cp -r frontend/build/* internal/server/spa/dist/
// The dist directory should contain index.html and the _app directory

//go:embed all:dist
var _spaFS embed.FS
var spaFS fs.FS

// Available indicates if the SPA distribution files are available
var Available bool

func init() {
	var err error
	spaFS, err = fs.Sub(_spaFS, "dist")
	if err != nil {
		log.WithError(err).Debug("SPA distribution not embedded")
		Available = false
		return
	}

	// Check if index.html exists to verify the SPA is properly built
	if _, err := fs.Stat(spaFS, "index.html"); err != nil {
		log.Debug("SPA index.html not found, SPA not available")
		Available = false
		return
	}

	Available = true
	log.Info("SPA distribution files are available")
}

// AddRoutes adds the SPA routes to the Fiber app
// This should be called after all API routes are added
func AddRoutes(app fiber.Router) {
	if !Available {
		log.Warn("SPA distribution not available, skipping SPA routes")
		return
	}

	overwriteDir := config.Get().Features.WebInterface.OverwriteDir

	// Serve static assets from the SPA build
	// This includes /_app/*, /favicon.ico, etc.
	app.Use(
		"/_app", filesystem.New(
			filesystem.Config{
				Root: fileio.NewLocalAndOtherSearcherFilesystem(
					fileio.JoinIfFirstNotEmpty(overwriteDir, "_app"),
					http.FS(mustSubFS(spaFS, "_app")),
				),
				MaxAge: 31536000, // 1 year for immutable assets
			},
		),
	)

	log.Info("SPA routes registered")
}

// HandleSPAFallback returns a handler for SPA fallback routing
// This serves index.html for any non-API, non-static routes
func HandleSPAFallback() fiber.Handler {
	if !Available {
		return nil
	}

	return func(ctx *fiber.Ctx) error {
		path := ctx.Path()

		// Don't handle API routes
		if strings.HasPrefix(path, apipath.Prefix) {
			return ctx.Next()
		}

		// Don't handle static asset routes (they're handled by filesystem middleware)
		if strings.HasPrefix(path, "/_app/") ||
			strings.HasPrefix(path, "/static/") {
			return ctx.Next()
		}

		// Check if the client accepts HTML
		if ctx.Accepts(fiber.MIMETextHTML, fiber.MIMETextHTMLCharsetUTF8) == "" {
			return ctx.Next()
		}

		// Serve the SPA index.html for client-side routing
		return serveIndex(ctx)
	}
}

// serveIndex serves the SPA index.html file
func serveIndex(ctx *fiber.Ctx) error {
	overwriteDir := config.Get().Features.WebInterface.OverwriteDir

	// Try to read from overwrite directory first
	if overwriteDir != "" {
		localFS := fileio.NewLocalAndOtherSearcherFilesystem(
			overwriteDir,
			http.FS(spaFS),
		)
		file, err := localFS.Open("index.html")
		if err == nil {
			defer file.Close()
			stat, _ := file.Stat()
			ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return ctx.SendStream(file, int(stat.Size()))
		}
	}

	// Serve from embedded files
	content, err := fs.ReadFile(spaFS, "index.html")
	if err != nil {
		log.WithError(err).Error("Failed to read SPA index.html")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return ctx.Send(content)
}

// mustSubFS creates a sub-filesystem or panics
func mustSubFS(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		log.WithError(err).WithField("dir", dir).Fatal("Failed to create sub-filesystem")
	}
	return sub
}
