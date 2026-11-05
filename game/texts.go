package game

// All player-facing strings live here. The premise is a placeholder and is
// deliberately cheap to replace: feature code references these constants only.

const (
	// Title is shown in the window title bar.
	Title = "Остання зміна"

	// HUD labels.
	HUDStamps   = "Марки"
	HUDDistance = "Шлях"
	HUDBest     = "Рекорд"

	// ControlsHint is the on-screen controls reminder.
	ControlsHint = "A/D або стрілки — смуга · Пробіл/W — стрибок · X/Enter — атака · Esc — вихід"

	// Death scene joke lines, shown one at a time.
	DeathJokeOne   = "Норберт знову вдав мертвого. Цього разу — настільки переконливо, що й сам повірив."
	DeathJokeTwo   = "Потойбічна канцелярія вже оформила ваш обморок. Бланк №13, печатка синя."
	DeathJokeThree = "Клієнт почекає. У нас тут, зрештою, ціла вічність."

	// End scene title and stat labels.
	EndTitle     = "Зміну закрито"
	EndStamps    = "Марки"
	EndDistance  = "Дистанція"
	EndDeaths    = "Обмороки"
	EndCompleted = "Доставлено"

	// Enemy names.
	EnemyNameCopier  = "Ксерокс-ревізор"
	EnemyNameCabinet = "Шафа-мімік"
	EnemyNameCoffee  = "Кавова машина"
)
