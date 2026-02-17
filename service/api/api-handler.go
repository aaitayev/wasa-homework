package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	rt.router.POST("/session", rt.wrap(rt.doLogin))
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))

	return rt.router
}
