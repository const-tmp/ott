# ott
Dead simple one-time tokens
```
// create store with 1 hour tokens with 1 hour autoremove period
store := NewStore(time.Hour, time.Hour)

// create new token
token := store.NewToken()

// pop token from store. if token not exists or expired, result is nil, false
token, ok := store.Pop(token.Data)

// remove all expired tokens
store.RemoveExpired()

// start background loop removing expired tokens
go store.RemoveExpiredLoop(context.TODO())
```
