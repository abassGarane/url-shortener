package api

import (
	"fmt"
	"log"
	"net/http"

	json "github.com/abassGarane/url_shortener/serializers/json"
	ms "github.com/abassGarane/url_shortener/serializers/msgpack"
	"github.com/abassGarane/url_shortener/shortener"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type RedirectHandler interface {
	Get(*fiber.Ctx) error
	Post(*fiber.Ctx) error
}

type handler struct {
	redirectService shortener.RedirectService
}

func NewHandler(redirectService shortener.RedirectService) RedirectHandler {
	return &handler{
		redirectService: redirectService,
	}
}

func setupResponse(c *fiber.Ctx, contentType string, statusCode int, body []byte) {
	c.Set("Content-Type", contentType)
	c.Status(statusCode)
	_, err := c.Write(body)
	if err != nil {
		log.Println(err)
	}
}

func (h *handler) serializer(contentType string) shortener.RedirectSerializer {
	if contentType == "application/x-msgpack" {
		return &ms.Redirect{}
	}
	return &json.Redirect{}
}
func (h *handler) Get(c *fiber.Ctx) error {
	code := c.Params("code")
	redirect, err := h.redirectService.Find(code)
	if err != nil {
		if errors.Cause(err) == shortener.ErrorRedirectNotFound {
			c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": fmt.Sprintf("Redirect of code %s not found", code),
			})
			return
		}
		c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
		return
	}
	c.Redirect(redirect.URL, http.StatusMovedPermanently)
}
func (h *handler) Post(c *fiber.Ctx) {
	contentType := c.GetReqHeaders()["Content-Type"]
	body := c.Body()
	if len(body) == 0 {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": fmt.Sprintf("Error %s", fiber.ErrInternalServerError.Error()),
		})
		return
	}
	redirect, err := h.serializer(contentType).Decode(body)
	if err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Internal server error",
		})
		return
	}
	if err = h.redirectService.Store(redirect); err != nil {
		if errors.Cause(err) == shortener.ErrorRedirectInvalid {
			c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"message": fmt.Sprintf("Bad request %s", err.Error()),
			})
			return
		}
	}
	responseBody, err := h.serializer(contentType).Encode(redirect)
	if err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Internal server error",
		})
		return
	}
	setupResponse(c, contentType, fiber.StatusCreated, responseBody)
}
