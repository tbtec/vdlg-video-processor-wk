package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/tbtec/vdlg/internal/infra/httpserver"
)

func NewNotFound() func(ctx *fiber.Ctx) error {
	return func(fc *fiber.Ctx) error {
		return fc.Status(http.StatusNotFound).
			JSON(httpserver.NewErrorMessage("404", "URL path not found"))
	}
}
