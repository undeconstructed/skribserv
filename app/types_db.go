package app

import "time"

type DBID string

type User struct {
	ID       DBID
	Name     string
	Email    string
	Password string

	Admin bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) Table() string {
	return "users"
}

type Course struct {
	ID      DBID
	OwnerID DBID `db:"owner"`
	OwnerX  User `ref:"owner" fk:"id"`
	Name    string
	Time    time.Time

	Lessons []Lesson `ref:"id" fk:"course"`
}

func (Course) Table() string {
	return "courses"
}

type Lesson struct {
	ID     DBID
	Course DBID
	Name   string
	Time   time.Time
}

func (Lesson) Table() string {
	return "lessons"
}

type Learner struct {
	ID       DBID
	UserID   DBID   `db:"user"`
	UserX    User   `ref:"user" fk:"id"`
	CourseID DBID   `db:"course"`
	CourseX  Course `ref:"course" fk:"id"`
}

func (Learner) Table() string {
	return "learners"
}

type Teacher struct {
	ID       DBID
	UserID   DBID `db:"user"`
	UserX    User `ref:"user" fk:"id"`
	CourseID DBID `db:"course"`
}

func (Teacher) Table() string {
	return "teachers"
}

type Homework struct {
	ID        DBID
	LearnerID DBID   `db:"learner"`
	LessonID  DBID   `db:"lesson"`
	LessonX   Lesson `ref:"lesson" fk:"id"`
	Text      string
}

func (Homework) Table() string {
	return "homeworks"
}
