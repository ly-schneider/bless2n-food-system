# rentro Go Backend

# Tools
- Echo
- Air
- Pgx
- Viper
- Testify
- Gotenv
- Godotenv
- GORM
- Atlas
- Flyway
- Docker

# Development

Run:
```bash
APP_ENV=development docker compose --env-file .env.development up --build
```

Generate SQL Migration for Flyway:
```bash
atlas migrate diff init_categories \
  --env gorm \
  --dev-url "docker://postgres/17/dev?search_path=public"
```

Generate oauth token from Auth0:
```bash
curl --request POST \
  --url https://YOUR_TENANT/oauth/token \
  --header 'content-type: application/json' \
  --data '{
     "grant_type":"client_credentials",
     "client_id":"YOUR_CLIENT_ID",
     "client_secret":"YOUR_CLIENT_SECRET",
     "audience":"YOUR_API_IDENTIFIER"
  }'

curl --request POST \
  --url https://dev-leys-services.eu.auth0.com/oauth/token \
  --header 'content-type: application/json' \
  --data '{
     "grant_type":"client_credentials",
     "client_id":"lvExokFKRdet5MvtBtakPA8j8ZDGDDSN",
     "client_secret":"xfYXiF3WGaPqWI3hy1jvmS7WijKL8OEf01TJCto5fujfJTL0yMAb-Bk-JM73ovfP",
     "audience":"https://dev-leys-services.eu.auth0.com/api/v2"
  }'

curl --request POST \
  --url 'https://dev-leys-services.eu.auth0.com/oauth/token' \
  --header 'content-type: application/x-www-form-urlencoded' \
  --data grant_type=client_credentials \
  --data client_id=lvExokFKRdet5MvtBtakPA8j8ZDGDDSN \
  --data client_secret=xfYXiF3WGaPqWI3hy1jvmS7WijKL8OEf01TJCto5fujfJTL0yMAb-Bk-JM73ovfP \
  --data audience=http://0.0.0.0:8080
```