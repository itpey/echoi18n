package echoi18n

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

// newServer creates and configures a new Echo server with i18n middleware.
func newServer() *echo.Echo {
	e := echo.New()
	e.Use(NewMiddleware())

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

// i18nApp is an instance of the Echo server configured with the i18n middleware.
var i18nApp = newServer()

// makeRequest performs an HTTP request to the Echo server for testing purposes.
// lang specifies the desired language for the request.
// name is the URL path parameter.
// app is the Echo server instance.
// Returns the HTTP response and any error encountered.
func makeRequest(lang language.Tag, name string, app *echo.Echo) (*http.Response, error) {
	path := "/" + name
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	if lang != language.Und {
		req.Header.Add("Accept-Language", lang.String())
	}

	app.ServeHTTP(rec, req)
	return rec.Result(), nil
}

// TestI18nEN tests the localization middleware with English language.
func TestI18nEN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		lang language.Tag
		url  string
		want string
	}{
		{"hello world", language.English, "", "hello"},
		{"hello alex", language.English, "alex", "hello alex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeRequest(tt.lang, tt.url, i18nApp)
			assert.NoError(t, err)
			body, err := io.ReadAll(got.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(body))
		})
	}
}

// TestI18nZH tests the localization middleware with Chinese language.
func TestI18nZH(t *testing.T) {
	tests := []struct {
		name string
		lang language.Tag
		url  string
		want string
	}{
		{"hello world", language.Chinese, "", "你好"},
		{"hello alex", language.Chinese, "alex", "你好 alex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeRequest(tt.lang, tt.url, i18nApp)
			assert.NoError(t, err)
			body, err := io.ReadAll(got.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(body))
		})
	}
}

// TestParallelI18n tests the localization middleware concurrently with multiple languages.
func TestParallelI18n(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		lang language.Tag
		url  string
		want string
	}{
		{"hello world", language.Chinese, "", "你好"},
		{"hello alex", language.Chinese, "alex", "你好 alex"},
		{"hello peter", language.English, "peter", "hello peter"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := makeRequest(tt.lang, tt.url, i18nApp)
			assert.NoError(t, err)
			body, err := io.ReadAll(got.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(body))
		})
	}
}

// TestLocalize tests the Localize function directly.
func TestLocalize(t *testing.T) {
	t.Parallel()
	app := echo.New()
	app.Use(NewMiddleware())
	app.GET("/", func(ctx echo.Context) error {
		localize, err := Localize(ctx, "welcome?")
		if err != nil {
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		return ctx.String(http.StatusOK, localize)
	})

	app.GET("/:name", func(ctx echo.Context) error {
		name := ctx.Param("name")
		localize, err := Localize(ctx, &i18n.LocalizeConfig{
			MessageID: "welcomeWithName",
			TemplateData: map[string]string{
				"name": name,
			},
		})
		if err != nil {
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		return ctx.String(http.StatusOK, localize)
	})

	t.Run("test localize", func(t *testing.T) {
		got, err := makeRequest(language.Chinese, "", app)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, got.StatusCode)
		body, _ := io.ReadAll(got.Body)
		assert.Equal(t, `i18n.Localize error: message "welcome?" not found in language "zh"`, string(body))

		got, err = makeRequest(language.English, "name", app)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, got.StatusCode)
		body, _ = io.ReadAll(got.Body)
		assert.Equal(t, "hello name", string(body))
	})
}

// Test_defaultLangHandler tests the default language handler.
func Test_defaultLangHandler(t *testing.T) {
	e := echo.New()
	e.Use(NewMiddleware())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, defaultLangHandler(nil, language.English.String()))
	})
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, defaultLangHandler(c, language.English.String()))
	})

	t.Parallel()
	t.Run("test nil ctx", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func() {
				defer wg.Done()
				got, err := makeRequest(language.English, "", e)
				assert.NoError(t, err)
				body, _ := io.ReadAll(got.Body)
				assert.Equal(t, "en", string(body))
			}()
		}
		wg.Wait()
	})

	t.Run("test query and header", func(t *testing.T) {
		got, err := makeRequest(language.Chinese, "test?lang=en", e)
		assert.NoError(t, err)
		body, _ := io.ReadAll(got.Body)
		assert.Equal(t, "en", string(body))

		got, err = makeRequest(language.Chinese, "test", e)
		assert.NoError(t, err)
		body, _ = io.ReadAll(got.Body)
		assert.Equal(t, "zh", string(body))
	})
}
