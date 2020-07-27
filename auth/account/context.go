package account

import (
	"context"
)

type accountKeyStruct struct{}

var accountKey = &accountKeyStruct{}

// StoreAccountInContext stores the account in the context and returns it.
func StoreAccountInContext(ctx context.Context, account *Account) context.Context {
	return context.WithValue(ctx, accountKey, account)
}

// CtxGetAccount gets account from provided context.
func CtxGetAccount(ctx context.Context) (*Account, bool) {
	acc, ok := ctx.Value(accountKey).(*Account)
	return acc, ok
}
