package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-rel/rel"
	"github.com/undeconstructed/skribserv/lib"
)

type front struct {
	back  *back
	ident *Authenticator
	log   lib.MakeContextLogger
}

func (a *front) Mount(mux lib.Router) {
	h := lib.APIHandler

	mux("POST", "/mi/ensaluti", h(a.Login))
	mux("POST", "/mi/elsaluti", h(a.Logout))
	mux("GET", "/mi", h(a.AboutMe), a.identify)

	mux("GET", "/uzantoj", h(a.GetUsers), a.forAdmin, a.identify)
	mux("POST", "/uzantoj", h(a.PostUsers), a.forAdmin, a.identify)
	mux("GET", "/uzantoj/{user}", h(a.GetUser), a.forAdminOrSelf, a.identify)

	mux("GET", "/kursoj", h(a.GetCourses), a.identify)
	mux("POST", "/kursoj", h(a.PostCourses), a.forAdmin, a.identify)
	mux("GET", "/kursoj/{course}", h(a.GetCourse), a.identify)

	mux("GET", "/kursoj/{course}/eroj", h(a.GetLessons), a.identify)
	mux("POST", "/kursoj/{course}/eroj", h(a.PostLessons), a.identify)
	mux("GET", "/kursoj/{course}/eroj/{lesson}", h(a.GetLesson), a.identify)

	mux("GET", "/kursoj/{course}/eroj/{lesson}/hejmtaskoj", h(a.GetHomeworksForCoursePart), a.identify)

	mux("POST", "/kursoj/{course}/lernantoj", h(a.PostLearners), a.forAdmin, a.identify)
	mux("GET", "/kursoj/{course}/lernantoj", h(a.GetLearners), a.identify)
	mux("GET", "/kursoj/{course}/lernantoj/{learner}", h(a.GetLearner), a.identify)

	mux("GET", "/uzantoj/{user}/kursoj", h(a.GetCoursesForUser), a.forAdminOrSelf, a.identify)

	mux("POST", "/uzantoj/{user}/hejmtaskoj", h(a.PostHomework), a.forAdminOrSelf, a.identify)
	mux("GET", "/uzantoj/{user}/hejmtaskoj", h(a.GetHomeworksForUser), a.forAdminOrSelf, a.identify)
	mux("GET", "/uzantoj/{user}/hejmtaskoj/{homework}", h(a.GetHomework), a.identify)
}

type ctxKey int

const ctxKeyUser ctxKey = 1

func (a *front) identify(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		type seancfn func() (*User, error)

		tryCookie := func() (*User, error) {
			if sessionCookie, er := r.Cookie("Seanco"); er == nil {
				sessionID := sessionCookie.Value

				user, er := a.ident.getSessionUser(sessionID)
				if er != nil {
					if errors.Is(er, ErrNoSession) {
						return nil, nil
					}

					return nil, lib.ErrHTTPUnauthorized
				}

				return user, nil
			}

			return nil, nil
		}

		tryHeader := func() (*User, error) {
			if email, password, ok := r.BasicAuth(); ok {
				user, er := a.back.getUserByEmail(ctx, email)
				if er != nil {
					if errors.Is(er, rel.ErrNotFound) {
						return nil, lib.ErrHTTPUnauthorized
					}

					return nil, er
				}

				if user.Password != password {
					return nil, lib.ErrHTTPUnauthorized
				}

				return &user, nil
			}

			return nil, nil
		}

		tryBasic := func() (*User, error) { return nil, nil }

		for _, f := range []seancfn{tryCookie, tryHeader, tryBasic} {
			user, er := f()
			if er != nil {
				lib.SendHTTPError(w, 0, er)
				return
			}

			if user != nil {
				ctx1 := context.WithValue(r.Context(), ctxKeyUser, user)
				r1 := r.WithContext(ctx1)

				a.log(ctx).Debug("auth", "user", user.ID)

				next(w, r1)

				return
			}
		}

		lib.SendHTTPError(w, 0, lib.ErrHTTPUnauthorized)
	}
}

func (a *front) forAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := a.userFromContext(r.Context())

		if !u.Admin {
			lib.SendHTTPError(w, 0, lib.ErrHTTPForbidden)
			return
		}

		next(w, r)
	}
}

func (a *front) forAdminOrSelf(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("user")

		u := a.userFromContext(r.Context())

		if !u.Admin && u.ID != DBID(userID) {
			lib.SendHTTPError(w, 0, lib.ErrHTTPForbidden)
			return
		}

		next(w, r)
	}
}

func (a *front) Login(ctx context.Context, r *http.Request) any {
	type loginReq struct {
		Email    string `json:"retpoŝto"`
		Password string `json:"pasvorto"`
	}

	req, err := DecodeBody(r, &loginReq{})
	if err != nil {
		return err
	}

	user, err := a.back.getUserByLogin(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return lib.HTTPResponse{Status: http.StatusForbidden}
		}
		return err
	}

	sID, err := a.ident.putSession(user)
	if err != nil {
		return err
	}

	sessionCookie := &http.Cookie{
		Name:    "Seanco",
		Value:   sID,
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	}

	return lib.HTTPResponse{
		Cookies: []*http.Cookie{sessionCookie},
		Data: EntityResponse{
			Message: "seanco " + sID,
			Entity: UserJSON{
				ID:    user.ID,
				Name:  user.Name,
				Admin: user.Admin,
			},
		},
	}
}

func (a *front) Logout(ctx context.Context, r *http.Request) any {
	cookie, err := r.Cookie("Seanco")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return lib.HTTPResponse{}
		}

		return err
	}

	sID := cookie.Value

	a.ident.deleteSession(sID)

	sessionCookie := &http.Cookie{
		Name:    "Seanco",
		Path:    "/",
		Expires: time.Time{},
	}

	return lib.HTTPResponse{
		Cookies: []*http.Cookie{sessionCookie},
		Data: EntityResponse{
			Message: "seanco " + sID,
		},
	}
}

func (a *front) userFromContext(ctx context.Context) *User {
	return ctx.Value(ctxKeyUser).(*User)
}

func (a *front) AboutMe(ctx context.Context, r *http.Request) any {
	user := ctx.Value(ctxKeyUser).(*User)

	return EntityResponse{
		Message: "uzanto",
		Entity: UserJSON{
			ID:    user.ID,
			Name:  user.Name,
			Admin: user.Admin,
		},
	}
}

func (a *front) GetUsers(ctx context.Context, r *http.Request) any {
	users, err := a.back.listUsers(ctx)
	if err != nil {
		return err
	}

	out := make([]UserJSON, 0, len(users))

	for _, u := range users {
		out = append(out, apiFromUser(u))
	}

	return EntityResponse{
		Message: "uzantoj",
		Entity:  out,
	}
}

func (a *front) PostUsers(ctx context.Context, r *http.Request) any {
	user0, err := DecodeBody(r, &UserJSON{})
	if err != nil {
		return err
	}

	user1, err := a.back.putUser(ctx, User{
		Name:     user0.Name,
		Email:    user0.Email,
		Password: user0.Password,
	})
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "nova uzanto",
		Entity:  apiFromUser(user1),
	}
}

func (a *front) GetUser(ctx context.Context, r *http.Request) any {
	userID := r.PathValue("user")
	if userID == "" {
		return lib.ErrHTTPNotFound
	}

	user, err := a.back.getUser(ctx, DBID(userID))
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "uzanto " + userID,
		Entity:  apiFromUser(user),
	}
}

func apiFromUser(in User) UserJSON {
	return UserJSON{
		ID:   in.ID,
		Name: in.Name,
	}
}

func (a *front) GetCourses(ctx context.Context, r *http.Request) any {
	courses, err := a.back.listCourses(ctx)
	if err != nil {
		return err
	}

	out := make([]CourseJSON, 0, len(courses))

	for _, k := range courses {
		out = append(out, apiFromCourse(k))
	}

	return EntityResponse{
		Message: "courses",
		Entity:  out,
	}
}

func (a *front) PostCourses(ctx context.Context, r *http.Request) any {
	user := a.userFromContext(ctx)

	course0, err := DecodeBody(r, &CourseJSON{})
	if err != nil {
		return err
	}

	owner := course0.Owner.ID
	if owner == "" {
		owner = user.ID
	}

	course1, err := a.back.putCourse(ctx, Course{
		OwnerID: owner,
		Name:    course0.Name,
		About:   course0.About,
		Time:    course0.Time,
	})
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "nova kurso",
		Entity:  apiFromCourse(course1),
	}
}

func (a *front) GetCourse(ctx context.Context, r *http.Request) any {
	courseID := r.PathValue("course")
	if courseID == "" {
		return lib.ErrHTTPNotFound
	}

	course, err := a.back.getCourse(ctx, DBID(courseID))
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "kurso",
		Entity:  apiFromCourse(course),
	}
}

func apiFromCourse(in Course) CourseJSON {
	return CourseJSON{
		ID: in.ID,
		Owner: UserJSON{
			ID:   in.OwnerX.ID,
			Name: in.OwnerX.Name,
		},
		Name:  in.Name,
		About: in.About,
		Time:  in.Time,
	}
}

func (a *front) GetLessons(ctx context.Context, r *http.Request) any {
	courseID := r.PathValue("course")
	if courseID == "" {
		return lib.ErrHTTPNotFound
	}

	lessons, err := a.back.getLessonsForCourse(ctx, DBID(courseID))
	if err != nil {
		return err
	}

	out := make([]LessonJSON, 0, len(lessons))

	for _, k := range lessons {
		out = append(out, apiFromLesson(k))
	}

	return EntityResponse{
		Message: "kurseroj de " + courseID,
		Entity:  out,
	}
}

func apiFromLesson(in Lesson) LessonJSON {
	return LessonJSON{
		ID: in.ID,
		Course: CourseJSON{
			ID: in.Course,
		},
		Name: in.Name,
		Time: in.Time,
	}
}

func (a *front) PostLessons(ctx context.Context, r *http.Request) any {
	courseID := r.PathValue("course")
	if courseID == "" {
		return lib.ErrHTTPNotFound
	}

	user := a.userFromContext(ctx)

	lesson0, err := DecodeBody(r, &LessonJSON{})
	if err != nil {
		return err
	}

	if lesson0.Course.ID != "" && lesson0.Course.ID != DBID(courseID) {
		return lib.ErrHTTPBadRequest
	}

	course, err := a.back.getCourse(ctx, DBID(courseID))
	if err != nil {
		return err
	}

	if course.OwnerID != user.ID && !user.Admin {
		return lib.ErrHTTPForbidden
	}

	lesson1, err := a.back.putLesson(ctx, Lesson{
		Course: course.ID,
		Name:   lesson0.Name,
		Time:   lesson0.Time,
	})
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "nova kursero",
		Entity: LessonJSON{
			ID:   lesson1.ID,
			Name: lesson1.Name,
			Course: CourseJSON{
				ID: course.ID,
			},
		},
	}
}

func (a *front) GetLesson(ctx context.Context, r *http.Request) any {
	return ErrUnimplemented
}

func (a *front) GetHomeworksForUser(ctx context.Context, r *http.Request) any {
	userID := r.PathValue("user")
	if userID == "" {
		return lib.ErrHTTPNotFound
	}

	homeworks, err := a.back.getHomeworksForUser(ctx, DBID(userID))
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "hejmtaskoj de " + userID,
		Entity:  homeworks,
	}
}

func (a *front) GetLearners(ctx context.Context, r *http.Request) any {
	courseID := r.PathValue("course")
	if courseID == "" {
		return lib.ErrHTTPNotFound
	}

	learners, err := a.back.getLearnersByCourse(ctx, DBID(courseID))
	if err != nil {
		return err
	}

	out := make([]LearnerJSON, 0, len(learners))

	for _, l := range learners {
		out = append(out, LearnerJSON{
			ID: l.ID,
			User: UserJSON{
				ID:   l.UserID,
				Name: l.UserX.Name,
			},
		})
	}

	return EntityResponse{
		Message: "lernantoj de " + courseID,
		Entity:  out,
	}
}

func (a *front) PostLearners(ctx context.Context, r *http.Request) any {
	courseID := r.PathValue("course")
	if courseID == "" {
		return lib.ErrHTTPNotFound
	}

	learner0, err := DecodeBody(r, &LearnerJSON{})
	if err != nil {
		return err
	}

	if learner0.Course.ID != "" && learner0.Course.ID != DBID(courseID) {
		return lib.ErrHTTPBadRequest
	}

	learner1, err := a.back.addUserToCourse(ctx, learner0.User.ID, DBID(courseID))
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "nova lernanto",
		Entity: LearnerJSON{
			ID: learner1.ID,
			User: UserJSON{
				ID: learner1.UserID,
			},
			Course: CourseJSON{
				ID: learner1.CourseID,
			},
		},
	}
}

func (a *front) GetLearner(ctx context.Context, r *http.Request) any {
	return ErrUnimplemented
}

func (a *front) GetCoursesForUser(ctx context.Context, r *http.Request) any {
	userID := r.PathValue("user")
	if userID == "" {
		return lib.ErrHTTPNotFound
	}

	courses, err := a.back.getLearnersByUser(ctx, DBID(userID))
	if err != nil {
		return err
	}

	out := make([]CourseJSON, 0, len(courses))

	for _, l := range courses {
		out = append(out, apiFromCourse(l.CourseX))
	}

	return EntityResponse{
		Message: "kursoj de " + userID,
		Entity:  out,
	}
}

func (a *front) PostHomework(ctx context.Context, r *http.Request) any {
	userID := r.PathValue("user")
	if userID == "" {
		return lib.ErrHTTPNotFound
	}

	user := a.userFromContext(ctx)

	if userID != string(user.ID) {
		return lib.ErrHTTPForbidden
	}

	homework0, err := DecodeBody(r, &HomeworkJSON{})
	if err != nil {
		return err
	}

	if homework0.Learner.ID != "" && homework0.Learner.ID != user.ID {
		return lib.ErrHTTPForbidden
	}

	if homework0.Text == "" {
		return fmt.Errorf("%w: mankas teksto", lib.ErrHTTPBadRequest)
	}

	homework1, err := a.back.putHomework(ctx, user.ID, homework0.Text)
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "nova hejmtasko",
		Entity: HomeworkJSON{
			ID: homework1.ID,
			Learner: UserJSON{
				ID: homework1.LearnerID,
			},
			Text: homework1.Text,
		},
	}
}

func (a *front) GetHomework(ctx context.Context, r *http.Request) any {
	homeworkID := r.PathValue("homework")

	homework, err := a.back.getHomework(ctx, DBID(homeworkID))
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "homework " + homeworkID,
		Entity: HomeworkJSON{
			ID: homework.ID,
			Learner: UserJSON{
				ID: homework.LearnerID,
			},
		},
	}
}

func (a *front) GetHomeworksForCoursePart(ctx context.Context, r *http.Request) any {
	course, lesson := r.PathValue("course"), r.PathValue("lesson")
	if course == "" || lesson == "" {
		return lib.ErrHTTPNotFound
	}

	homeworks, err := a.back.getHomeworksForLesson(ctx, DBID(course), DBID(lesson))
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "hejmtasko pri " + lesson,
		Entity:  homeworks,
	}
}

func DecodeBody[T any](r *http.Request, t *T) (*T, error) {
	dec := json.NewDecoder(r.Body)

	err := dec.Decode(t)
	if err != nil {
		return t, err
	}

	return t, nil
}
