# Register

## Endpoint
`POST /register`

## URI
`http://18.182.22.91:8080/register`

## Description
Check for existing entries and create a new entry in the `auth` database.

## Request

### Headers
- `Content-Type: application/json`

### Body
The request body should contain a JSON object with the following fields:

| Field  | Type   | Description                 |
|--------|--------|-----------------------------|
| `email`| string | The email of the user.      |
| `pwd`  | string | The password of the user.   |

### Example
```json
{
    "email": "123@123.com",
    "pwd": "123456"
}
```

## Response

### Success (200 OK)
If the registration is successful, the response will be a JSON object with a status message and the user information.

#### Example
```json
{
    "err_msg": "ok", // this will return an error message otherwise
    "body": [{
        "email": "123@123.com",
        "pwd": "123456"
    }]
}
```

## Errors
- `400 Bad Request`: The request body is invalid or there was an error during registration.

# Login

## Endpoint
`POST /login`

## URI
`http://18.182.22.91:8080/login`

## Description
Check for email and password in the `auth` database, and verify if the password is correct.

## Request

### Headers
- `Content-Type: application/json`

### Body
The request body should contain a JSON object with the following fields:

| Field  | Type   | Description                 |
|--------|--------|-----------------------------|
| `email`| string | The email of the user.      |
| `pwd`  | string | The password of the user.   |

### Example
```json
{
    "email": "123@123.com",
    "pwd": "123456"
}
```

## Response

### Success (200 OK)
If the login is successful, the response will be a JSON object with a status message, user information, and user ID.

#### Example
```json
{
    "err_msg": "ok", // this will return an error message otherwise
    "body": [{
        "email": "123@123.com",
        "pwd": "123456"
    },
    123456 // id
    ]
}
```

## Errors
- `400 Bad Request`: The request body is invalid or the email/password combination is incorrect.

# Get All Profiles

## Endpoint
`GET /profiles`

## URI
`http://18.182.22.91:8080/profiles`

## Description
Returns a list of all existing profiles.

## Response

### Success (200 OK)
If the request is successful, the response will be a JSON object with a status message and a list of profiles.

#### Example
```json
{
    "err_msg": "ok", // this will return an error message otherwise
    "body": "[[...]]" // list of list of profiles
}
```

## Errors
- `400 Bad Request`: There was an error retrieving the profiles.

# Add New Profile

## Endpoint
`POST /profiles`

## URI
`http://18.182.22.91:8080/profiles`

## Description
Add a `profile` to the `auth` database.

## Request

### Headers
- `Content-Type: application/json`

### Body
The request body should contain a JSON object with the following fields:

| Field | Type   | Description                                                                                        |
|-------|--------|----------------------------------------------------------------------------------------------------|
| `id`  | string | The ID linked to the auth database, this ID must also exist in the auth database.                  |
| `name`| string | The name of the user.                                                                              |
| `age` | int    | The age of the user.                                                                               |
| `bio` | string | A short biography of the user.                                                                     |
| `pfp` | string | URL to the profile picture of the user.                                                            |

### Example
```json
{
    "id": "21",
    "name": "brudda",
    "age": 21,
    "bio": "gegagededadedado",
    "pfp": "https://yt3.ggpht.com/GDXAFeIXqsN_EPSHvxa2fbg_Fy3iJr3PuTNhMAXYBjjtZde8i8IEsyidJfpCov_WOe5_6oVGxA=s88-c-k-c0x00ffffff-no-rj"
}
```

## Response

### Success (200 OK)
If the profile addition is successful, the response will be a JSON object with a status message and the profile information.

#### Example
```json
{
    "err_msg": "ok", // this will return an error message otherwise
    "body": [{
        "id": "21",
        "name": "brudda",
        "age": 21,
        "bio": "gegagededadedado",
        "pfp": "https://yt3.ggpht.com/GDXAFeIXqsN_EPSHvxa2fbg_Fy3iJr3PuTNhMAXYBjjtZde8i8IEsyidJfpCov_WOe5_6oVGxA=s88-c-k-c0x00ffffff-no-rj"
    }]
}
```

## Errors
- `400 Bad Request`: The request body is invalid or there was an error adding the profile.

# Put an Image to S3 Bucket

## Endpoint
`PUT /profiles/{filename}`

## URI
`https://3x4ub88a07.execute-api.ap-northeast-1.amazonaws.com/orbital/orbital-media/{filename}`

## Description
Add `{filename}` to the `orbital-media` S3 bucket.

## Request

### Headers
- `Content-Type: multipart/form-data`

### Body
- Attach a file.

## Response

### Success (200 OK)
If the file upload is successful, the response will be a status message.

## Errors
- `400 Bad Request`: There was an error uploading the file.

# Get an Image from S3 Bucket

## Endpoint
`GET /profiles/{filename}`

## URI
`https://3x4ub88a07.execute-api.ap-northeast-1.amazonaws.com/orbital/orbital-media/{filename}`

## Description
Get `{filename}` from the `orbital-media` S3 bucket.

## Response

### Success (200 OK)
The response body will be the requested file.

## Errors
`404 Not Found`: The file does not exist in the S3 bucket.


# Edit Profile API

## Endpoint
`PATCH /profile`

## Description
This endpoint allows the user to update their profile information.

## Request

### Headers
- `Content-Type: application/json`

### Body
The request body should contain a JSON object with the following fields:

| Field | Type   | Description                        |
|-------|--------|------------------------------------|
| `id`  | string | The unique identifier of the user. |
| `name`| string | The name of the user.              |
| `age` | int    | The age of the user.               |
| `bio` | string | A short biography of the user.     |
| `pfp` | string | URL to the profile picture of the user. |

### Example
```json
{
    "id": "123",
    "name": "John Doe",
    "age": 30,
    "bio": "Software developer from NY",
    "pfp": "https://example.com/profile.jpg"
}
```

# Add Tag

## Endpoint
`PUT /tag`

## Description
Adds a new tag to the `tags` database.

## Request

### Headers
- `Content-Type: application/json`

### Body
The request body should contain a JSON object with the following fields:

| Field | Type   | Description              |
|-------|--------|--------------------------|
| `id`  | string | The unique identifier.   |
| `tag` | string | The tag to be added.     |

### Example
```json
{
    "id": "1",
    "tag": "exampleTag"
}
```

## Response

### Success (200 OK)
If the tag addition is successful, the response will be a JSON object with a status message and the tag information.

#### Example
```json
{
    "status": "ok",
    "data": [
        {
            "id": "1",
            "tag": "exampleTag"
        }
    ]
}
```

## Errors
- `400 Bad Request`: The request body is invalid or there was an error adding the tag.
- `Maximum amount of tags reached!` in the `err_msg` of the response json: The requested id have exceeded the maximum 
  allowed amount of tags (currently the maximum is 6)
- `Tag existed!` in the `err_msg` of the response json: the requested tag has already been added

# Query Tag

## Endpoint
`POST /tag/query`

## Description
Retrieves all tags associated with a given ID from the `tags` database.

## Request

### Headers
- `Content-Type: application/json`

### Body
The request body should contain a JSON object with the following field:

| Field | Type   | Description              |
|-------|--------|--------------------------|
| `id`  | string | The unique identifier.   |

### Example
```json
{
    "id": "1"
}
```

## Response

### Success (200 OK)
If the query is successful, the response will be a JSON object with a status message and a list of tags associated with the provided ID.

#### Example
```json
{
    "status": "ok",
    "data": [
        {
            "id": "1",
            "tag": "exampleTag1"
        },
        {
            "id": "1",
            "tag": "exampleTag2"
        }
    ]
}
```

## Errors
- `400 Bad Request`: The request body is invalid or there was an error querying the tags.

# Delete Tag

## Endpoint
`DELETE /tag`

## Description
Deletes a tag from the `tags` database based on the provided ID and tag.

## Request

### Headers
- `Content-Type: application/json`

### Body
The request body should contain a JSON object with the following fields:

| Field | Type   | Description              |
|-------|--------|--------------------------|
| `id`  | string | The unique identifier.   |
| `tag` | string | The tag to be deleted.   |

### Example
```json
{
    "id": "1",
    "tag": "exampleTag"
}
```

## Response

### Success (200 OK)
If the tag deletion is successful, the response will be a JSON object with a status message and the tag information that was deleted.

#### Example
```json
{
    "status": "ok",
    "data": [
        {
            "id": "1",
            "tag": "exampleTag"
        }
    ]
}
```

### Errors
- `400 Bad Request`: The request body is invalid or there was an error deleting the tag.
