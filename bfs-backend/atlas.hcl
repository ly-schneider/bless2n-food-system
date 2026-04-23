env "local" {
  dev = "docker://postgres/18/dev?search_path=public"
  migration {
    dir = "file://db/migrations"
  }
}

env "ci" {
  dev = "docker://postgres/18/dev?search_path=public"
  migration {
    dir = "file://db/migrations"
  }
}

env "deploy" {
  url = getenv("DATABASE_URL")
  migration {
    dir              = "file://db/migrations"
    exec_order       = LINEAR
    revisions_schema = "public"
  }
}
