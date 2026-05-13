package v1

import (
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/rabbitmq/rmq_rpc/server"
	"github.com/go-playground/validator/v10"
)

// NewRoutes -.
func NewRoutes(
	routes map[string]server.CallHandler,
	t usecase.Translation,
	u usecase.User,
	tk usecase.Task,
	id usecase.ImportantDay,
	n usecase.Notification,
	d usecase.DeviceToken,
	j *jwt.Manager,
	l logger.Interface,
) {
	r := &V1{t: t, u: u, tk: tk, id: id, n: n, d: d, j: j, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	routes["v1.auth.register"] = r.register()
	routes["v1.auth.login"] = r.login()

	routes["v1.translation.getHistory"] = r.getHistory()
	routes["v1.translation.translate"] = r.translate()

	routes["v1.task.create"] = r.createTask()
	routes["v1.task.get"] = r.getTask()
	routes["v1.task.list"] = r.listTasks()
	routes["v1.task.update"] = r.updateTask()
	routes["v1.task.transition"] = r.transitionTask()
	routes["v1.task.delete"] = r.deleteTask()

	routes["v1.important_day.create"] = r.createImportantDay()
	routes["v1.important_day.get"] = r.getImportantDay()
	routes["v1.important_day.list"] = r.listImportantDays()
	routes["v1.important_day.upcoming"] = r.upcomingImportantDays()
	routes["v1.important_day.update"] = r.updateImportantDay()
	routes["v1.important_day.replaceReminders"] = r.replaceImportantDayReminders()
	routes["v1.important_day.delete"] = r.deleteImportantDay()

	routes["v1.notification.list"] = r.listNotifications()
	routes["v1.notification.markRead"] = r.markNotificationRead()
	routes["v1.notification.markAllRead"] = r.markAllNotificationsRead()

	routes["v1.device.register"] = r.registerDevice()
	routes["v1.device.delete"] = r.deleteDevice()
}
