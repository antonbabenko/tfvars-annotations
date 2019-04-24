package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/antonbabenko/dynamic-tfvars/util"

	"github.com/pkg/errors"
	"github.com/vosmith/pancake"
)

var version = flag.Bool("version", false, "print version information and exit")

// Main filename to work with
var tfvarsFile = "terraform.tfvars"

var err error

// Deliberately uninitialized. See below.
var build_version string

// versionInfo returns a string containing the version information of the
// current build. It's empty by default, but can be included as part of the
// build process by setting the main.build_version variable.
func versionInfo() string {
	if build_version != "" {
		return build_version
	} else {
		return "unknown"
	}
}

func main() {
	flag.Parse()

	if *version == true {
		fmt.Printf("%s version %s\n", os.Args[0], versionInfo())
		os.Exit(0)
	}

	// Relative path to original tfvars file
	var tfvarsDir = flag.Arg(0)

	if _, err = os.Stat(tfvarsDir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Full relative path to original tfvars file
	var tfvarsFullpath = filepath.Join(tfvarsDir, tfvarsFile)

	// Relative path to destination ".terraform" working directory
	var terraformWorkingDir = findTerraformWorkingDir(tfvarsDir)

	// Full relative path to destination tfvars file (inside .terragrunt-cache/.../.../.terraform)
	var terraformWorkingDirTfvarsFullPath = filepath.Join(terraformWorkingDir, tfvarsFile)

	// Map of all keys and values to replace in tfvars file
	allKeyValues := make(map[string]string)

	fmt.Printf("Processing file: %s", tfvarsFullpath)
	fmt.Println()
	fmt.Println()

	tfvarsContent, isDisabled := checkIsDisabled(tfvarsFullpath)
	if isDisabled {
		fmt.Printf("Dynamic update has been disabled in %s. Nothing to do.", tfvarsFile)
		os.Exit(1)
	}

	// Find all keys to replace
	keysToReplace := findKeys(tfvarsContent)

	if keysToReplace == nil {
		fmt.Println("There are no keys to replace")
		os.Exit(1)
	}

	for _, key := range keysToReplace {
		fmt.Printf("Key: %s", key)
		fmt.Println()

		split := strings.Split(key, ".")

		dirName := ""
		outputName := ""
		convertToType := ""

		if len(split) == 0 {
			continue
		}

		if len(split) > 1 {
			dirName = split[1]
		}

		if len(split) > 2 {
			outputName = split[2]
		}

		if len(split) > 3 {
			convertToType = split[3]
		}

		workDir := filepath.Join(tfvarsDir, "../", dirName)
		//fmt.Println(workDir)

		resultValue, _, errResult := getResultFromTerragruntOutput(workDir, outputName)

		if errResult != nil {
			fmt.Printf("Can't update value of %s in %s because key \"%s\" does not exist in output", key, tfvarsFullpath, outputName)
			fmt.Println()
		}

		// Format value as proper JSON
		formattedResultValue, errResult := json.Marshal(resultValue)

		if errResult != nil {
			fmt.Println("error:", errResult)
		}

		resultValue = string(formattedResultValue)

		if convertToType == "to_list" {
			resultValue = fmt.Sprintf("[%s]", resultValue)
		}

		allKeyValues[key] = resultValue.(string)

		fmt.Printf("Value: %s", resultValue)
		fmt.Println()
		fmt.Println()

	}

	//fmt.Println(allKeyValues)

	err = replaceAllKeysInTfvarsFile(tfvarsFullpath, allKeyValues)

	if err != nil {
		fmt.Printf("%s: Can't replace all keys in %s", err, tfvarsFullpath)
		os.Exit(1)
	}

	fmt.Printf("Copying updated %s into %s", tfvarsFullpath, terraformWorkingDirTfvarsFullPath)
	fmt.Println()
	fmt.Println()

	_, err = util.CopyFile(tfvarsFullpath, terraformWorkingDirTfvarsFullPath)

	if err != nil {
		fmt.Printf("%s: Can't copy file to %s", err, terraformWorkingDirTfvarsFullPath)
		os.Exit(1)
	}

	fmt.Println("Done!")

	os.Exit(0)
}

func findTerraformWorkingDir(tfvarsDir string) string {

	var terraformDir string

	// from https://flaviocopes.com/go-list-files/
	// @todo: Make sure to find inside .terragrunt-cache folder to prevent from finding wrong .terraform directory
	err := filepath.Walk(tfvarsDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if matched, _ := regexp.MatchString(`\.terraform$`, path); matched {
				terraformDir = path
				// @todo: exit from walk function once directory name has been found
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return terraformDir
}

func checkIsDisabled(tfvarsFullpath string) (string, bool) {

	bytes, err := ioutil.ReadFile(tfvarsFullpath)
	if err != nil {
		return "", true
	}

	if regexp.MustCompile(`@modulestf:disable_values_updates`).Match(bytes) {
		return string(bytes), true
	}

	return string(bytes), false
}

func findKeys(tfvarsContent string) []string {

	allKeys := regexp.MustCompile(`@modulestf:terraform_output\.[^ \n]*`).FindAllStringSubmatch(tfvarsContent, -1)

	if len(allKeys) == 0 {
		return nil
	}

	flattenKeys, _ := pancake.Strings(allKeys)

	//sort.Strings(flattenKeys)

	flattenKeys = util.UniqueNonEmptyElementsOf(flattenKeys)

	return flattenKeys
}

func getResultFromTerragruntOutput(dirName string, outputName string) (interface{}, string, error) {

	// @todo: call terragrunt for real

	//lsCmd := exec.Command("terragrunt", "output", "-json", output_name)
	lsCmd := exec.Command("cat", outputName)
	lsCmd.Dir = dirName
	lsOut, err := lsCmd.Output()

	if err != nil {
		return "", "", errors.Wrapf(err, "running terragrunt output -json %s", outputName)
	}

	//fmt.Println(string(lsOut))

	var TerraformOutput map[string]interface{}

	if err := json.Unmarshal([]byte(lsOut), &TerraformOutput); err != nil {
		panic(err)
	}

	return TerraformOutput["value"], TerraformOutput["type"].(string), nil
}

func replaceAllKeysInTfvarsFile(tfvarsFullpath string, allKeyValues map[string]string) error {

	input, err := ioutil.ReadFile(tfvarsFullpath)
	if err != nil {
		return err
	}

	content := string(input)

	for key, value := range allKeyValues {
		regexpFindLine := fmt.Sprintf(`^(.+) =.+(\#)+(.*)%s(.*)`, key)
		replacement := fmt.Sprintf(`$1 = %s $2$3%s$4`, value, key)

		//fmt.Println(regexpFindLine)
		//fmt.Println("KEY======>>>> ", key, "==", value)

		content = regexp.MustCompilePOSIX(regexpFindLine).ReplaceAllString(content, replacement)
	}

	//fmt.Println("REPLACED ===", content)

	if err = ioutil.WriteFile(tfvarsFullpath, []byte(content), 0644); err != nil {
		return err
	}

	return nil
}
