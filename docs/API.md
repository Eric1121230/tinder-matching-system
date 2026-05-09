# API Documentation

Base URL: `http://localhost:8080`

## 1) AddSinglePersonAndMatch

- **Method**: `POST`
- **Path**: `/api/v1/people/match`
- **Description**: 新增一位使用者，並立即為該使用者尋找所有可能的配對對象。

### Request body

```json
{
  "name": "Bob",
  "height": 180,
  "gender": "male",
  "wanted_dates": 2
}
```

### Response 200

```json
{
  "matches": [
    {
      "boy_name": "Bob",
      "girl_name": "Alice"
    }
  ]
}
```

### Response 400

```json
{
  "error_code": "INVALID_JSON",
  "message": "請求格式錯誤"
}
或
{
  "error_code": "ADD_PERSON_FAILED",
  "message": "wanted_dates must be positive" 
}

```

### Response 409

```json
{
  "error_code": "USER_ALREADY_EXISTS",
  "message": "該姓名已在系統中"
}

```

---

## 2) RemoveSinglePerson

- **Method**: `DELETE`
- **Path**: `/api/v1/people/{name}`
- **Description**: 從配對池中移除特定人員。

### Response 204

No body.

### Response 404

```json
{
  "error_code": "PERSON_NOT_FOUND",
  "message": "找不到該人員"
} 
```

---

## 3) QuerySinglePerson

- **Method**: `GET`
- **Path**: `/api/v1/people/{name}`
- **Description**: 查詢單一特定人員的詳細資料。

### Response 200

```json
{
  "name": "Amy",
  "height": 160,
  "gender": "female",
  "remaining_dates": 2
}
```

### Response 400

```json
{
  "error_code": "MISSING_NAME",
  "message": "姓名為必填欄位"
}
```

### Response 404

```json
{
  "error_code": "PERSON_NOT_FOUND",
  "message": "找不到該人員"
}
```

---

## 4) QueryPersonMatches

- **Method**: `GET`
- **Path**: `/api/v1/people/{name}/matches?top={N}`
- **Description**: 為特定人員回傳前 N 位合適的配對對象。

### Example

`GET /api/v1/people/Amy/matches?top=2`

### Response 200

```json
{
  "name": "Amy",
  "matches": [
    {
      "name": "Ben",
      "height": 180,
      "gender": "male",
      "remaining_dates": 1
    }
  ]
}
```

### Response 400

```json
{
  "error_code": "INVALID_TOP",
  "message": "top 必須為正整數"
}
```

### Response 404

```json
{
  "error_code": "PERSON_NOT_FOUND",
  "message": "找不到該人員或其配對資料"
}
```

---

## 5) QuerySinglePeople (List All Members)

- **Method**: `GET`
- **Path**: `/api/v1/people`
- **Description**: 列出目前系統中所有的單身成員。

### Response 200

```json
{
  "people": [
    {
      "name": "Amy",
      "height": 160,
      "gender": "female",
      "remaining_dates": 2
    }
  ]
}
```

