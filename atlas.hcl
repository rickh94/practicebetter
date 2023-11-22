variable "dbpath" {
  type = string
  default = getenv("DB_PATH")
}
env "dev" {
	src = "./sql/schema.sql"
	url = "sqlite://${var.dbpath}?_fk=1&_journal=WAL&_timeout=5000&_synchronous=normal"
	dev = "sqlite://dev?mode=memory&_journal=WAL&_timeout=5000&_fk=1"
	migration {
		dir = "file://sql/migrations"
		format = atlas
	}
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
