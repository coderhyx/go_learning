package main

import (
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var validate *validator.Validate
var trans ut.Translator

func InitValidate() {
	fmt.Println("init validate初始化了")
	validate = validator.New()
	uni := ut.New(zh.New())            // 创建中文翻译器
	trans, _ = uni.GetTranslator("zh") // 获取中文翻译器

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册中文翻译
		_ = zh_translations.RegisterDefaultTranslations(v, trans)

		// 自定义错误消息模板
		v.RegisterTranslation("required", trans, func(ut ut.Translator) error {
			return ut.Add("required", "{0} 是必填字段", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("required", fe.Field())
			return t
		})
		// 自定义邮箱错误模板
		v.RegisterTranslation("email", trans, func(ut ut.Translator) error {
			return ut.Add("email", "{0} 不是有效的邮箱地址", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("email", fe.Field())
			return t
		})
		// 自定义最小和最大长度错误模板
		v.RegisterTranslation("min", trans, func(ut ut.Translator) error {
			return ut.Add("min", "{0} 的长度不能小于 {1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("min", fe.Field(), fe.Param())
			return t
		})

		v.RegisterTranslation("max", trans, func(ut ut.Translator) error {
			return ut.Add("max", "{0} 的长度不能大于 {1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("max", fe.Field(), fe.Param())
			return t
		})
		v.RegisterTranslation("max", trans, func(ut ut.Translator) error {
			return ut.Add("alphanum", "{0} 只能包含字母和数字", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("max", fe.Field())
			return t
		})
	}
}

func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return err
	}
	return nil
}
func TranslateValidationError(err error) string {
	errs := err.(validator.ValidationErrors)
	var errMsg string
	for _, e := range errs {
		errMsg += e.Translate(trans) + ","
	}
	return errMsg
}
