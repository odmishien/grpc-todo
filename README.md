# grpctodo

gRPCを使ってtodoアプリを作った回


## サーバ

```
go run main.go
```

`localhost:50051` に立つ

## Firestoreエミュレータ

```
firebase emulators:start --only firestore --project grpctodo
```

## Firebase Authentication 匿名ログイン

```
curl "https://www.googleapis.com/identitytoolkit/v3/relyingparty/signupNewUser?key=${API_KEY}" -H 'Content-Type: application/json' -d '{"returnSecureToken": true }'
```

idトークンが返ってくるので控えておく

## API叩いてみる

```
grpcurl -plaintext -d '{"task": "wash dishes"}' -H 'authorization: Bearer ${YOUR_ID_TOKEN}' localhost:50051 todo.todoService/addTodo
{
  "id": "...",
  "task": "wash dishes",
  "userId": "..."
}
```

```
grpcurl -plaintext -H 'authorization: Bearer ${YOUR_ID_TOKEN}' localhost:50051 todo.todoService/getTodos
{
  "todos": [
    {
      "id": "...",
      "task": "風呂掃除",
      "userId": "..."
    },
    {
      "id": "...",
      "task": "トイレ掃除",
      "userId": "..."
    }
  ]
}
```
