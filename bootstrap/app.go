package bootstrap

import (
	"fmt"

	"Contruction-Project/lib/logger"
	"Contruction-Project/lib/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	idTranslations "github.com/go-playground/validator/v10/translations/id"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

// App instance of the skeleton
type App struct {
	Debug      bool
	R          *chi.Mux
	DB         *pgxpool.Pool
	Config     utils.Config
	Validator  *Validator
	Log        logger.Contract
	Redis      *redis.Client
	RedisCache *redis.Client
}

// Validator set validator instance
type Validator struct {
	Driver     *validator.Validate
	Uni        *ut.UniversalTranslator
	Translator ut.Translator
}

// SetupValidator create new instance of validator driver
func SetupValidator(config utils.Config) *Validator {
	en := en.New()
	id := id.New()
	uni := ut.New(en, id)

	transEN, _ := uni.GetTranslator("en")
	transID, _ := uni.GetTranslator("id")

	validatorDriver := validator.New()

	_ = enTranslations.RegisterDefaultTranslations(validatorDriver, transEN)
	_ = idTranslations.RegisterDefaultTranslations(validatorDriver, transID)

	var trans ut.Translator
	switch config.GetString("app.locale") {
	case "en":
		trans = transEN
	case "id":
		trans = transID
	}

	return &Validator{Driver: validatorDriver, Uni: uni, Translator: trans}
}

// SetupLogger create new instance of logger pacakge
func SetupLogger(config utils.Config) logger.Contract {
	def := config.GetString("log.default")
	source := fmt.Sprintf("log.%s.source", def)
	return logger.New(
		def, config.GetString(source),
	)
}

// SetupRedis ...
func SetupRedis(addr string, pass string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})

	return rdb, nil
}
