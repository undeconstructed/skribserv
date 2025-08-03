package app

import "time"

type EntityResponse struct {
	Message string `json:"mesaĝo"`
	Entity  any    `json:"ento"`
}

type UserJSON struct {
	ID       DBID   `json:"id"`
	Name     string `json:"nomo,omitzero"`
	Email    string `json:"retpoŝto,omitzero"`
	Password string `json:"pasvorto,omitzero"`
	Admin    bool   `json:"admina,omitzero"`
}

type CourseJSON struct {
	ID    DBID      `json:"id"`
	Owner UserJSON  `json:"posedanto,omitzero"`
	Name  string    `json:"nomo,omitzero"`
	Time  time.Time `json:"kiamo,omitzero"`
}

type LessonJSON struct {
	ID     DBID       `json:"id"`
	Course CourseJSON `json:"kurso,omitzero"`
	Name   string     `json:"nomo,omitzero"`
	Time   time.Time  `json:"kiamo,omitzero"`
}

type LearnerJSON struct {
	ID     DBID       `json:"id"`
	Course CourseJSON `json:"kurso,omitzero"`
	User   UserJSON   `json:"uzanto,omitzero"`
}

type HomeworkJSON struct {
	ID      DBID     `json:"id"`
	Learner UserJSON `json:"lernanto,omitzero"`
	Text    string   `json:"teksto,omitzero"`
}
