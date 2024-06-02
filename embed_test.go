package echoi18n

import (
	"embed"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

//go:embed example/localizeJSON/*
var fs embed.FS

// newEmbedServer creates and configures a new Echo server with i18n middleware.
func newEmbedServer() *echo.Echo {
	e := echo.New()
	e.Use(NewMiddleware(&Config{
		Loader:           &EmbedLoader{fs},
		UnmarshalFunc:    json.Unmarshal,
		RootPath:         "./example/localizeJSON/",
		FormatBundleFile: "json",
	}))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, MustLocalize(c, "welcome"))
	})
	e.GET("/:name", func(c echo.Context) error {
		return c.String(http.StatusOK, MustLocalize(c, &i18n.LocalizeConfig{
			MessageID: "welcomeWithName",
			TemplateData: map[string]string{
				"name": c.Param("name"),
			},
		}))
	})
	return e
}

// embedApp is an instance of the Echo server configured with the embedded filesystem.
var embedApp = newEmbedServer()

// TestEmbedLoader_LoadMessage tests the EmbedLoader's LoadMessage method with various scenarios.
func TestEmbedLoader_LoadMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		lang language.Tag
		url  string
		want string
	}{
		{"hello world", language.English, "", "hello"},
		{"hello alex", language.Chinese, "", "你好"},
		{"hello alex", language.English, "alex", "hello alex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeRequest(tt.lang, tt.url, embedApp)
			assert.NoError(t, err)
			body, err := io.ReadAll(got.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(body))
		})
	}
}
