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
