package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func Setup() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v
	}
}
