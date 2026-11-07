// In-run texts: death causes, the stomp, the snip, the chase, the shield,
// the boss fight phases and the HUD labels, in English and modern academic
// Ukrainian. Every value is a string literal so the font step can scan the
// file.
package gameplay

// Line is one player-facing string in English and Ukrainian. The render
// layer picks a field by the campaign language setting.
type Line struct {
	EN string
	UA string
}

// CauseLines holds the deadpan narrator line for each death cause, indexed
// by the Cause value. CauseNone carries the line shown while nothing has
// failed yet.
var CauseLines = []Line{
	CauseNone: {
		EN: "Downald D is still running. The narrator is still not impressed.",
		UA: "Downald D досі біжить. Оповідач досі не вражений.",
	},
	CausePit: {
		EN: "\"That pit was fake.\" — Downald D, from inside the pit. The pit has been notified.",
		UA: "«Це була підроблена яма», — заявив Downald D просто з ями. Яму вже повідомлено.",
	},
	CauseBarrier: {
		EN: "\"The barrier was clearly in the wrong lane.\" — Downald D, wedged into a barrier.",
		UA: "«Цей бар'єр стояв не на тій смузі», — заявив Downald D, застрягши в бар'єрі.",
	},
	CauseEnemy: {
		EN: "\"That enemy started it.\" — Downald D. The enemy was not consulted.",
		UA: "«Першим почав той ворог», — заявив Downald D. У ворога не питали.",
	},
	CauseSnip: {
		EN: "Copycat won this round. Downald D is now aerodynamic.",
		UA: "Цей раунд виграв Копікет. Downald D тепер аеродинамічний.",
	},
}

// StompedMudskipper is shown every time the hero stomps Gary, the mangrove
// mudskipper. The gag is identical each time.
var StompedMudskipper = Line{
	EN: "Gary the mudskipper has been stomped. Gary's family has been notified. Again.",
	UA: "Мулистого стрибуна Гриця розчавлено. Казали ж йому: не ходи, Грицю.",
}

// SnipLine is shown when Copycat's scissors land.
var SnipLine = Line{
	EN: "Snip. Copycat took a lock. Downald D is fine. Downald D is not fine.",
	UA: "Чик! Копікет відчикрижив пасмо. Downald D тримається. Downald D не тримається.",
}

// ChaseLine is shown when Copycat closes in and the chase begins.
var ChaseLine = Line{
	EN: "Copycat is close. Downald D blames the controller, the cat and the wind.",
	UA: "Копікет уже близько. Downald D звинувачує контролер, кота і вітер.",
}

// ShieldLine is shown when a hairspray shield absorbs a hit.
var ShieldLine = Line{
	EN: "The hairspray held. Downald D credits the hairspray, not the reflexes.",
	UA: "Лак витримав. Downald D завдячує лаку, а не рефлексам.",
}

// BossPhaseLines holds one short Copycat line for each of the three fight
// phases, indexed by phase minus one.
var BossPhaseLines = []Line{
	{EN: "Copycat growls and clicks his scissors. Business hours are over.", UA: "Копікет гарчить і клацає ножицями. Прийом завершено."},
	{EN: "Copycat speeds up. Downald D speeds up his excuses.", UA: "Копікет пришвидшується. Downald D пришвидшує виправдання."},
	{EN: "Copycat is furious. The haircut is now personal.", UA: "Копікет розлючений. Тепер це особиста стрижка."},
}

// HUDCombs labels the comb counter.
var HUDCombs = Line{EN: "Combs", UA: "Гребінці"}

// HUDLives labels the lives counter.
var HUDLives = Line{EN: "Lives", UA: "Життя"}

// HUDShield labels the hairspray shield indicator.
var HUDShield = Line{EN: "Shield", UA: "Щит"}

// HUDChase labels the Copycat chase meter.
var HUDChase = Line{EN: "Chase", UA: "Погоня"}

// HUDDistance labels the distance counter.
var HUDDistance = Line{EN: "Distance", UA: "Відстань"}

// HUDPause labels the paused banner.
var HUDPause = Line{EN: "Pause", UA: "Пауза"}
