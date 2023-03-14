package cmd

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type WorkflowDefinitionParser struct {
}

type workflowDefinition struct {
	Name          string
	Trigger       *TriggerDefinition
	rawDefinition *workflowDefinitionInt
}

type ValidationError struct {
	message string
}

func (validationError ValidationError) Error() string {
	return validationError.message
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
	definition := workflowDefinition{Name: parsedYaml.Name, Trigger: trigger, rawDefinition: &parsedYaml}
	return &definition, nil
}

func (workflow workflowDefinition) GetJobNames() ([]string, error) {
	jobs := workflow.rawDefinition.Jobs
	var jobNames []string
	for jobKey, jobDescription := range jobs {
		jobName := getJobName(jobKey, &jobDescription)
		if jobDescription.Strategy != nil && len(jobDescription.Strategy.Matrix) != 0 {
			jobNamesForThisJob, err := fillJobNameParametersForMatrixBuild(jobDescription, jobName)
			if err != nil {
				return nil, err
			}
			jobNames = append(jobNames, jobNamesForThisJob...)
		} else {
			jobNames = append(jobNames, jobName)
		}
	}
	return jobNames, nil
}

func fillJobNameParametersForMatrixBuild(jobDescription JobDescriptionInt, jobName string) (jobNames []string, err error) {
	matrix := jobDescription.Strategy.Matrix
	keys := make([]string, 0, len(matrix))
	for key := range matrix {
		keys = append(keys, key)
	}
	if strings.Contains(jobName, "${{") {
		jobNames = append(jobNames, replaceParametersInJobName(jobName, matrix, keys, 0)...)
	} else {
		if len(keys) == 1 {
			filledNames, err := addParametersToJobName(jobName, matrix[keys[0]])
			if err != nil {
				return nil, err
			}
			jobNames = append(jobNames, filledNames...)
		} else {
			return nil, ValidationError{"multi dimensional matrix github-action jobs with no explicit name are not supported. Please add a name field to the job that combines the matrix parameters into a more readable name. For example \"Build with Go ${{matrix.go}} and Exasol ${{ matrix.db }}\""}
		}
	}
	return jobNames, nil
}

func addParametersToJobName(jobName string, parameterValues []interface{}) (result []string, err error) {
	for _, value := range parameterValues {
		_, isMap := value.(map[interface{}]interface{})
		if isMap {
			return nil, ValidationError{"matrix github-action jobs with object parameters and no job name are not supported. Please add a name field to the job that combines the matrix parameters into a more readable name. For example \"Build with Go ${{matrix.go}} and Exasol ${{ matrix.db }}\""}
		}
		extendedJobName := jobName + " (" + convertValueToString(value) + ")"
		result = append(result, extendedJobName)
	}
	return result, nil
}

func replaceParametersInJobName(jobName string, matrix map[string][]interface{}, keys []string, parameterCursor int) (result []string) {
	if parameterCursor >= len(keys) {
		result = append(result, jobName)
	} else {
		key := keys[parameterCursor]
		pattern, err := regexp.Compile("\\${\\{\\s*matrix.\\Q" + key + "\\E\\s*\\}\\}")
		if err != nil {
			panic(err)
		}
		for _, value := range matrix[key] {
			filledJobName := replaceSpecificParameterInJobName(jobName, value, pattern)
			result = append(result, replaceParametersInJobName(filledJobName, matrix, keys, parameterCursor+1)...)
		}
	}
	return result
}

func replaceSpecificParameterInJobName(jobName string, value interface{}, pattern *regexp.Regexp) string {
	switch value := value.(type) {
	case string:
		return pattern.ReplaceAllString(jobName, value)
	case float64:
		return pattern.ReplaceAllString(jobName, fmt.Sprintf("%.1f", value))
	case int:
		return pattern.ReplaceAllString(jobName, fmt.Sprintf("%d", value))
	case bool:
		return pattern.ReplaceAllString(jobName, fmt.Sprintf("%t", value))
	case map[interface{}]interface{}:
		filledJobName := jobName
		for objectKey, objectValue := range value {
			objectKeyString := objectKey.(string)
			objectValueString := convertValueToString(objectValue)
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

func convertValueToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func getJobName(jobKey string, jobDescription *JobDescriptionInt) string {
	if jobDescription.Name != nil {
		return *jobDescription.Name
	} else {
		return jobKey
	}
}

type TriggerDefinition struct {
	TriggerOnPr              bool
	TriggerOnPushToBranches  []string
	TriggerOnPushToAnyBranch bool
}

func getTriggersOfWorkflowDefinition(parsedYaml *workflowDefinitionInt) (*TriggerDefinition, error) {
	if triggersFromArray := tryReadingTriggersFromArray(parsedYaml.On); triggersFromArray != nil {
		return triggersFromArray, nil
	} else if triggersFromMap := tryReadingTriggersFromMap(parsedYaml.On); triggersFromMap != nil {
		return triggersFromMap, nil
	} else {
		return nil, fmt.Errorf("the GitHub workflow '%v' uses an unsupported trigger definition style: %v", parsedYaml.Name, parsedYaml.On)
	}
}

func tryReadingTriggersFromArray(parsedYaml interface{}) *TriggerDefinition {
	triggersAsList, ok := parsedYaml.([]interface{})
	if ok && triggersAsList != nil {
		var result TriggerDefinition
		for _, trigger := range triggersAsList {
			lowerTrigger := strings.ToLower(trigger.(string))
			if lowerTrigger == "push" {
				result.TriggerOnPushToAnyBranch = true
			} else if lowerTrigger == "pull_request" {
				result.TriggerOnPr = true
			}
		}
		return &result
	}
	return nil
}

func tryReadingTriggersFromMap(parsedYaml interface{}) *TriggerDefinition {
	triggersAsMap, ok := parsedYaml.(map[string]interface{})
	if ok {
		var result TriggerDefinition
		for trigger, triggerParams := range triggersAsMap {
			lowerTrigger := strings.ToLower(trigger)
			if lowerTrigger == "pull_request" {
				result.TriggerOnPr = true
			} else if lowerTrigger == "push" {
				branches := readBranchesList(triggerParams)
				if branches == nil {
					result.TriggerOnPushToAnyBranch = true
				} else {
					result.TriggerOnPushToBranches = append(result.TriggerOnPushToBranches, branches...)
				}
			}
		}
		return &result
	}
	return nil
}

func readBranchesList(triggerParams interface{}) []string {
	triggerParamMap, ok := triggerParams.(map[interface{}]interface{})
	if !ok {
		return nil
	}
	branches := triggerParamMap["branches"]
	if branches == nil {
		return nil
	}
	branchesList, ok := branches.([]interface{})
	if !ok {
		return nil
	}
	var result = make([]string, 0, len(branchesList))
	for _, branch := range branchesList {
		result = append(result, branch.(string))
	}
	return result
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
