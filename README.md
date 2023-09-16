# go_user_micro

mock microservice to be used when I need to stub out a user service, will have full unit and integration tests, 
authentication and whatever extra bells and whistles I can think of.

This service only implements stateless auth, and should not be used where more granular control is needed. Anyone and 
everyone is free to use this service if they would like, it is well tested and functioning as intended. 

I did my best to not rely on an env so any repo can plug and play this, but you will need to make email.env file and fill in your own smtp server info.

## Usage
1. Go mod install the repo.
2. Either create your own postgres instance or use the defaults in the dockerfile
3. Make migrate_up to run the migrations, requires <a href="https://github.com/golang-migrate/migrate">Go Migrate</a>
4. Configure the cors in the main.go file, or an env if you prefer
5. Configure your own smtp server in an email.env file, or change the name in the email.env
6. Make run to run the service
7. Make test to run the tests


## TODO
- [x] Implement basic auth
- [x] Implement basic user creation
- [x] Implement basic user deletion
- [x] Implement basic user update
- [x] Implement basic user retrieval
- [x] Implement basic user retrieval by email
- [x] Implement panic recovery middleware
- [x] Implement validation service
- [x] Implement Postgres connection
- [x] Implement Postgres test database
- [x] Implement Migrations
- [x] Implement Graceful shutdown
- [x] Implement email service
- [x] env variable management
- [x] Sending email as password reset
- [x] Password Reset as background go routine within handler
- [x] Helper function to recover from panics within background tasks
- [x] Implement CORS
- [x] General cleanup - tests are very verbose and repetitive, The mock handlers are now table tests, still deciding if I want them all like that.
- 

