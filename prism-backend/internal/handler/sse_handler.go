package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ridofiqri79/prism-backend/internal/sse"
)

func SSEHandler(broker *sse.Broker) echo.HandlerFunc {
	return func(c echo.Context) error {
		clientID := c.QueryParam("user_id")
		if clientID == "" {
			clientID = c.RealIP()
		}

		eventCh, unsubscribe := broker.Subscribe(clientID)
		defer unsubscribe()

		res := c.Response()
		res.Header().Set(echo.HeaderContentType, "text/event-stream")
		res.Header().Set(echo.HeaderCacheControl, "no-cache")
		res.Header().Set(echo.HeaderConnection, "keep-alive")
		res.WriteHeader(http.StatusOK)
		res.Flush()

		for {
			select {
			case <-c.Request().Context().Done():
				return nil
			case payload, ok := <-eventCh:
				if !ok {
					return nil
				}

				if _, err := fmt.Fprintf(res, "data: %s\n\n", payload); err != nil {
					return err
				}
				res.Flush()
			}
		}
	}
}
