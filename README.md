# go_user_micro

mock microservice to be used when I need to stub out a user service, will have full unit and integration tests, 
authentication and whatever extra bells and whistles I can think of.

This service only implments stateless auth, and should not be used where more granular control is needed. Anyone and 
everyone is free to use this service if they would like, it is well tested and functioning as intended. 

## TODO
- [x] Implement basic auth
- [x] Implement basic user creation
- [x] Implement basic user deletion
- [x] Implement basic user update
- [x] Implement basic user retrieval
- [x] Implement basic user retrieval by email
- [x] Implement panic recovery middleware
- [] Implement CORS
- [] General cleanup - tests are very verbose and repitive, I'm okay with that as they're also to document how to use the 
     service but there's also likely a better way
