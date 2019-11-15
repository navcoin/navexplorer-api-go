package framework

var parameters = make(map[string]interface{})

func GetParameter(name string, defaultValue interface{}) interface{} {
	if parameters[name] == nil {
		return defaultValue
	}

	return parameters[name]
}

func SetParameter(name string, value interface{}) {
	parameters[name] = value
}
