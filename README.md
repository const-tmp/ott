# ott
Dead simple one-time tokens
```
// create store with 1 hour tokens
store := NewStore(3600)

// create new token
token := store.NewToken()

// pop token from store. if token not exists or expired, result is nil, false
token, ok := store.Pop(token.Data)

// remove all expired tokens
store.RemoveExpired()

// start backgroud loop removing expired
go store.RemoveExpiredLoop(context.TODO())
```
