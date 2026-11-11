// Menu and screen texts for the campaign, in English and modern academic
// Ukrainian. Every value is a string literal so the font step can scan the
// file; the story texts reuse the Line type defined here.
package campaign

// Line is one player-facing string in English and Ukrainian. The render
// layer picks a field by the campaign language setting.
type Line struct {
	EN string
	UA string
}

// GameTitle is the main menu's heading.
var GameTitle = Line{EN: "Hair Raid", UA: "Hair Raid"}

// TitleTagline is the one-line pitch shown under the main menu heading.
var TitleTagline = Line{EN: "Outrun the cat. Keep the hair.", UA: "Випереди кота. Збережи зачіску."}

// MainMenuItems holds the main menu entries in selection order.
var MainMenuItems = []Line{
	{EN: "Continue", UA: "Продовжити"},
	{EN: "New Game", UA: "Нова гра"},
	{EN: "Map", UA: "Мапа"},
	{EN: "Endless Run", UA: "Нескінченний забіг"},
	{EN: "Settings", UA: "Налаштування"},
	{EN: "Quit", UA: "Вийти"},
}

// NewGameConfirmTitle is the New Game confirmation heading.
var NewGameConfirmTitle = Line{EN: "Start a New Game?", UA: "Почати нову гру?"}

// NewGameConfirmItems holds the New Game confirmation choices in selection
// order.
var NewGameConfirmItems = []Line{
	{EN: "Yes", UA: "Так"},
	{EN: "No", UA: "Ні"},
}

// MapTitle is the level selection map heading.
var MapTitle = Line{EN: "Choose a Level", UA: "Оберіть рівень"}

// MapLocked marks a level the player has not unlocked yet.
var MapLocked = Line{EN: "Locked", UA: "Заблоковано"}

// LevelNames holds the ten campaign level names in level ID order. The last
// one is the Copycat boss level.
var LevelNames = []Line{
	{EN: "The Swamp That Waves Back", UA: "Болото, що махає у відповідь"},
	{EN: "Root Canal", UA: "Кореневий канал"},
	{EN: "Gary's Neighborhood", UA: "Грицеве подвір'я"},
	{EN: "Containers of Concern", UA: "Контейнери тривоги"},
	{EN: "Gull Trouble", UA: "Чайкові клопоти"},
	{EN: "The Rat Race", UA: "Щурячі перегони"},
	{EN: "Chairs on Wheels", UA: "Крісла на колесах"},
	{EN: "Vacuums of Doom", UA: "Пилососи приреченості"},
	{EN: "The Street That Never Dims", UA: "Вулиця, що ніколи не гасне"},
	{EN: "The Last Haircut", UA: "Остання стрижка"},
}

// SettingsTitle is the settings screen heading.
var SettingsTitle = Line{EN: "Settings", UA: "Налаштування"}

// SettingsItems holds the settings entries in selection order.
var SettingsItems = []Line{
	{EN: "Music Volume", UA: "Гучність музики"},
	{EN: "Sound Volume", UA: "Гучність звуків"},
	{EN: "Fullscreen", UA: "Повноекранний режим"},
	{EN: "Language", UA: "Мова"},
}

// LanguagePickTitle is the first-run language choice heading.
var LanguagePickTitle = Line{EN: "Choose Your Language", UA: "Оберіть мову"}

// LanguageEnglishName is the English language choice, written as the
// language calls itself.
var LanguageEnglishName = Line{EN: "English", UA: "English"}

// LanguageUkrainianName is the Ukrainian language choice, written as the
// language calls itself.
var LanguageUkrainianName = Line{EN: "Ukrainian", UA: "Українська"}

// PauseTitle is the pause screen heading.
var PauseTitle = Line{EN: "Paused", UA: "Пауза"}

// PauseHint tells the player how to leave the pause screen.
var PauseHint = Line{EN: "Press P to resume, or choose an action", UA: "Натисніть P, щоб продовжити, або оберіть дію"}

// PauseItems holds the pause choices in selection order.
var PauseItems = []Line{
	{EN: "Resume", UA: "Продовжити гру"},
	{EN: "Restart", UA: "Спробувати знову"},
	{EN: "Back to Menu", UA: "Назад у меню"},
}

// LevelEndTitle is the campaign level-end heading.
var LevelEndTitle = Line{EN: "Level Complete", UA: "Рівень пройдено"}

// LevelEndHint tells the player how to continue past the level end.
var LevelEndHint = Line{EN: "Confirm to continue", UA: "Підтвердіть, щоб продовжити"}

// EndlessLevelEndTitle is the endless run's end heading.
var EndlessLevelEndTitle = Line{EN: "Run Over", UA: "Забіг завершено"}

// GameOverTitle is the game-over heading.
var GameOverTitle = Line{EN: "Game Over", UA: "Гру завершено"}

// GameOverEpitaph is the deadpan line under the game-over heading.
var GameOverEpitaph = Line{
	EN: "Here lies Downald D. Tremendous hair. Bad at pits.",
	UA: "Тут спочиває Downald D. Зачіска — розкішна. З ямами не склалося.",
}

// GameOverItems holds the game-over choices in selection order.
var GameOverItems = []Line{
	{EN: "Retry", UA: "Спробувати знову"},
	{EN: "Menu", UA: "Меню"},
}
