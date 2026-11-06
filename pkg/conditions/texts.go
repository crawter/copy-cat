package conditions

// SurvivalComplete is the player-facing line shown when the hero survives
// the set survival time. It is the PDD's survival completion line verbatim,
// so the wording stands once here and is not restated anywhere else.
const SurvivalComplete = "You survived. The district sends its regards, and an invoice for the chase."

// CollectionComplete is the player-facing line shown when the hero collects
// the set quota of jar lids. The PDD fixes no wording for it, so it is
// composed here in the gallows register and stays easy to change later.
const CollectionComplete = "You gathered every lid the district values. A grand haul, fit to be listed at full retail and sold for nothing."

// Fail is the player-facing line shown when the crew catches the hero.
// The PDD fixes no wording for it, so it is composed here in the gallows
// register and stays easy to change later.
const Fail = "The crew caught you. You are charged for the time spent catching you, plus a convenience fee."
