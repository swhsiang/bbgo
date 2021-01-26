package bbgo

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/c9s/bbgo/pkg/service"
	"github.com/c9s/bbgo/pkg/types"
)

func RunServer(ctx context.Context, userConfig *Config, environ *Environment) error {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.GET("/api/trading-volume", func(c *gin.Context) {
		period := c.DefaultQuery("period", "day")
		startTimeString := c.DefaultQuery("start-time", time.Now().AddDate(0, 0, -7).Format(time.RFC3339))

		startTime, err := time.Parse(time.RFC3339, startTimeString)
		if err != nil {
			c.Status(http.StatusBadRequest)
			log.WithError(err).Error("start-time format incorrect")
			return
		}

		rows, err := environ.TradeService.QueryTradingVolume(startTime, service.TradingVolumeQueryOptions{
			GroupByExchange: false,
			GroupByPeriod:   period,
		})
		if err != nil {
			log.WithError(err).Error("trading volume query error")
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, gin.H{"tradingVolumes": rows})
		return
	})

	r.GET("/api/sessions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"sessions": userConfig.Sessions})
	})

	r.GET("/api/assets", func(c *gin.Context) {
		totalAssets := types.AssetMap{}

		for _, session := range environ.sessions {
			balances := session.Account.Balances()

			if err := session.UpdatePrices(ctx); err != nil {
				log.WithError(err).Error("price update failed")
				c.Status(http.StatusInternalServerError)
				return
			}

			assets := balances.Assets(session.lastPrices)

			for currency, asset := range assets {
				totalAssets[currency] = asset
			}
		}

		c.JSON(http.StatusOK, gin.H{"assets": totalAssets})
	})

	r.GET("/api/sessions/:session", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"session": session})
	})

	r.GET("/api/sessions/:session/trades", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"trades": session.Trades})
	})

	r.GET("/api/sessions/:session/open-orders", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		marketOrders := make(map[string][]types.Order)
		for symbol, orderStore := range session.orderStores {
			marketOrders[symbol] = orderStore.Orders()
		}

		c.JSON(http.StatusOK, gin.H{"orders": marketOrders})
	})

	r.GET("/api/sessions/:session/account", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"account": session.Account})
	})

	r.GET("/api/sessions/:session/account/balances", func(c *gin.Context) {
		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		if session.Account == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("the account of session %s is nil", sessionName)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"balances": session.Account.Balances()})
	})

	r.GET("/api/sessions/:session/symbols", func(c *gin.Context) {

		sessionName := c.Param("session")
		session, ok := environ.Session(sessionName)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("session %s not found", sessionName)})
			return
		}

		var symbols []string
		for s := range session.loadedSymbols {
			symbols = append(symbols, s)
		}

		c.JSON(http.StatusOK, gin.H{"symbols": symbols})
	})

	r.GET("/api/sessions/:session/pnl", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/api/sessions/:session/market/:symbol/open-orders", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/api/sessions/:session/market/:symbol/trades", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/api/sessions/:session/market/:symbol/pnl", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	return r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}