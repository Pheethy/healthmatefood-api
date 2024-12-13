package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/joncalhoun/qson"
)

func (m *GoMiddleware) InputForm() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// สร้าง function สำหรับ validate form
		if err := Form(ctx); err != nil {
			var code int
			var message interface{}

			// ตรวจสอบว่าเป็น fiber.Error หรือไม่
			if ferr, ok := err.(*fiber.Error); ok {
				code = ferr.Code
				message = ferr.Message
			}

			return fiber.NewError(code, fmt.Sprintf("%v", message))
		}
		return ctx.Next()
	}
}

func Form(c *fiber.Ctx) error {
	data := map[string]interface{}{}
	reqMethod := c.Method()

	if reqMethod == http.MethodPost || reqMethod == http.MethodPut || reqMethod == http.MethodDelete {
		contentType := c.Get("Content-Type")

		if strings.Contains(contentType, "multipart/form-data") {
			/* Handler MULTIPART FORM */
			form, err := c.MultipartForm()
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, http.ErrMissingBoundary.Error()+" or has not any parameter")
			}

			bu, _ := qson.ToJSON(url.Values(form.Value).Encode())
			json.Unmarshal(bu, &data)

			data, err = parseOnKeyData(data)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}

			/* Handle file uploads */
			setFileLocals(c, form, "files")
			setFileLocals(c, form, "cover")
			setFileLocals(c, form, "images")
			setFileLocals(c, form, "other_imgs")

		} else if strings.Contains(strings.ToLower(contentType), "application/json") {
			/* Handler JSON */
			if err := c.BodyParser(&data); err != nil && err != io.EOF {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}

			var err error
			data, err = parseOnKeyData(data)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}

		} else if strings.Contains(strings.ToLower(contentType), "application/x-www-form-urlencoded") {
			/* Handler FORM Data*/
			if reqMethod == http.MethodDelete {
				body := c.Body()
				if len(body) > 0 {
					postForm, _ := url.ParseQuery(string(body))
					if len(postForm) > 0 {
						bu, _ := qson.ToJSON(postForm.Encode())
						json.Unmarshal(bu, &data)
					}
				}
			} else {
				postForm := c.Request().PostArgs()
				if postForm.Len() > 0 {
					formMap := make(url.Values)
					postForm.VisitAll(func(key, value []byte) {
						formMap.Add(string(key), string(value))
					})
					bu, _ := qson.ToJSON(formMap.Encode())
					json.Unmarshal(bu, &data)
				}
			}

			var err error
			data, err = parseOnKeyData(data)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
		}
	}

	if len(data) > 0 {
		c.Locals("params", data)
	}
	return nil
}

func setFileLocals(c *fiber.Ctx, form *multipart.Form, key string) {
	if files := form.File[key]; files != nil {
		c.Locals(key, files)
	} else {
		c.Locals(key, make([]*multipart.FileHeader, 0))
	}
}

// parseOnKeyData function remains the same
func parseOnKeyData(data map[string]interface{}) (map[string]interface{}, error) {
	if data != nil && len(data) == 1 {
		if v, ok := data["data"]; ok {
			valueType := reflect.ValueOf(v).Kind()
			if valueType == reflect.Map {
				data = v.(map[string]interface{})
			} else if valueType == reflect.String {
				data = map[string]interface{}{}
				if err := json.Unmarshal([]byte(v.(string)), &data); err != nil {
					return data, err
				}
			}
		}
	}
	return data, nil
}
