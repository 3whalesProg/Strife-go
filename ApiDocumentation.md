
# Документация к API

## Эндпоинты

### 1. Получить список друзей
**Метод:** GET  
**URL:** `/api/v1/friend/friends`  
**Заголовки:**
- `Authorization`: Bearer токен пользователя

**Ответ:**
```json
[
    {
        "id": 1,
        "email": "example1@example.com",
        "role": "user"
    },
    {
        "id": 2,
        "email": "example2@example.com",
        "role": "user"
    }
]
```

### 2. Отправить запрос на добавление в друзья
**Метод:** POST  
**URL:** `/api/v1/friend/request`  
**Заголовки:**
- `Authorization`: Bearer токен пользователя

**Тело запроса:**
```json
{
    "recipient_login": "exampleLogin"
}
```

**Ответ:**
```json
{
    "message": "Friend request sent"
}
```

### 3. Получить список запросов на дружбу
**Метод:** GET  
**URL:** `/api/v1/friend/reqlis`  
**Заголовки:**
- `Authorization`: Bearer токен пользователя

**Ответ:**
```json
[
    {
        "sender_id": 1,
        "recipient_id": 2
    },
    {
        "sender_id": 3,
        "recipient_id": 1
    }
]
```

### 4. Ответить на запрос на дружбу
**Метод:** POST  
**URL:** `/api/v1/friend/response`  
**Заголовки:**
- `Authorization`: Bearer токен пользователя

**Тело запроса:**
```json
{
    "accepted": true
}
```

**Ответ:**
```json
{
    "message": "Response processed"
}
```

### 5. Удалить запрос на дружбу
**Метод:** DELETE  
**URL:** `/api/v1/friend/drequest`  
**Заголовки:**
- `Authorization`: Bearer токен пользователя

**Тело запроса:**
```json
{
    "recipient_id": 2
}
```

**Ответ:**
```json
{
    "message": "Friend request deleted"
}
```

### 6. Удалить друга
**Метод:** DELETE  
**URL:** `/api/v1/friend/dfriend`  
**Заголовки:**
- `Authorization`: Bearer токен пользователя

**Тело запроса:**
```json
{
    "friend_id": 2
}
```

**Ответ:**
```json
{
    "message": "Friend removed"
}
```
