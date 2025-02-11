package svc

type RegistryOption func(*Registry)

func ExecDbMigration(migrate bool) RegistryOption {
	return func(reg *Registry) {
		reg.execDbMigrations = migrate
	}
}
