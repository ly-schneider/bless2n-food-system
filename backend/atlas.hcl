data "external_schema" "gorm" {
  program = [
    "go", "run", "-mod=mod",
    "ariga.io/atlas-provider-gorm", "load",
    "--path", "./models/models",
    "--dialect", "postgres"
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/17/dev?search_path=public"
  migration {
    dir = "file://migrations?format=flyway"
  }
  format {
    migrate { diff = "{{ sql . \"  \" }}" }
  }
}
