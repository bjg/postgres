package database

type Model interface {
	GetInstance() interface{}
}
