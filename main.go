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

	"github.com/antonbabenko/tfvars-annotations/util"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

var (
	version = flag.Bool("version", false, "print version information and exit")
	debug   = flag.Bool("debug", false, "enable debug logging")

	// Main filename to work with
	tfvarsFile = "terraform.tfvars"

	// Dir where terragrunt cache lives
	terragruntCacheDir = ".terragrunt-cache"

	// Create a new instance of the logger. You can have any number of instances.
	log = logrus.New()

	// Deliberately uninitialized. See below.
	buildVersion string

	_ = spew.Config

	err error
)

// versionInfo returns a string containing the version information of the
// current build. It's empty by default, but can be included as part of the
// build process by setting the main.buildVersion variable.
func versionInfo() string {
	if buildVersion != "" {
		return buildVersion
	}

	return "unknown"
}

func main() {
	flag.Parse()

	if *debug == true {
		log.Level = logrus.DebugLevel
	} else {
		log.Level = logrus.InfoLevel
	}

	if *version == true {
		fmt.Printf("%s version %s\n", os.Args[0], versionInfo())
		os.Exit(0)
	}

	// Relative path to original tfvars file
	tfvarsDir := flag.Arg(0)

	if tfvarsDir == "" {
		log.Errorf("Specify tfvars directory where %s is located", tfvarsFile)
		os.Exit(1)
	}

	if _, err = os.Stat(tfvarsDir); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// Full relative path to original tfvars file
	tfvarsFullpath := filepath.Join(tfvarsDir, tfvarsFile)

	// Relative path to destination ".terraform" working directory
	terraformWorkingDir := findWorkingDir(tfvarsDir)
	log.Infoln("Working dir: ", terraformWorkingDir)

	// Full relative path to destination tfvars file (inside .terragrunt-cache/.../.../.terraform)
	var terraformWorkingDirTfvarsFullPath = filepath.Join(terraformWorkingDir, tfvarsFile)

	// Map of all keys and values to replace in tfvars file
	allKeyValues := make(map[string]interface{})

	log.Infof("Processing file: %s", tfvarsFullpath)
	log.Println()

	tfvarsContent, err := readTfvarsFile(tfvarsFullpath)
	if err != nil {
		log.Fatalf("Can't read file: %s", err)
	}

	astf, err := parseContent(&tfvarsContent)
	if err != nil {
		log.Fatalln("Can't parse content as HCL", err)
	}

	isDisabled, keysToReplace := scanComments(astf)
	if isDisabled {
		log.Fatalf("Dynamic update has been disabled in %s. Nothing to do.", tfvarsFile)

	}

	if len(keysToReplace) == 0 {
		log.Infoln("There are no keys to replace")
	}

	for _, key := range keysToReplace {
		log.Infof("Key: %s", key)

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

		//if len(split) > 3 {
		//	convertToType = split[3]
		//}

		workDir := filepath.Join(tfvarsDir, "../", dirName)
		//fmt.Println(workDir)

		resultValue, resultType, errResult := getResultFromTerragruntOutput(workDir, outputName)

		if errResult != nil {
			log.Warnf("Can't update value of %s in %s because key \"%s\"", key, tfvarsFullpath, outputName)
			log.Warnf("Error from terragrunt:", errResult)
			log.Println()
		}

		_ = resultType
		_ = convertToType

		// @todo: add support for to_list
		//if convertToType == "to_list" {
		//	resultValue = fmt.Sprintf("[%s]", resultValue)
		//}

		allKeyValues[key] = resultValue

		log.Infof("Value: %s", spew.Sdump(resultValue))
		//log.Infof("Value: %s", formattedResultValue)
		log.Infoln()
		log.Infoln()

	}

	log.Debugln("All key values:")
	log.Debugln(spew.Sdump(allKeyValues))

	astfUpdated, err2 := updateValuesInTfvarsFile(astf, allKeyValues)

	if err2 != nil {
		log.Fatalf("%s: Can't replace all keys in %s", err2, tfvarsFullpath)
	}

	//spew.Dump(astfUpdated)

	tfvarsFullpathTmp := ""

	if !*debug {
		tfvarsFullpathTmp = tfvarsFullpath
	}

	hclFormatted, err := fprintToFile(&astfUpdated, tfvarsFullpathTmp)
	if err != nil {
		log.Fatalf("Can't fprint AST to file %s, Error: %s", tfvarsFullpathTmp, err)
	}

	log.Infoln("FINAL HCL:")
	log.Infoln(string(hclFormatted))
	_ = hclFormatted

	log.Infoln()
	log.Infof("Copying updated %s into %s", tfvarsFullpath, terraformWorkingDirTfvarsFullPath)
	log.Infoln()

	_, err = util.CopyFile(tfvarsFullpath, terraformWorkingDirTfvarsFullPath)

	if err != nil {
		log.Fatalf("%s: Can't copy file to %s", err, terraformWorkingDirTfvarsFullPath)
	}

	log.Infoln("Done!")

	os.Exit(0)
}

func findWorkingDir(tfvarsDir string) string {

	var workingDir string

	_ = filepath.Walk(tfvarsDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && strings.Contains(path, terragruntCacheDir) && len(workingDir) == 0 {

			// eg: examples/project1-terragrunt/eu-west-1/app/.terragrunt-cache/F0pCE6ytQ7SNCsEA3BS4Wg57FJs/w9zgoLbGjuT9Afe34Zp8rkEMzXI
			if matched, _ := regexp.MatchString(terragruntCacheDir+`/[^/]+/[^/]+$`, path); matched {
				workingDir = path
			}
		}
		return nil
	})

	return workingDir
}

func readTfvarsFile(tfvarsFullpath string) (string, error) {
	bytes, err := ioutil.ReadFile(tfvarsFullpath)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func getResultFromTerragruntOutput(dirName string, outputName string) (interface{}, string, error) {

	lsCmd := exec.Command("terragrunt", "output", "-json", outputName)
	//lsCmd := exec.Command("cat", outputName)
	lsCmd.Dir = dirName
	lsOut, err := lsCmd.Output()

	if err != nil {
		log.Debugln(spew.Sdump(lsCmd))

		return "", "", errors.Wrapf(err, "running terragrunt output -json %s", outputName)
	}

	//fmt.Println("terragrunt value = ", string(lsOut))

	// Unmarshal output into JSON
	var TerragruntOutput map[string]interface{}

	if err := json.Unmarshal(lsOut, &TerragruntOutput); err != nil {
		panic(err)
	}

	return TerragruntOutput["value"], TerragruntOutput["type"].(string), nil
}
