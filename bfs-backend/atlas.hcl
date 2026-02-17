data "external_schema" "ent" {
  program = [
    "go",
    "run",
    "ariga.io/atlas-provider-ent@latest",
    "--path", "./internal/schema",
    "--dialect", "postgres",
  ]
}

data "composite_schema" "app" {
  schema "app" {
    url = data.external_schema.ent.url
  }
  schema "app" {
    url = "file://db/schema/auth.sql"
  }
}

env "local" {
  src = data.composite_schema.app.url
  dev = "docker://postgres/18/dev?search_path=app"
  migration {
    dir = "file://db/migrations"
  }
}

env "ci" {
  src = data.composite_schema.app.url
  dev = "docker://postgres/18/dev?search_path=app"
  migration {
    dir = "file://db/migrations"
  }
}
