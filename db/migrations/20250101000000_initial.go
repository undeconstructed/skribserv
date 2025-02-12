package migrations

import (
	"github.com/go-rel/rel"
)

func MigrateCreateInitial(schema *rel.Schema) {
	schema.CreateTable("uzantoj", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("nomo", rel.Required(true))
		t.String("retpoŝto", rel.Required(true))
		t.String("pasvorto", rel.Required(true))

		t.DateTime("kreita_je", rel.Required(true))
		t.DateTime("ŝanĝita_je", rel.Required(true))
	})

	schema.CreateTable("tekstoj", func(t *rel.Table) {
		t.String("id", rel.Primary(true))
		t.String("uzanto_id", rel.Required(true))
		t.Text("teksto", rel.Required(true))

		t.DateTime("kreita_je", rel.Required(true))
		t.DateTime("ŝanĝita_je", rel.Required(true))

		t.ForeignKey("uzanto_id", "uzantoj", "id", rel.OnDelete("cascade"))
	})
}

func RollbackCreateInitial(schema *rel.Schema) {
	schema.DropTable("tekstoj")

	schema.DropTable("uzantoj")
}
