# disco API Schema

## ERROR RESPONSE
For any request that returns an error, the response body may contain an app-specific error code:
```
{
  errcode: <error code>
}
```
### Registration Error Codes:
- ACCOUNT_EMAIL_EXISTS
  - an account with the given email already exists
- ACCOUNT_USERNAME_EXISTS
  - an account with the given username already exists
- BAD_REGISTRATION_INFO
  - invalid registration info provided
- BAD_EMAIL
  - invalid email address provided

### Access Error Codes:
- NO_ACCESS_TOKEN
  - request did not provide access token
- NO_REFRESH_TOKEN
  - request did not provide refresh token
- TOKEN_EXPIRED
  - the provided token has expired
- TOKEN_INVALID
  - the provided token is not valid
- REFRESH_TOKEN_INVALIDATED
  - the provided refresh token is old and has been invalidated
- BAD_AUTH_HEADER
  - provided auth header has invalid format
- BAD_CLAIMS
  - token's embedded claims are invalid
- PASSWORD_INCORRECT
  - the provided password is not correct for the given account
- ACCESS_NOT_ALLOWED
  - account is not authorized to view the requested resource
- PASSWORD_AUTH_REQUIRED
  - basic authentication header is required for this request and was not provided or was invalid
- NOT_FOUND
  - the requested resource could not be found
- ILLEGAL_ARGUMENT
  - an invalid argument was provided (general)

### Internal Error Codes:
- INTERNAL_ERROR
  - an internal error occurred
- DATABASE_ERROR
  - an internal database error occurred


## AUTH

### Identity
#### POST /auth/me
```
Headers:  
  Cookie: access, refresh
```
#### Response if authenticated:  
```
Content-Type: application/json

Body:
  {  
    "userid": account id,  
    "username": username,  
    "exp": access token expiration,  
    < other embedded JWT rows >  
  } 
```

---
### Register:  
#### POST /auth/register    
``` 
Content-Type: multipart/form-data

FormData:
  "email": email address
  "username" : username
  "password": password
```
#### Response if accepted:
```
[200 OK]

Headers: 
  set-cookie: 
    refresh=<refresh token> 
    access=<access token>
```

---
### Login:  
#### POST /auth/login  
  
```
Content-Type: multipart/form-data

FormData:
    "emailorusername": email or username
    "password": password
```
#### Response if authenticated:
```
[200 OK]
Headers: 
  set-cookie: sets access and refresh 
  cookies for future api requests
```

---
### Logout:
#### POST /auth/logout
```
Headers:
  Cookie: refresh=<refresh token>
```
#### Response if accepted:
```
[200 OK]
Headers:
  set-cookie: 
    access=<expired> 
    refresh=<expired>
```

## ACCOUNT

### Create account
***see Auth | Register***

---
### Get account by ID
#### GET /accounts/[id]
```
Headers:
  Cookie:
    access=<access token>
    refresh=<refresh token>
```
#### Response if accepted:
```
[200 OK]
account: {
  id: account id,
  email: account email,
  username: account username,
  picture: url of profile picture,
  bio: account bio,
  created: time account was created
}
```
### Update account
#### PATCH /accounts/[id]
```
< OMIT UNCHANGED FIELDS >

account: {
  new_email: new email 
  new_username: new username
  new_picture: url of profile picture
  new_bio: new bio
}
```
#### Response if accepted:
```
[200 OK]
account: {
  id: account id,
  email: account email,
  username: account username,
  picture: url of profile picture,
  bio: account bio,
  created: time account was created
}
```

### Delete account
#### DELETE /accounts/[id]
***Auth tokens AND Basic Authorization header are required for account deletion.***
```
Headers:
  Authorization: Basic <Base64<username:password>>
  Cookie:
    access=<access token>
    refresh=<refresh token>
```
#### Response if accepted:
```
[200] OK
Headers:
  set-cookie: 
    access=<expired> 
    refresh=<expired>
```

## SET

