package dependency

var typeInitializerMap = make(map[ContainerType]func() interface{}, 0)

func RegisterGenerator(name ContainerType, fn func() interface{}) {
	typeInitializerMap[name] = fn
}

func GetInstance(name ContainerType) interface{} {
	return typeInitializerMap[name]()
}
