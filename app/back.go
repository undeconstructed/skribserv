package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-rel/rel"
	"github.com/go-rel/rel/where"
	"github.com/undeconstructed/skribserv/lib"
)

type back struct {
	db  rel.Repository
	log func(context.Context) *lib.Logger
}

func (a *back) listUsers(ctx context.Context) ([]*User, error) {
	var out []*User

	err := a.db.FindAll(ctx, &out)
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) putUser(ctx context.Context, id DBID, name, email, password string) (*User, error) {
	if id == "" {
		id = makeRandomID("u", 5)
	}

	user1 := &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
	}

	if err := a.db.Insert(ctx, user1); err != nil {
		return nil, err
	}

	return user1, nil
}

func (a *back) getUser(ctx context.Context, id DBID) (*User, error) {
	user := &User{}

	err := a.db.Find(ctx, user, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (read): %w", err)
	}

	return user, nil
}

func (a *back) getUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}

	err := a.db.Find(ctx, user, where.Eq("email", email))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (read): %w", err)
	}

	return user, nil
}

func (a *back) listCourses(ctx context.Context) ([]*Course, error) {
	var out []*Course

	err := a.db.FindAll(ctx, &out)
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) putCourse(ctx context.Context, id DBID, owner DBID, name string, time time.Time) (*Course, error) {
	if id == "" {
		id = makeRandomID("k", 5)
	}

	course1 := &Course{
		ID:      id,
		Name:    name,
		OwnerID: owner,
		Time:    time,
	}

	if err := a.db.Insert(ctx, course1); err != nil {
		return nil, err
	}

	return course1, nil
}

func (a *back) getCourse(ctx context.Context, id DBID) (*Course, error) {
	course := &Course{}

	err := a.db.Find(ctx, course, rel.Select("*", "owner_x.*").JoinAssoc("owner_x"), where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (read): %w", err)
	}

	if err := a.db.Preload(ctx, course, "lessons"); err != nil {
		return nil, err
	}

	return course, nil
}

func (a *back) putLearner(ctx context.Context, id, user, course DBID) (*Learner, error) {
	if id == "" {
		id = makeRandomID("u", 5)
	}

	learner1 := &Learner{
		ID:       id,
		UserID:   user,
		CourseID: course,
	}

	if err := a.db.Insert(ctx, learner1); err != nil {
		return nil, err
	}

	return learner1, nil
}

func (a *back) getLearnersByCourse(ctx context.Context, course DBID) ([]*Learner, error) {
	var out []*Learner

	err := a.db.FindAll(ctx, &out, where.Eq("course", course))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) getLearnersByUser(ctx context.Context, user DBID) ([]*Learner, error) {
	var out []*Learner

	err := a.db.FindAll(ctx, &out, rel.Select("*", "course_x.*").JoinAssoc("course_x"), where.Eq("user", user))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) putLesson(ctx context.Context, id, course DBID, nomo string, kiamo time.Time) (*Lesson, error) {
	if id == "" {
		id = makeRandomID("ke", 5)
	}

	coursePart1 := &Lesson{
		ID:     id,
		Course: course,
		Name:   nomo,
		Time:   kiamo,
	}

	if err := a.db.Insert(ctx, coursePart1); err != nil {
		return nil, err
	}

	return coursePart1, nil
}

func (a *back) getLesson(ctx context.Context, id DBID) (*Lesson, error) {
	coursePart := &Lesson{}

	err := a.db.Find(ctx, coursePart, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (read): %w", err)
	}

	return coursePart, nil
}

func (a *back) getLessonsForCourse(ctx context.Context, course DBID) ([]*Lesson, error) {
	var out []*Lesson

	err := a.db.FindAll(ctx, &out, where.Eq("course", course), rel.SortDesc("time"))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) putHomework(ctx context.Context, id, lernanto DBID, teksto string) (*Homework, error) {
	if id == "" {
		id = makeRandomID("ht", 5)
	}

	homework1 := &Homework{
		ID:        id,
		LearnerID: lernanto,
		Text:      teksto,
	}

	if err := a.db.Insert(ctx, homework1); err != nil {
		return nil, err
	}

	return homework1, nil
}

func (a *back) getHomeworksForUser(ctx context.Context, userID DBID) ([]*Homework, error) {
	var out []*Homework

	err := a.db.FindAll(ctx, &out, where.Eq("user", userID))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) getHomeworksForLesson(ctx context.Context, course, coursePart DBID) ([]*Homework, error) {
	var out []*Homework

	err := a.db.FindAll(ctx, &out, where.Eq("lesson", coursePart))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) getHomework(ctx context.Context, id DBID) (*Homework, error) {
	homework := &Homework{}

	err := a.db.Find(ctx, homework, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (read): %w", err)
	}

	return homework, nil
}
