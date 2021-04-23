package main

type MyContext struct {
	AuthService   AuthService
	UserStore     UserStore
	PasswordStore PasswordStore
}

func newMyContextForTest() *MyContext {
	userStore := newUserStoreImpl()
	passwordStore := newPasswordStoreImpl()
	authService := newAuthServiceImpl()
	authService.SetUserStore(userStore)
	authService.SetPasswordStore(passwordStore)
	return &MyContext{
		UserStore:     userStore,
		PasswordStore: passwordStore,
		AuthService:   authService,
	}
}

func newMyContext() *MyContext {
	options := getDatabaseConnectionVariables()
	db := CreateDatabaseConnection(options)
	userStore := newUserStoreDBImpl(db)
	passwordStore := newPasswordStoreDBImpl(db)
	authService := newAuthServiceImpl()
	authService.SetUserStore(userStore)
	authService.SetPasswordStore(passwordStore)
	return &MyContext{
		UserStore:     userStore,
		PasswordStore: passwordStore,
		AuthService:   authService,
	}
}
