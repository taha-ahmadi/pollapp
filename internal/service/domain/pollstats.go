package domain

type OptionStat struct {
	Option string
	Count  int
}

type PollStats struct {
	PollID int
	Votes  []OptionStat
}
