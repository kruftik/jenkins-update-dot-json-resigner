package jenkins_update_center

var (
	jsonSymbolReplacementsMap = []jsonSymbolReplacementRuleT{
		{[]byte("\\u0026"), []byte("&")},
		{[]byte("\\u003c/"), []byte("<\\/")},
		{[]byte("\\u003c"), []byte("<")},
		{[]byte("\\u003e"), []byte(">")},
	}

	JenkinsUCJSON JenkinsUCJSONT
)
