package api

import (
	"errors"
	"fmt"
	"net/http"
	"tfm_backend/config"
	"tfm_backend/orm"

	"github.com/google/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v3"
)

type Server struct {
	e             *echo.Echo
	db            *orm.Database
	cfg           *config.ConfigServer
	requiresLogin echo.MiddlewareFunc
	optionalLogin echo.MiddlewareFunc
}

const msgErrorIdToInt = "Failed to convert ID to int64"

func NewServer(cfg config.ConfigServer, db *orm.Database) *Server {
	s := Server{e: echo.New(), cfg: &cfg, db: db}

	s.requiresLogin = echojwt.WithConfig(echojwt.Config{SigningKey: []byte(s.cfg.JWTSecret)})
	s.optionalLogin = echojwt.WithConfig(
		echojwt.Config{
			SigningKey:             []byte(s.cfg.JWTSecret),
			ContinueOnIgnoredError: true,
			ErrorHandler: func(c echo.Context, err error) error {
				fmt.Println(err)
				if errors.Is(err, echojwt.ErrJWTMissing) {
					return nil
				}
				return err
			},
		},
	)

	s.e.HideBanner = true
	s.e.Logger = lecho.From(log.Logger)

	// s.e.Use(middleware.Logger())
	s.e.Use(middleware.Recover())

	s.e.HTTPErrorHandler = customHTTPErrorHandler

	return &s
}

func customHTTPErrorHandler(err error, c echo.Context) {
	uuid := uuid.NewString()
	log.Error().Err(err).Str("uuid", uuid).Msg("Reflection")
	if he, ok := err.(*echo.HTTPError); ok {
		c.JSON(he.Code, map[string]interface{}{"message": he.Message, "reflection": uuid})
	} else {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": err.Error(), "reflection": uuid})
	}
}

func (s *Server) Listen() error {
	s.e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "TFM Backend API")
	})

	// Configuration API
	gConfiguration := s.e.Group("/configuration")
	gConfiguration.GET("/", s.ConfigurationDetails, s.requiresLogin, requiresRestaurador)
	gConfiguration.PATCH("/", s.ConfigurationModify, s.requiresLogin, requiresRestaurador)

	// User API
	gUser := s.e.Group("/user")
	gUser.POST("/login", s.Login)
	gUser.POST("/", s.UserCreate)
	gUser.GET("/:id", s.UserDetails)
	gUser.PATCH("/:id", s.UserModify, s.requiresLogin)
	gUser.DELETE("/:id", s.UserDelete, s.requiresLogin)

	// Dishes API
	gDishes := s.e.Group("/dish")
	gDishes.GET("/:id", s.DishDetails)
	gDishes.POST("/", s.DishCreate, s.requiresLogin, requiresRestaurador)
	gDishes.PATCH("/:id", s.DishModify, s.requiresLogin, requiresRestaurador)
	gDishes.DELETE("/:id", s.DishDelete, s.requiresLogin, requiresRestaurador)
	// /dishes/ is authenticated (show list of favourite dishes for user) and unauthenticated (show list of favourite dishes for everybody)
	s.e.GET("/dishes/", s.DishList, s.optionalLogin)

	// Promotions API
	gPromotions := s.e.Group("/promotion")
	gPromotions.GET("/:id", s.PromotionDetails)
	gPromotions.POST("/", s.PromotionCreate, s.requiresLogin, requiresRestaurador)
	gPromotions.PATCH("/:id", s.PromotionModify, s.requiresLogin, requiresRestaurador)
	gPromotions.DELETE("/:id", s.PromotionDelete, s.requiresLogin, requiresRestaurador)
	s.e.GET("/promotions/", s.PromotionList)

	// Cart API
	gCarts := s.e.Group("/cart")
	gCarts.GET("/", s.CartDetails, s.requiresLogin)
	gCarts.POST("/", s.CartSave, s.requiresLogin)
	gCarts.DELETE("/", s.CartDelete, s.requiresLogin)

	// Orders API
	gOrders := s.e.Group("/order")
	gOrders.POST("/", s.OrderCreateFromCart, s.requiresLogin)
	gOrders.GET("/:id", s.OrderDetails, s.requiresLogin)
	gOrders.DELETE("/:id", s.OrderCancel, s.requiresLogin)
	gOrders.POST("/:id/line/", s.OrderLineCreate, s.requiresLogin)
	gOrders.PATCH("/:id/line/:lineid", s.OrderLineModify, s.requiresLogin)
	gOrders.DELETE("/:id/line/:lineid", s.OrderLineDelete, s.requiresLogin)
	s.e.GET("/orders/", s.OrderList, s.requiresLogin)

	return s.e.Start(fmt.Sprintf(`:%d`, s.cfg.Port))
}
