[//]: # "Title: echoi18n"
[//]: # "Author: itpey"
[//]: # "Attendees: itpey"
[//]: # "Tags: #itpey #go #echo #i18n #go-lang #http #api #https #echo-i18n #echoi18n #middleware"

<h1 align="center">
Echo i18n Middleware
</h1>

<p align="center">
Echo i18n Middleware is a middleware package for the <a href="https://github.com/labstack/echo">Echo</a> web framework that provides internationalization (i18n) support.
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/itpey/echoi18n">
    <img src="https://pkg.go.dev/badge/github.com/itpey/echoi18n.svg" alt="itpey i18n echo middleware Go Reference">
  </a>
  <a href="https://github.com/itpey/echoi18n/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/itpey/echoi18n" alt="itpey i18n echo middleware license">
  </a>
</p>

# Features

- Seamless integration with Echo web framework.
- Support for loading message bundles in various formats such as YAML.
- Flexible language negotiation using HTTP Accept-Language header or query parameter.
- Panic-free message localization with error handling.

# Installation

```bash
go get github.com/itpey/echoi18n
```

# Usage

```go
package main

import (
	"net/http"

	"github.com/itpey/echoi18n"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func main() {
	e := echo.New()
	e.Use(
		echoi18n.NewMiddleware(&echoi18n.Config{
			RootPath:        "./localize",
			AcceptLanguages: []language.Tag{language.Chinese, language.English},
			DefaultLanguage: language.Chinese,
		}),
	)
	e.GET("/", func(c echo.Context) error {
		localize, err := echoi18n.Localize(c, "welcome")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, localize)
	})
	e.GET("/:name", func(c echo.Context) error {
		return c.String(http.StatusOK, echoi18n.MustLocalize(c, &i18n.LocalizeConfig{
			MessageID: "welcomeWithName",
			TemplateData: map[string]string{
				"name": c.QueryParam("name"),
			},
		}))
	})
	e.Logger.Fatal(e.Start(":1323"))
}
```

# Feedback and Contributions

If you encounter any issues or have suggestions for improvement, please [open an issue](https://github.com/itpey/echoi18n/issues) on GitHub.

We welcome contributions! Fork the repository, make your changes, and submit a pull request.

# License

echoi18n is open-source software released under the MIT License. You can find a copy of the license in the [LICENSE](https://github.com/itpey/echoi18n/blob/main/LICENSE) file.

# Author

echoi18n was created by [itpey](https://github.com/itpey)
