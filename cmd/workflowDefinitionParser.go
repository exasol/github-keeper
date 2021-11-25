package cmd

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"reflect"
	"regexp"
	"strings"
)

type WorkflowDefinitionParser struct {
}

func (parser WorkflowDefinitionParser) ParseWorkflowDefinition(content string) (*workflowDefinition, error) {
	parsedYaml := workflowDefinitionInt{}
	err := yaml.Unmarshal([]byte(content), &parsedYaml)
	if err != nil {
		return nil, err
	}
	trigger, err := getTriggersOfWorkflowDefinition(&parsedYaml)
	if err != nil {
		return nil, err
	}
	var jobNames []string
	for jobKey, jobDescription := range parsedYaml.Jobs {
		jobName := parser.getJobName(jobKey, &jobDescription)
		if jobDescription.Strategy != nil && len(jobDescription.Strategy.Matrix) != 0 {
			jobNamesForThisJob, err := parser.fillJobNameParametersForMatrixBuild(jobDescription, jobName)
			if err != nil {
				return nil, err
			}
			jobNames = append(jobNames, jobNamesForThisJob...)
		} else {
			jobNames = append(jobNames, jobName)
		}
	}
	definition := workflowDefinition{Name: parsedYaml.Name, JobsNames: jobNames, Trigger: trigger}
	return &definition, nil
}

func (parser WorkflowDefinitionParser) fillJobNameParametersForMatrixBuild(jobDescription JobDescriptionInt, jobName string) (jobNames []string, err error) {
	matrix := jobDescription.Strategy.Matrix
	keys := make([]string, 0, len(matrix))
	for key := range matrix {
		keys = append(keys, key)
	}
	if strings.Contains(jobName, "${{") {
		jobNames = append(jobNames, parser.replaceParametersInJobName(jobName, matrix, keys, 0)...)
	} else {
		if len(keys) == 1 {
			filledNames, err := parser.addParametersToJobName(jobName, matrix[keys[0]])
			if err != nil {
				return nil, err
			}
			jobNames = append(jobNames, filledNames...)
		} else {
			return nil, fmt.Errorf("multi dimensional matrix github-action jobs with no explicit name are not supported. Please add a name field to the job that combines the matrix parameters into a more readable name. For example \"Build with Go ${{matrix.go}} and Exasol ${{ matrix.db }}\"")
		}
	}
	return jobNames, nil
}

func (parser WorkflowDefinitionParser) addParametersToJobName(jobName string, parameterValues []interface{}) (result []string, err error) {
	for _, value := range parameterValues {
		_, isMap := value.(map[interface{}]interface{})
		if isMap {
			return nil, fmt.Errorf("matrix github-action jobs with object parameters and no job name are not supported. Please add a name field to the job that combines the matrix parameters into a more readable name. For example \"Build with Go ${{matrix.go}} and Exasol ${{ matrix.db }}\"")
		}
		extendedJobName := jobName + " (" + parser.convertValueToString(value) + ")"
		result = append(result, extendedJobName)
	}
	return result, nil
}

func (parser WorkflowDefinitionParser) replaceParametersInJobName(jobName string, matrix map[string][]interface{}, keys []string, parameterCursor int) (result []string) {
	if parameterCursor >= len(keys) {
		result = append(result, jobName)
	} else {
		key := keys[parameterCursor]
		pattern, err := regexp.Compile("\\${\\{\\s*matrix.\\Q" + key + "\\E\\s*\\}\\}")
		if err != nil {
			panic(err)
		}
		for _, value := range matrix[key] {
			filledJobName := parser.replaceSpecificParameterInJobName(jobName, value, pattern)
			result = append(result, parser.replaceParametersInJobName(filledJobName, matrix, keys, parameterCursor+1)...)
		}
	}
	return result
}

func (parser WorkflowDefinitionParser) replaceSpecificParameterInJobName(jobName string, value interface{}, pattern *regexp.Regexp) string {
	switch value := value.(type) {
	case string:
		return pattern.ReplaceAllString(jobName, value)
	case map[interface{}]interface{}:
		filledJobName := jobName
		for objectKey, objectValue := range value {
			objectKeyString := objectKey.(string)
			objectValueString := parser.convertValueToString(objectValue)
			objectPattern, err := regexp.Compile("\\${\\{\\s*matrix.\\Q" + objectKeyString + "\\E\\s*\\}\\}")
			if err != nil {
				panic(err)
			}
			filledJobName = objectPattern.ReplaceAllString(filledJobName, objectValueString)
		}
		return filledJobName
	default:
		panic(fmt.Sprintf("unsupported type %v", reflect.TypeOf(value)))
	}
}

func (parser WorkflowDefinitionParser) convertValueToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (parser WorkflowDefinitionParser) getJobName(jobKey string, jobDescription *JobDescriptionInt) string {
	if jobDescription.Name != nil {
		return *jobDescription.Name
	} else {
		return jobKey
	}
}

func getTriggersOfWorkflowDefinition(parsedYaml *workflowDefinitionInt) ([]string, error) {
	if triggerMap, hasTriggerMap := parsedYaml.On.(map[interface{}]interface{}); hasTriggerMap {
		var result []string
		for trigger := range triggerMap {
			result = append(result, trigger.(string))
		}
		return result, nil
	} else if triggerList, hasTriggerList := parsedYaml.On.([]interface{}); hasTriggerList {
		var result []string
		for _, trigger := range triggerList {
			result = append(result, trigger.(string))
		}
		return result, nil
	} else {
		return nil, fmt.Errorf("the GitHub workflow '%v' has a unimplemented trigger definition style", parsedYaml.Name)
	}
}

type strategyDescriptionInt struct {
	Matrix map[string][]interface{} `yaml:"matrix"`
}

type JobDescriptionInt struct {
	Strategy *strategyDescriptionInt `yaml:"strategy"`
	Name     *string                 `yaml:"name"`
}

type workflowDefinitionInt struct {
	Name string                       `yaml:"name"`
	On   interface{}                  `yaml:"on"`
	Jobs map[string]JobDescriptionInt `yaml:"jobs"`
}
