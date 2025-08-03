package migrations

import (
	"github.com/go-rel/rel"
)

func MigrateCreateInitial(schema *rel.Schema) {
	// users: detaloj bezonataj por ke iu ensalutu
	schema.CreateTable("users", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("name", rel.Required(true))
		t.String("email", rel.Required(true), rel.Unique(true))
		t.String("password", rel.Required(true))

		t.Bool("admin", rel.Required(true), rel.Default(false))

		t.DateTime("created_at", rel.Required(true))
		t.DateTime("updated_at", rel.Required(true))
	})

	// courses: difinas ke iu sinsekvo de lecionoj okazu
	schema.CreateTable("courses", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("owner", rel.Required(true))
		t.String("name", rel.Required(true))

		t.DateTime("time", rel.Required(true))

		t.ForeignKey("owner", "users", "id", rel.OnDelete("cascade"))
	})

	// lessons: eroj da course, ekz. unu leciono
	schema.CreateTable("lessons", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("course", rel.Required(true))
		t.String("name", rel.Required(true))

		t.DateTime("time", rel.Required(true))

		t.ForeignKey("course", "courses", "id", rel.OnDelete("cascade"))
	})

	// learners: por ke users partoprenu kurson
	schema.CreateTable("learners", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("user", rel.Required(true))
		t.String("course", rel.Required(true))

		t.Unique([]string{"user", "course"})

		t.ForeignKey("user", "users", "id", rel.OnDelete("cascade"))
		t.ForeignKey("course", "courses", "id", rel.OnDelete("cascade"))
	})

	// teachers: por ke users instruu kurson
	schema.CreateTable("teachers", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("user", rel.Required(true))
		t.String("course", rel.Required(true))

		t.ForeignKey("user", "users", "id", rel.OnDelete("cascade"))
		t.ForeignKey("course", "courses", "id", rel.OnDelete("cascade"))
	})

	// homeworks: ensendita respondoj pri korektado
	schema.CreateTable("homeworks", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("learner", rel.Required(true))
		t.String("lesson", rel.Required(true))

		t.Text("teksto", rel.Required(true))

		t.ForeignKey("learner", "learners", "id", rel.OnDelete("cascade"))
		t.ForeignKey("lesson", "lessons", "id", rel.OnDelete("cascade"))
	})
}

func RollbackCreateInitial(schema *rel.Schema) {
	schema.DropTable("homeworks")

	schema.DropTable("teachers")

	schema.DropTable("learners")

	schema.DropTable("lessons")

	schema.DropTable("courses")

	schema.DropTable("users")
}
