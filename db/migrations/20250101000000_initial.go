package migrations

import (
	"github.com/go-rel/rel"
)

func MigrateCreateInitial(schema *rel.Schema) {
	// uzantoj: detaloj bezonataj por ke iu ensalutu
	schema.CreateTable("uzantoj", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("nomo", rel.Required(true))
		t.String("retpoŝto", rel.Required(true), rel.Unique(true))
		t.String("pasvorto", rel.Required(true))

		t.Bool("admina", rel.Required(true), rel.Default(false))

		t.DateTime("kreita_je", rel.Required(true))
		t.DateTime("ŝanĝita_je", rel.Required(true))
	})

	// kursoj: difinas ke iu sinsekvo de lecionoj okazu
	schema.CreateTable("kursoj", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("posedanto", rel.Required(true))
		t.String("nomo", rel.Required(true))

		t.DateTime("kiamo", rel.Required(true))

		t.ForeignKey("posedanto", "uzantoj", "id", rel.OnDelete("cascade"))
	})

	// kurseroj: eroj da kurso, ekz. unu leciono
	schema.CreateTable("kurseroj", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("kurso", rel.Required(true))
		t.String("nomo", rel.Required(true))

		t.DateTime("kiamo", rel.Required(true))

		t.ForeignKey("kurso", "kursoj", "id", rel.OnDelete("cascade"))
	})

	// lernantoj: por ke uzantoj partoprenu kurson
	schema.CreateTable("lernantoj", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("uzanto", rel.Required(true))
		t.String("kurso", rel.Required(true))

		t.Unique([]string{"uzanto", "kurso"})

		t.ForeignKey("uzanto", "uzantoj", "id", rel.OnDelete("cascade"))
		t.ForeignKey("kurso", "kursoj", "id", rel.OnDelete("cascade"))
	})

	// instruistoj: por ke uzantoj instruu kurson
	schema.CreateTable("instruistoj", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("uzanto", rel.Required(true))
		t.String("kurso", rel.Required(true))

		t.ForeignKey("uzanto", "uzantoj", "id", rel.OnDelete("cascade"))
		t.ForeignKey("kurso", "kursoj", "id", rel.OnDelete("cascade"))
	})

	// hejmtaskoj: ensendita respondoj pri korektado
	schema.CreateTable("hejmtaskoj", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("lernanto", rel.Required(true))
		t.String("kursero", rel.Required(true))

		t.Text("teksto", rel.Required(true))

		t.ForeignKey("lernanto", "lernantoj", "id", rel.OnDelete("cascade"))
		t.ForeignKey("kursero", "kurseroj", "id", rel.OnDelete("cascade"))
	})
}

func RollbackCreateInitial(schema *rel.Schema) {
	schema.DropTable("hejmtaskoj")

	schema.DropTable("instruistoj")

	schema.DropTable("lernantoj")

	schema.DropTable("kurseroj")

	schema.DropTable("kursoj")

	schema.DropTable("uzantoj")
}
