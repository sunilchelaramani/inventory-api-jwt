# Inventory Management API with JWT security
Simple Inventory Management API example with JWT
- Project is built using Gorilla mux and 
- You will need to setup MySQL DB, structure is available under inventory.sql, update constants under constants.go
- Update your secret key under jwt.go
- For simplicity user and password is hardcoded in jwt.go
- You can change port

## Usage

To get JWT token

curl -v -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"password"}' http://localhost:8080/login

To use JWT token

curl -v -H "Authorization: bearer_token_goes_here" http://localhost:8080/products

## License

[MIT](https://choosealicense.com/licenses/mit/)

