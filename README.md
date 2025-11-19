# Billing Engine
## Description
This is a billing engine that is designed to do:
* Loan schedule for a given loan
* Outstanding Amount for a given loan
* Status of weather the customer is Delinquent or not

### Tech Stack
* Language: Golang (1.23+)
* Database: MySQL 8.0
* Architecture: Clean Architecture / Hexagonal
* Containerization: Docker & Docker Compose

## Start Guide
### Prerequisites
* Docker
* docker-compose

### How to start
You can start the service using
```
make up
```

you can stop the service using
```
make down
```

## API Documentation
### Create Loan
Initializes a loan of 5,000,000 IDR for 50 weeks.\
Endpoint: `POST /loans/create`\
Body:
```
{
    "principal": 5000000,
    "weeks": 50
}
```
### Get Loan Status
Returns the outstanding balance, delinquency status, and full schedule.\
Endpoint: `GET /loans/status?id={loan_id}`

### Make Payment
Pays a specific installment. Must be exact amount (110,000 IDR for standard setup).\
Endpoint: `POST /loans/pay?id={loan_id}`\
Body:
```
{
    "amount": 110000
}
```

