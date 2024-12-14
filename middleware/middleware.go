package middleware

import (
	"context"
	"fmt"
	"healthmatefood-api/config"
	"healthmatefood-api/service/auth"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sirupsen/logrus"
)

type GoMiddlewareInf interface {
	SetTracer() fiber.Handler
	Cors() fiber.Handler
	Logger() fiber.Handler
	InputForm() fiber.Handler
}

type GoMiddleware struct {
	ctx      context.Context
	cfg      config.Iconfig
	authRepo auth.IAuthRepository
}

// InitMiddleware intialize the middleware
func InitMiddleware(cfg config.Iconfig, authRepo auth.IAuthRepository) GoMiddlewareInf {
	return &GoMiddleware{
		ctx:      context.TODO(),
		cfg:      cfg,
		authRepo: authRepo,
	}
}

func (m GoMiddleware) SetTracer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var span opentracing.Span
		ctx := c.UserContext()
		spanName := fmt.Sprintf("%s %s %s", string(c.Context().Request.URI().Scheme()), c.Method(), c.Path())
		spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.GetReqHeaders()))
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		switch err {
		case nil:
			span = opentracing.StartSpan(spanName, ext.RPCServerOption(spanCtx))
		case opentracing.ErrSpanContextNotFound:
			span, ctx = opentracing.StartSpanFromContext(ctx, spanName)
		default:
			logrus.Println("error default")
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		defer span.Finish()

		c.SetUserContext(ctx)
		// Proceed to the next handler
		err = c.Next()

		m.setTagByFiber(span, c)
		m.setLogByFiber(span, c)

		if err != nil {
			m.setError(span, c, err)
		} else {
			span.SetTag("error", false)
			span.SetTag("http.status_code", c.Response().StatusCode())
		}

		return nil
	}
}

func (m GoMiddleware) Cors() fiber.Handler {
	return cors.New(cors.Config{
		Next:             cors.ConfigDefault.Next,
		AllowOrigins:     "*",
		AllowMethods:     "GET, POST, PUT, PATCH, HEAD, DELETE",
		AllowHeaders:     "",
		AllowCredentials: false,
		ExposeHeaders:    "",
		MaxAge:           0,
	})
}

func (m GoMiddleware) JwtAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		mapClaims, err := m.authRepo.ParseToken(token)
		if err != nil {
			return fiber.NewError(http.StatusUnauthorized, err.Error())
		}
		if !m.authRepo.FindAccessToken(ctx, mapClaims.Payload.Id, token) {
			return fiber.NewError(http.StatusUnauthorized, "no permission to access")
		}
		c.Locals("user_id", mapClaims.Payload.Id)
		c.Locals("role_id", mapClaims.Payload.RoleId)
		return c.Next()
	}
}

func (m GoMiddleware) Logger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "👽 ${time} [${ip}] ${status} - ${method} ${path}\n",
		TimeFormat: "2006-01-02",
		TimeZone:   "Bangkok/Asia",
	})
}

func (m GoMiddleware) setTagByFiber(span opentracing.Span, c *fiber.Ctx) {
	span.SetTag("host", c.Hostname())
	span.SetTag("User-Agent", c.Get("User-Agent"))
	span.SetTag("http.method", c.Method())
	span.SetTag("http.url", c.OriginalURL())
}

func (m GoMiddleware) setLogByFiber(span opentracing.Span, c *fiber.Ctx) {
	span.LogFields(
		log.String("querystring", c.Context().QueryArgs().String()),
	)
}

func (m GoMiddleware) setError(span opentracing.Span, c *fiber.Ctx, err error) {
	isError := err != nil && c.Response().StatusCode() >= http.StatusBadRequest
	span.SetTag("error", isError)
	if isError {
		span.SetTag("http.status_code", c.Response().StatusCode())
		span.LogFields(log.Message(err.Error()))
	}
}
