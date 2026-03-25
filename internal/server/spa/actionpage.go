package spa

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// RenderActionResultPage renders a standalone HTML page for action results
// (email verification, token recreation, etc.) styled consistently with the SPA
func RenderActionResultPage(ctx *fiber.Ctx, status int, title, message string, isSuccess bool) error {
	icon := "check-circle"
	colorClass := "success"
	if !isSuccess {
		if status >= 500 {
			icon = "times-circle"
			colorClass = "danger"
		} else {
			icon = "exclamation-triangle"
			colorClass = "warning"
		}
	}

	html := fmt.Sprintf(actionPageTemplate, title, icon, colorClass, title, message)
	ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return ctx.Status(status).SendString(html)
}

const actionPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>%s - mytoken</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.4/css/all.min.css">
    <style>
        :root, [data-bs-theme="light"] {
            --mytoken-primary: #df691a;
            --mytoken-primary-hover: #c73500;
        }
        [data-bs-theme="dark"] {
            --mytoken-primary: #df691a;
            --mytoken-primary-hover: #e8823a;
        }
        .bg-primary { background-color: var(--mytoken-primary) !important; }
        .btn-primary {
            background-color: var(--mytoken-primary);
            border-color: var(--mytoken-primary);
        }
        .btn-primary:hover {
            background-color: var(--mytoken-primary-hover);
            border-color: var(--mytoken-primary-hover);
        }
        .navbar-brand img { filter: brightness(0) invert(1); }
        .app { min-height: 100vh; display: flex; flex-direction: column; }
        main { flex-grow: 1; }
        .footer a { color: rgba(255,255,255,0.85); text-decoration: none; }
        .footer a:hover { text-decoration: underline; }
    </style>
    <script>
        (function() {
            var stored = localStorage.getItem('mytoken-theme');
            var theme = 'light';
            if (stored === 'dark') theme = 'dark';
            else if (stored === 'auto' || !stored) {
                theme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
            }
            document.documentElement.setAttribute('data-bs-theme', theme);
        })();
    </script>
</head>
<body>
<div class="app">
    <nav class="navbar navbar-dark bg-primary">
        <div class="container-fluid">
            <a class="navbar-brand d-flex align-items-center" href="/">
                <img src="/static/img/mytoken.png" alt="mytoken" height="30" class="me-2">
                mytoken
            </a>
        </div>
    </nav>
    
    <main class="container py-5">
        <div class="row justify-content-center">
            <div class="col-md-6 col-lg-5">
                <div class="card text-center shadow-sm">
                    <div class="card-body py-5">
                        <i class="fas fa-%s fa-4x mb-4 text-%s"></i>
                        <h2 class="card-title mb-3">%s</h2>
                        <p class="card-text text-muted mb-4">%s</p>
                        <a href="/" class="btn btn-primary">
                            <i class="fas fa-home me-2"></i>Go to Home
                        </a>
                    </div>
                </div>
            </div>
        </div>
    </main>
    
    <footer class="footer mt-auto py-3 bg-dark text-light">
        <div class="container">
            <div class="d-flex flex-wrap justify-content-center align-items-center gap-3">
                <span class="text-muted">&copy; 2024 KIT</span>
                <a href="/privacy">Privacy</a>
                <a href="https://mytoken-docs.data.kit.edu" target="_blank">Documentation</a>
                <a href="https://github.com/oidc-mytoken/server" target="_blank">
                    <i class="fab fa-github me-1"></i>Source
                </a>
            </div>
        </div>
    </footer>
</div>
</body>
</html>`
