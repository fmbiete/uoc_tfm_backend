package api

import (
	"net/http"
	"strconv"
	"tfm_backend/models"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func (s *Server) OrderCount(c echo.Context) error {
	fromDate, err := time.Parse("2006-01-02", c.QueryParam("from"))
	if err != nil {
		log.Error().Err(err).Str("from", c.QueryParam("from")).Msg("Failed to convert to date")
		return c.NoContent(http.StatusBadRequest)
	}
	toDate, err := time.Parse("2006-01-02", c.QueryParam("to"))
	if err != nil {
		log.Error().Err(err).Str("to", c.QueryParam("to")).Msg("Failed to convert to date")
		return c.NoContent(http.StatusBadRequest)
	}

	counts, err := s.db.OrderCount(fromDate, toDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count Order")
		return err
	}

	return c.JSON(http.StatusOK, counts)
}

func (s *Server) OrderDelete(c echo.Context) error {
	orderId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		log.Error().Err(err).Str("id", c.Param("id")).Msg(msgErrorIdToInt)
		return c.NoContent(http.StatusBadRequest)
	}

	var userId = authenticatedUserId(c)

	// Only the owner of the Order can delete it
	err = s.db.OrderDelete(userId, orderId)
	if err != nil {
		log.Error().Err(err).Uint64("id", orderId).Msg("Failed to delete order")
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) OrderCreate(c echo.Context) error {
	var order models.Order
	err := c.Bind(&order)
	if err != nil {
		log.Error().Err(err).Msg("Failed to bind order")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	order.UserID = authenticatedUserId(c)

	order, err = s.db.OrderCreate(order)
	if err != nil {
		log.Error().Err(err).Interface("order", order).Msg("Failed to create order")
		return err
	}

	return c.JSON(http.StatusCreated, order)
}

func (s *Server) OrderDetails(c echo.Context) error {
	orderId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		log.Error().Err(err).Str("id", c.Param("id")).Msg(msgErrorIdToInt)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var userId int64 = int64(authenticatedUserId(c))
	if authenticatedIsAdministrator(c) {
		userId = -1
	}

	// Only the owner of the Order or an Administrator can see details

	order, err := s.db.OrderDetails(userId, orderId)
	if err != nil {
		log.Error().Err(err).Uint64("id", orderId).Msg("Failed to read order")
		return err
	}

	return c.JSON(http.StatusOK, order)
}

func (s *Server) OrderList(c echo.Context) error {
	var userId int64 = int64(authenticatedUserId(c))
	if authenticatedIsAdministrator(c) {
		userId = -1
	}

	var dayFilter string = c.QueryParam("day")

	limit, page, offset := parsePagination(c)

	orders, err := s.db.OrderList(userId, dayFilter, limit, offset)
	if err != nil {
		log.Error().Err(err).Int64("userId", userId).Str("day", dayFilter).Msg("Failed to list orders")
		return err
	}

	return c.JSON(http.StatusOK, models.PaginationOrders{Limit: limit, Page: page, Orders: orders})
}

func (s *Server) OrderLineCreate(c echo.Context) error {
	orderId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		log.Error().Err(err).Str("id", c.Param("id")).Msg(msgErrorIdToInt)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var line models.OrderLine
	err = c.Bind(&line)
	if err != nil {
		log.Error().Err(err).Msg("Failed to bind order line")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	line.OrderID = orderId

	// Only the owner of the Order can add lines
	var userId = authenticatedUserId(c)

	order, err := s.db.OrderLineCreate(userId, orderId, line)
	if err != nil {
		log.Error().Err(err).Interface("order", order).Msg("Failed to create order line")
		return err
	}

	return c.JSON(http.StatusOK, order)
}

func (s *Server) OrderLineDelete(c echo.Context) error {
	orderId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		log.Error().Err(err).Str("id", c.Param("id")).Msg(msgErrorIdToInt)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	lineId, err := strconv.ParseUint(c.Param("lineid"), 10, 64)
	if err != nil {
		log.Error().Err(err).Str("lineid", c.Param("lineid")).Msg(msgErrorIdToInt)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var userId = authenticatedUserId(c)

	// Only the owner of the Order can delete lines
	order, err := s.db.OrderLineDelete(userId, orderId, lineId)
	if err != nil {
		log.Error().Err(err).Interface("order", order).Msg("Failed to delete order line")
		return err
	}

	return c.JSON(http.StatusOK, order)
}

func (s *Server) OrderLineModify(c echo.Context) error {
	orderId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		log.Error().Err(err).Str("id", c.Param("id")).Msg(msgErrorIdToInt)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	lineId, err := strconv.ParseUint(c.Param("lineid"), 10, 64)
	if err != nil {
		log.Error().Err(err).Str("lineid", c.Param("lineid")).Msg(msgErrorIdToInt)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var line models.OrderLine
	err = c.Bind(&line)
	if err != nil {
		log.Error().Err(err).Msg("Failed to bind order line")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	line.OrderID = orderId
	line.ID = lineId

	var userId = authenticatedUserId(c)

	// Only the owner of the Order can modify lines
	order, err := s.db.OrderLineModify(userId, line)
	if err != nil {
		log.Error().Err(err).Interface("order", order).Msg("Failed to modify order line")
		return err
	}

	return c.JSON(http.StatusOK, order)
}

func (s *Server) OrderModifiable(c echo.Context) error {
	var userId = authenticatedUserId(c)
	orderId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		log.Error().Err(err).Str("id", c.Param("id")).Msg(msgErrorIdToInt)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	modifiable, err := s.db.OrderModifiable(orderId, userId)
	if err != nil {
		log.Error().Err(err).Uint64("orderId", orderId).Uint64("userId", userId).Msg("Failed to check if order is modifiable")
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"modifiable": modifiable})
}

func (s *Server) OrderSubvention(c echo.Context) error {
	var userId = authenticatedUserId(c)

	subvention, err := s.db.OrderSubvention(userId)
	if err != nil {
		log.Error().Err(err).Uint64("userId", userId).Msg("Failed to get today's subvention")
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"subvention": subvention})
}
