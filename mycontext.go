package main

type MyContext struct {
	AuthService   AuthService
	UserStore     UserStore
	PasswordStore PasswordStore
	RoomStore     RoomStore
	RoomProvider  RoomProvider
}

func newMyContextForTest() *MyContext {
	userStore := newUserStoreImpl()
	passwordStore := newPasswordStoreImpl()
	authService := newAuthServiceImpl()
	authService.SetUserStore(userStore)
	authService.SetPasswordStore(passwordStore)
	roomStore := NewRoomStoreImpl()
	roomProvider := NewRoomProviderImpl(roomStore)
	return &MyContext{
		UserStore:     userStore,
		PasswordStore: passwordStore,
		AuthService:   authService,
		RoomStore:     roomStore,
		RoomProvider:  roomProvider,
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
	roomStore := NewRoomStoreDatabaseImpl(db)
	roomProvider := NewRoomProviderImpl(roomStore)
	return &MyContext{
		UserStore:     userStore,
		PasswordStore: passwordStore,
		AuthService:   authService,
		RoomStore:     roomStore,
		RoomProvider:  roomProvider,
	}
}