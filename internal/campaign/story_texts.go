// Story texts for the campaign: the between-world story cards, the ending,
// the useless loading tips, the medals of shame, the per-cause epitaphs and
// the credits, in English and modern academic Ukrainian. Every value is a
// string literal so the font step can scan the file; the Line type comes
// from texts.go.
package campaign

import "copycat/internal/gameplay"

// StoryCards holds one mockumentary paragraph per StoryKind, in StoryKind
// order: the world 2 opening, the world 3 opening and the ending. The ending
// pays off the D joke without revealing what the D stands for.
var StoryCards = []Line{
	StoryWorld2: {
		EN: "The proboscis monkey, Downald D, has escaped the mangroves. Gary's family has finally stopped filing paperwork. Now the rusty harbor opens ahead, a place where even the seagulls keep receipts. The narrator would wish him luck. The narrator has seen the containers.",
		UA: "Носач Downald D утік із мангрів. Родина Гриця нарешті припинила подавати документи. Попереду розкривається іржава гавань — місце, де навіть чайки зберігають чеки. Оповідач побажав би йому удачі. Оповідач бачив ті контейнери.",
	},
	StoryWorld3: {
		EN: "Chic-Chic Street lights Downald D's hair like a very confident traffic light. Nobody here knows what the D stands for, and the narrator has stopped asking. Copycat is close, scissors first. The pity, as always, is genuine.",
		UA: "Вулиця Чик-Чик підсвічує зачіску Downald D, наче дуже самовпевнений світлофор. Ніхто тут не знає, що означає «D», і оповідач уже не питає. Копікет поруч, ножиці напоготові. Жаль, як завжди, щирий.",
	},
	StoryEnding: {
		EN: "And so Downald D outran the cat, the scissors and several very patient obstacles. He is the greatest runner in the history of running, according to the only source he trusts: himself. The D remains a secret. Some say Dazzling. The narrator says nothing, but the narrator's eyebrow says everything.",
		UA: "І так Downald D обігнав кота, ножиці та кілька дуже терплячих перешкод. Він — найкращий бігун в історії бігу, за словами єдиного джерела, якому він довіряє: себе самого. Таємниця «D» не розкрита. Хтось каже — «Дивовижний». Оповідач мовчить, але брова оповідача каже все.",
	},
}

// LoadingTips holds useless loading advice, shown one at a time while a level
// loads. The D joke and the hairspray warning are the two the runner always
// ignores.
var LoadingTips = []Line{
	{
		EN: "Nobody knows what the D stands for. Downald says \"Dazzling\". The pits say \"Down\".",
		UA: "Ніхто не знає, що означає «D». Downald каже: «Дивовижний». Ями кажуть: «Донизу».",
	},
	{
		EN: "Hairspray protects you from one hit. It does not protect you from your decisions.",
		UA: "Лак захищає від одного удару. Він не захищає від ваших рішень.",
	},
	{
		EN: "Gary has been stomped again. Loading will not help Gary.",
		UA: "Гриця знову розчавили. Завантаження Грицеві не допоможе.",
	},
	{
		EN: "A pit is not a shortcut. A pit is a strong opinion.",
		UA: "Яма — це не короткий шлях. Яма — це тверда думка.",
	},
	{
		EN: "Mirrors make you faster, but only if you stop looking.",
		UA: "Дзеркала роблять вас швидшими, але лише якщо ви перестанете в них дивитися.",
	},
	{
		EN: "Copycat copies everyone. That is why he wants Downald D's hair. The D is still a trade secret.",
		UA: "Копікет мавпує всіх підряд. Тому йому й заманулося зачіски Downald D. Літера «D» досі комерційна таємниця.",
	},
}

// MedalsOfShame holds one mocking medal per gameplay.Cause, indexed by the
// Cause value. CauseNone carries the medal shown while the run has not
// failed yet.
var MedalsOfShame = []Line{
	gameplay.CauseNone: {
		EN: "No medal of shame yet. Downald D is merely warming up his disasters.",
		UA: "Медалі сорому ще немає. Downald D лише розминається перед катастрофами.",
	},
	gameplay.CausePit: {
		EN: "Medal of Shame: Pit Scholar. He fell in, then explained that the pit started it.",
		UA: "Медаль сорому «Дослідник ям». Упав, а потім пояснив, що першою почала яма.",
	},
	gameplay.CauseBarrier: {
		EN: "Medal of Shame: Wallflower. He met the barrier face-first and called it rude.",
		UA: "Медаль сорому «Лобова атака». Познайомився з бар'єром обличчям і назвав його невихованим.",
	},
	gameplay.CauseEnemy: {
		EN: "Medal of Shame: Gary Lobbyist. Beaten by something that lives in mud and never apologizes.",
		UA: "Медаль сорому «Захисник Гриця». Переможений істотою, що живе в мулі й ніколи не вибачається.",
	},
	gameplay.CauseSnip: {
		EN: "Medal of Shame: Philanthropist. He donated his hair to a cat holding scissors.",
		UA: "Медаль сорому «Філантроп». Пожертвував зачіску коту з ножицями.",
	},
}

// Epitaphs holds one game-over epitaph per gameplay.Cause, indexed by the
// Cause value. Every epitaph names Downald D in Latin. CauseNone carries the
// epitaph shown when the narrator refuses to wait for an actual death.
var Epitaphs = []Line{
	gameplay.CauseNone: {
		EN: "Here lies Downald D. Not dead yet, but the narrator is a planner.",
		UA: "Тут спочиває Downald D. Ще не помер, але оповідач планує наперед.",
	},
	gameplay.CausePit: {
		EN: "Here lies Downald D. He argued with a hole. The hole won.",
		UA: "Тут спочиває Downald D. Сперечався з ямою. Яма перемогла.",
	},
	gameplay.CauseBarrier: {
		EN: "Here lies Downald D. The barrier stopped him, and his excuses stopped nothing.",
		UA: "Тут спочиває Downald D. Бар'єр зупинив його, а виправдання не зупинили нічого.",
	},
	gameplay.CauseEnemy: {
		EN: "Here lies Downald D. Outsmarted by something smaller, muddier and better rested.",
		UA: "Тут спочиває Downald D. Перехитрило його щось менше, брудніше й більш відпочиле.",
	},
	gameplay.CauseSnip: {
		EN: "Here lies Downald D. He left behind a great haircut and no answers.",
		UA: "Тут спочиває Downald D. Після нього лишилася чудова стрижка і жодної відповіді.",
	},
}

// Credits holds the credit roll entries in display order: the roles of the
// game, played by nobody in particular.
var Credits = []Line{
	{EN: "Downald D — hair, heroism and lifelong excuse-making", UA: "Downald D — зачіска, героїзм і довічні виправдання"},
	{EN: "Copycat — villain, barber and very fast with scissors", UA: "Копікет — лиходій, цирульник і дуже швидкий з ножицями"},
	{EN: "Gary — the mudskipper who never learns", UA: "Гриць — мулистий стрибун, який ніколи не вчиться"},
	{EN: "The Narrator — pity, delivered in a monotone", UA: "Оповідач — жаль, виголошений монотонно"},
	{EN: "The D — a secret the writer kept on purpose", UA: "Літера «D» — таємниця, яку сценарист зберіг навмисно"},
	{EN: "Pits — the real winners of every level", UA: "Ями — справжні переможці кожного рівня"},
}

// CreditsClosing is the final line of the credit roll.
var CreditsClosing = Line{
	EN: "No monkeys were harmed. Gary was consulted once and declined to comment.",
	UA: "Жодна мавпа не постраждала. Гриця запитали один раз — він відмовився від коментаря.",
}
