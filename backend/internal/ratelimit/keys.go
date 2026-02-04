package ratelimit

import "github.com/gin-gonic/gin"

func ByIP(c *gin.Context) string {
	return Key("ip", c.ClientIP())
}

func ByUser(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		return Key("user", v.(string))
	}
	return ByIP(c)
}

func ByIPAndUsername(c *gin.Context) string {
	username := c.PostForm("username")
	return Key("login", c.ClientIP(), username)
}
