// SPDX-License-Identifier: Unlicense OR MIT

package localization

func init() {
	russian["connection.title"] = "Аккаунт не подключён"
	russian["connection.failed"] = "Не удалось подключить аккаунт к Telegram, и новые сообщения не придут. Сохранённое можно читать и без подключения."
	russian["connection.in_use"] = "Этот аккаунт уже подключён: в другом окне или в другой копии приложения. Закройте её и повторите; сохранённое можно читать и без подключения."
	russian["connection.retry"] = "Повторить"
	russian["fatal.start"] = "Приложение не запустилось:"
	russian["fatal.crashed"] = "В прошлый раз приложение завершилось из-за ошибки. Отчёт сохранён:"

	english["connection.title"] = "Account not connected"
	english["connection.failed"] = "Could not connect the account to Telegram, so new messages will not arrive. What is saved can be read without a connection."
	english["connection.in_use"] = "This account is already connected, in another window or another copy of the application. Close it and try again; what is saved can be read without a connection."
	english["connection.retry"] = "Try again"
	english["fatal.start"] = "The application did not start:"
	english["fatal.crashed"] = "The application ended with an error last time. The report is saved:"
}
