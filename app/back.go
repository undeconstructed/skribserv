package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-rel/rel"
	"github.com/go-rel/rel/where"
	"github.com/undeconstructed/skribserv/lib"
)

const adminEmail = "admin@admin"

type back struct {
	db  rel.Repository
	log lib.MakeContextLogger
}

func (a *back) EnsureAdmin(ctx context.Context, password string) (User, error) {
	user := &User{}

	err := a.db.Find(ctx, user, where.Eq("email", adminEmail))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			user1, err := a.putUser(ctx, User{
				Name:     "Admin User",
				Email:    adminEmail,
				Password: password,
				Admin:    true,
			})
			if err != nil {
				return User{}, err
			}

			a.log(ctx).Info("created admin user")

			return user1, nil
		}

		return User{}, fmt.Errorf("db (read): %w", err)
	}

	if user.Password == password {
		return *user, nil
	}

	err = a.db.Update(ctx, user, rel.Set("password", password))
	if err != nil {
		return User{}, fmt.Errorf("db (write): %w", err)
	}

	a.log(ctx).Info("reset admin password")

	return *user, nil
}

func (a *back) listUsers(ctx context.Context) ([]User, error) {
	var out []User

	err := a.db.FindAll(ctx, &out)
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) putUser(ctx context.Context, user0 User) (User, error) {
	if user0.ID == "" {
		user0.ID = makeRandomID("u", 5)
	}

	if err := a.db.Insert(ctx, &user0); err != nil {
		return User{}, err
	}

	return user0, nil
}

func (a *back) getUser(ctx context.Context, id DBID) (User, error) {
	user := &User{}

	err := a.db.Find(ctx, user, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return User{}, err
		}

		return User{}, fmt.Errorf("db (read): %w", err)
	}

	return *user, nil
}

func (a *back) getUserByLogin(ctx context.Context, email, password string) (User, error) {
	user := &User{}

	err := a.db.Find(ctx, user, where.Eq("email", email), where.Eq("password", password))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return User{}, err
		}

		return User{}, fmt.Errorf("db (read): %w", err)
	}

	return *user, nil
}

func (a *back) getUserByEmail(ctx context.Context, email string) (User, error) {
	user := &User{}

	err := a.db.Find(ctx, user, where.Eq("email", email))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return User{}, err
		}

		return User{}, fmt.Errorf("db (read): %w", err)
	}

	return *user, nil
}

func (a *back) listCourses(ctx context.Context) ([]Course, error) {
	var out []Course

	err := a.db.FindAll(ctx, &out)
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) putCourse(ctx context.Context, course Course) (Course, error) {
	if course.ID == "" {
		course.ID = makeRandomID("k", 5)
	}

	if err := a.db.Insert(ctx, &course); err != nil {
		return Course{}, err
	}

	return course, nil
}

func (a *back) getCourse(ctx context.Context, id DBID) (Course, error) {
	course := &Course{}

	err := a.db.Find(ctx, course, rel.Select("*", "owner_x.*").JoinAssoc("owner_x"), where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return Course{}, err
		}

		return Course{}, fmt.Errorf("db (read): %w", err)
	}

	if err := a.db.Preload(ctx, course, "lessons"); err != nil {
		return Course{}, err
	}

	return *course, nil
}

func (a *back) addUserToCourse(ctx context.Context, user, course DBID) (Learner, error) {
	learner := &Learner{
		ID:       makeRandomID("l", 5),
		UserID:   user,
		CourseID: course,
	}

	if err := a.db.Insert(ctx, learner); err != nil {
		return Learner{}, err
	}

	return *learner, nil
}

func (a *back) getLearnersByCourse(ctx context.Context, course DBID) ([]Learner, error) {
	var out []Learner

	err := a.db.FindAll(ctx, &out, where.Eq("course", course))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) getLearnersByUser(ctx context.Context, user DBID) ([]Learner, error) {
	var out []Learner

	err := a.db.FindAll(ctx, &out, rel.Select("*", "course_x.*").JoinAssoc("course_x"), where.Eq("user", user))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) putLesson(ctx context.Context, lesson Lesson) (Lesson, error) {
	if lesson.ID == "" {
		lesson.ID = makeRandomID("ke", 5)
	}

	if err := a.db.Insert(ctx, &lesson); err != nil {
		return Lesson{}, err
	}

	return lesson, nil
}

func (a *back) getLesson(ctx context.Context, id DBID) (Lesson, error) {
	lesson := &Lesson{}

	err := a.db.Find(ctx, lesson, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return Lesson{}, err
		}

		return Lesson{}, fmt.Errorf("db (read): %w", err)
	}

	return *lesson, nil
}

func (a *back) getLessonsForCourse(ctx context.Context, course DBID) ([]Lesson, error) {
	var out []Lesson

	err := a.db.FindAll(ctx, &out, where.Eq("course", course), rel.SortDesc("time"))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) putHomework(ctx context.Context, lernanto DBID, teksto string) (Homework, error) {
	homework1 := &Homework{
		ID:        makeRandomID("ht", 5),
		LearnerID: lernanto,
		Text:      teksto,
	}

	if err := a.db.Insert(ctx, homework1); err != nil {
		return Homework{}, err
	}

	return *homework1, nil
}

func (a *back) getHomeworksForUser(ctx context.Context, userID DBID) ([]Homework, error) {
	var out []Homework

	err := a.db.FindAll(ctx, &out, where.Eq("user", userID))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) getHomeworksForLesson(ctx context.Context, course, lessonID DBID) ([]Homework, error) {
	var out []Homework

	err := a.db.FindAll(ctx, &out, where.Eq("lesson", lessonID))
	if err != nil {
		return nil, fmt.Errorf("db (read): %w", err)
	}

	return out, nil
}

func (a *back) getHomework(ctx context.Context, id DBID) (Homework, error) {
	homework := &Homework{}

	err := a.db.Find(ctx, homework, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return Homework{}, err
		}

		return Homework{}, fmt.Errorf("db (read): %w", err)
	}

	return *homework, nil
}
