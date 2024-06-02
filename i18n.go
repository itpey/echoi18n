package echoi18n

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// localsKey is the key used to store the i18n Config in the Echo Context.
const localsKey = "echoi18n"

// Config holds the configuration for the i18n middleware.
type Config struct {
	DefaultLanguage  language.Tag                      // Default language to use if no language is determined.
	AcceptLanguages  []language.Tag                    // Supported languages.
	FormatBundleFile string                            // File format for message bundles.
	Loader           Loader                            // Loader interface to load message files.
	RootPath         string                            // Root directory path for message files.
	LangHandler      func(echo.Context, string) string // Language handler function.
	bundle           *i18n.Bundle                      // i18n message bundle.
	localizerMap     *sync.Map                         // Map of localizers for each language.
	mu               sync.Mutex                        // Mutex for thread safety.
	UnmarshalFunc    i18n.UnmarshalFunc                // Function to unmarshal message files.
}

// Loader is the interface for loading message files.
type Loader interface {
	LoadMessage(path string) ([]byte, error)
}

// LoaderFunc is a function type that implements the Loader interface.
type LoaderFunc func(path string) ([]byte, error)

// LoadMessage loads a message file using the LoaderFunc.
func (f LoaderFunc) LoadMessage(path string) ([]byte, error) {
	return f(path)
}

// loadMessage loads a single message file for a given language.
func (c *Config) loadMessage(filepath string) {
	buf, err := c.Loader.LoadMessage(filepath)
	if err != nil {
		panic(err)
	}
	if _, err := c.bundle.ParseMessageFileBytes(buf, filepath); err != nil {
		panic(err)
	}
}

// loadMessages loads all message files for the supported languages.
func (c *Config) loadMessages() {
	for _, lang := range c.AcceptLanguages {
		bundleFilePath := fmt.Sprintf("%s.%s", lang.String(), c.FormatBundleFile)
		filepath := path.Join(c.RootPath, bundleFilePath)
		c.loadMessage(filepath)
	}
}

// initLocalizerMap initializes localizers for each supported language.
func (c *Config) initLocalizerMap() {
	localizerMap := &sync.Map{}

	for _, lang := range c.AcceptLanguages {
		s := lang.String()
		localizerMap.Store(s, i18n.NewLocalizer(c.bundle, s))
	}

	lang := c.DefaultLanguage.String()
	if _, ok := localizerMap.Load(lang); !ok {
		localizerMap.Store(lang, i18n.NewLocalizer(c.bundle, lang))
	}
	c.mu.Lock()
	c.localizerMap = localizerMap
	c.mu.Unlock()
}

// Localize localizes a message using the provided context and parameters.
func Localize(c echo.Context, params interface{}) (string, error) {
	local := c.Get(localsKey)
	if local == nil {
		return "", fmt.Errorf("i18n.Localize error: %v", "Config is nil")
	}

	appCfg, ok := local.(*Config)
	if !ok {
		return "", fmt.Errorf("i18n.Localize error: %v", "Config is not *Config type")
	}

	lang := appCfg.LangHandler(c, appCfg.DefaultLanguage.String())
	localizer, _ := appCfg.localizerMap.Load(lang)

	if localizer == nil {
		defaultLang := appCfg.DefaultLanguage.String()
		localizer, _ = appCfg.localizerMap.Load(defaultLang)
	}

	var localizeConfig *i18n.LocalizeConfig
	switch paramValue := params.(type) {
	case string:
		localizeConfig = &i18n.LocalizeConfig{MessageID: paramValue}
	case *i18n.LocalizeConfig:
		localizeConfig = paramValue
	default:
		return "", fmt.Errorf("i18n.Localize error: %v", "Invalid params type")
	}

	message, err := localizer.(*i18n.Localizer).Localize(localizeConfig)
	if err != nil {
		return "", fmt.Errorf("i18n.Localize error: %v", err)
	}
	return message, nil
}

// MustLocalize is a helper function to localize a message, panicking on error.
func MustLocalize(c echo.Context, params interface{}) string {
	message, err := Localize(c, params)
	if err != nil {
		panic(err)
	}
	return message
}

// NewMiddleware creates a new i18n middleware handler with the provided configuration.
func NewMiddleware(config ...*Config) echo.MiddlewareFunc {
	cfg := configDefault(config...)
	bundle := i18n.NewBundle(cfg.DefaultLanguage)
	bundle.RegisterUnmarshalFunc(cfg.FormatBundleFile, cfg.UnmarshalFunc)
	cfg.bundle = bundle

	cfg.loadMessages()
	cfg.initLocalizerMap()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(localsKey, cfg)
			return next(c)
		}
	}
}

var ConfigDefault = &Config{
	DefaultLanguage:  language.English,
	AcceptLanguages:  []language.Tag{language.Chinese, language.English},
	FormatBundleFile: "yaml",
	Loader:           LoaderFunc(os.ReadFile),
	RootPath:         "./example/localize",
	LangHandler:      defaultLangHandler,
	UnmarshalFunc:    yaml.Unmarshal,
}

// configDefault provides default values for the configuration
func configDefault(config ...*Config) *Config {

	if len(config) == 0 {
		return ConfigDefault
	}

	cfg := config[0]

	if cfg.DefaultLanguage == language.Und {
		cfg.DefaultLanguage = language.English
	}
	if cfg.AcceptLanguages == nil {
		cfg.AcceptLanguages = []language.Tag{language.Chinese, language.English}
	}
	if cfg.FormatBundleFile == "" {
		cfg.FormatBundleFile = "yaml"
	}
	if cfg.Loader == nil {
		cfg.Loader = LoaderFunc(os.ReadFile)
	}
	if cfg.RootPath == "" {
		cfg.RootPath = "./example/localize"
	}
	if cfg.LangHandler == nil {
		cfg.LangHandler = defaultLangHandler
	}

	if cfg.UnmarshalFunc == nil {
		cfg.UnmarshalFunc = yaml.Unmarshal
	}
	return cfg
}

// defaultLangHandler returns the default language based on the request context.
func defaultLangHandler(c echo.Context, defaultLang string) string {
	if c == nil || c.Request() == nil {
		return defaultLang
	}
	var lang string
	lang = c.QueryParam("lang")
	if lang != "" {
		return lang
	}
	lang = c.Request().Header.Get("Accept-Language")
	if lang != "" {
		return lang
	}

	return defaultLang
}
