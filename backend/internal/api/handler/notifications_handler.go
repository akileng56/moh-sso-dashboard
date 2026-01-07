package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
)

type NotificationsHandler struct {
	NotificationsSvc service.NotificationsService
}

func NewNotificationsHandler(svc service.NotificationsService) *NotificationsHandler {
	return &NotificationsHandler{
		NotificationsSvc: svc,
	}
}

func (h *NotificationsHandler) Notify(c *gin.Context) {
	var input model.Notification
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	n, err := h.NotificationsSvc.Notify(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create notification"})
		return
	}

	c.JSON(http.StatusCreated, n)
}

func (h *NotificationsHandler) ListNotifications(c *gin.Context) {
	role := getTargetRole(c)

	// unread (optional)
	var unread *bool
	if v := c.Query("unread"); v != "" {
		b := v == "true"
		unread = &b
	}

	// limit
	limit := int32(20)
	if v := c.Query("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	// offset
	offset := int32(0)
	if v := c.Query("offset"); v != "" {
		if o, err := strconv.Atoi(v); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	notifications, err := h.NotificationsSvc.ListNotifications(
		c.Request.Context(),
		role,
		unread,
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list notifications",
		})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *NotificationsHandler) GetNotificationByID(c *gin.Context) {
	id := c.Param("id")

	n, err := h.NotificationsSvc.GetNotificationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, n)
}

func (h *NotificationsHandler) MarkNotificationAsRead(c *gin.Context) {
	id := c.Param("id")
	if err := h.NotificationsSvc.MarkNotificationAsRead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as read"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *NotificationsHandler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")
	if err := h.NotificationsSvc.DeleteNotification(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notification"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *NotificationsHandler) DeleteOldNotifications(c *gin.Context) {
	if err := h.NotificationsSvc.DeleteOldNotifications(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cleanup notifications"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *NotificationsHandler) CountNotifications(c *gin.Context) {
	role := getTargetRole(c)

	count, err := h.NotificationsSvc.CountNotifications(c.Request.Context(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *NotificationsHandler) CountUnreadNotificationsCount(c *gin.Context) {
	role := getTargetRole(c)

	count, err := h.NotificationsSvc.CountUnreadNotificationsCount(c.Request.Context(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count unread notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func getTargetRole(c *gin.Context) string {
	if isAdmin, ok := c.Get("is_admin"); ok && isAdmin == true {
		return "admin"
	}

	if roles, ok := c.Get("client_roles"); ok {
		if cr, ok := roles.(map[string][]string); ok {
			if len(cr) > 0 {
				return "user"
			}
		}
	}

	return "user"
}
