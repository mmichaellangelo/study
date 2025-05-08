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
```  
### Identity:
```
POST /me
credentials: include
Response if auth:
    {
        ""
    }
```


