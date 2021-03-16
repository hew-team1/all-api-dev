# api-dev

## How to Use

### 環境を構築する
```console
$make compose_build
```

### 環境を立ち上げる
```console
$make compose_start
```

### Dynamo-local Adminにアクセスする
```
localhost:8008
```

## アクセス
### EndUserAPI
#### POST  [登録]
[値へ](#post--登録-2)
```
http://localhost:60001/users
```

#### GET  [全件取得]
[値へ](#get--全件取得-4)
```
http://localhost:60001/users
```

#### GET  [投稿中の取得]
[値へ](#get--投稿中の取得-1)
```
http://localhost:60001/users/in-posts
```

#### GET  [参加中の取得]
[値へ](#get--参加中の取得-1)
```
http://localhost:60001/users/in-join
```

---

### RecruitAPI
#### POST  [登録]
[値へ](#post--登録-3)
```
http://localhost:60002/recruits
```

#### GET  [全件取得]
[値へ](#get--全件取得-5)
```
http://localhost:60002/recruits
```

#### GET  [idのrecruit取得]
[値へ](#get--idのrecruit取得-1)
```
http://localhost:60002/recruits/{id}
```

#### PUT  [idの募集の参加メンバーの追加]
[値へ](#put--idの募集の参加メンバーの追加-1)
```
http://localhost:60002/recruits/{id}/members
```

---

### ConnpassAPI
#### GET  [直近2ヶ月のハッカソンデータ取得]
[値へ](#get--直近2ヶ月のハッカソンデータ取得-1)
```
http://localhost:60003/connpass
```

---

### Admin EndUserAPI
#### GET  [全件取得]
[値へ](#get--全件取得-6)
```
http://localhost:60011/users
```

#### PUT  [isActiveの変更・アカウント停止の操作]
[値へ](#put--isactiveの変更アカウント停止の操作-1)
```
http://localhost:600011/users/active
```

---

### Admin RecruitAPI
#### GET  [全件取得]
[値へ](#get--全件取得-7)
```
http://localhost:60012/recruits
```

#### PUT  [isActiveの変更・ボード停止の操作]
[値へ](#put--isactiveの変更ボード停止の操作-1)
```
http://localhost:60012/recruits/active
```

---

## APIの値
### EndUserAPI
#### POST  [登録]
```
// リクエスト
{
  "uid":   string, // 必須
  "name":  stirng, // 必須
  "email": string, // 必須
}
```

#### GET  [全件取得]
```
// レスポンス
[
  {
    "uid":   string,
    "name":  string,
    "email": string,
  },
  {}, ...
]
```

#### GET  [投稿中の取得]
```
// リクエスト　[header]
key: uid
value: ユーザーID

// レスポンス
[
  {
    "id":          int,
    "masterId":    string,
    "title":       string,
    "eventDay":    string,
    "day":         string,   
    "organizer":   string,
    "commit":      string,
    "beginner":    stirng,
    "message":     string,
    "slackUrl":    string,
    "totalMember": string,   
    "position":    string,
    "reword":      string,
    "members": [
      {"uid": string, "position": string},
      {}, ...
    ],
    "created":  string,
    "updated":  string,
  },
  {}, ...
]
```

#### GET  [参加中の取得]
```
// リクエスト　[header]
key: uid
value: ユーザーID

// レスポンス
[
  {
    "id":          int,
    "masterId":    string,
    "title":       string,
    "eventDay":    string,
    "day":         string,   
    "organizer":   string,
    "commit":      string,
    "beginner":    stirng,
    "message":     string,
    "slackUrl":    string,
    "totalMember": string,   
    "position":    string,
    "reword":      string,
    "members": [
      {"uid": string, "position": string},
      {}, ...
    ],
    "created":  string,
    "updated":  string,
  },
  {}, ...
]
```


---

### RecruitAPI
#### POST  [登録]
```
// リクエスト
{
  "masterId":    string, // 必須
  "title":       string, // 必須
  "eventDay":    string, // 必須
  "day":         string, // 必須
  "organizer":   string, // 必須
  "commit":      string, // 必須
  "beginner":    stirng, // 必須
  "message":     string, // 必須
  "slackUrl":    string, // 必須
  "totalMember": string, // 必須
  "position":    string, // 必須
  "reword":      string, // 必須
}
```

#### GET  [全件取得]
```
// レスポンス
[
  {
    "id":          int,
    "masterId":    string,
    "title":       string,
    "eventDay":    string,
    "day":         string,   
    "organizer":   string,
    "commit":      string,
    "beginner":    stirng,
    "message":     string,
    "slackUrl":    string,
    "totalMember": stirng,   
    "position":    string,
    "reword":      string,
    "members": [
      {"uid": string, "position": string},
      {}, ...
    ],
    "created": string,
    "updated": string,
  },
  {}, ...
]
```

#### GET  [idのrecruit取得]
```
// レスポンス
{
  "id":          int,
  "masterId":    string,
  "title":       string,
  "eventDay":    string,
  "day":         string,
  "organizer":   string,
  "commit":      string,
  "beginner":    stirng,
  "message":     string,
  "slackUrl":    string,
  "totalMember": string,
  "position":    string,
  "reword":      string,
  "members": [
    {"uid": string, "position": string},
    {}, ...
  ],
  "created": string,
  "updated": string,
}
```

#### PUT  [idの募集の参加メンバーの追加]
```
// リクエスト
{
  "uid":      int,    // 必須
  "position": string, // 必須
}
```

---

### ConnpassAPI
#### GET  [直近2ヶ月のハッカソンデータ取得]

```
// レスポンス
[
  {
    "event_id":   int,
    "event_url":  string,
    "title":      string,
    "started_at": string,
    "ended_at":   string,
  },
  {}, ...
]
```

---

### Admin EndUserAPI
#### GET  [全件取得]

```
// レスポンス
[
  {
    "uid":      string,
    "name":     string,
    "email":    string,
    "created":  string,
    "updated":  string,
    "isLogin":  bool,
    "isActive": bool,
  },
  {}, ...
]
```

#### PUT  [isActiveの変更・アカウント停止の操作]

```
// リクエスト
{
  "uid":      string, // 必須
  "isActive": bool,   // 必須
}

```

---

### Admin RecruitAPI
#### GET  [全件取得]
```
// レスポンス
[
  {
    "id":          int,
    "masterId":    string,
    "title":       string,
    "eventDay":    string,
    "day":         string,   
    "organizer":   string,
    "commit":      string,
    "beginner":    stirng,
    "message":     string,
    "slackUrl":    string,
    "totalMember": string,   
    "position":    string,
    "members": [
      {"uid": string, "position": string},
      {}, ...
    ],
    "created":  string,
    "updated":  string,
    "isActive": bool,
  },
  {}, ...
]
```

#### PUT  [isActiveの変更・ボード停止の操作]
```
// リクエスト
{
  "id":       int,  // 必須
  "isActive": bool, // 必須
}
```