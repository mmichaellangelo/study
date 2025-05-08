# disco API Schema

## Auth

### Register:  
```
POST /register
Content-Type: multipart/form-data
FormData:
    "email": email address
    "username" : username
    "password": password
```  
### Login:  
```
POST /login
Content-Type: multipart/form-data
FormData:
    "emailorusername": email address or username
    "password": password
Response if authenticated:
    Headers:
        "set-cookie": sets access and refresh tokens
```  
### Identity:
```
POST /me
credentials: include
Response if authenticated:
    Content-Type: application/json,
    Body:
        {
            "exp": access token expiration,
            "userid": account id,
            "username": username
        }
```


