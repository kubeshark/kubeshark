package dependency

var typeIntializerMap = make(map[DependencyContainerType]func() interface{}, 0)

func RegisterGenerator(name DependencyContainerType, fn func() interface{}) {
	typeIntializerMap[name] = fn
}

func GetInstance(name DependencyContainerType) interface{} {
	return typeIntializerMap[name]()
}
