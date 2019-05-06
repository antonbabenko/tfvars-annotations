package main

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/antonbabenko/tfvars-annotations/util"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/sirupsen/logrus"

	//"github.com/hashicorp/hcl/hcl/token"
	"io/ioutil"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/rodaine/hclencoder"

	"github.com/davecgh/go-spew/spew"
)

var (
	// String marker
	tfvarsDisableAnnotations = `@tfvars:disable_annotations`

	// Regexp
	tfvarsTerragruntOutputRegexp = regexp.MustCompile(`@tfvars:terragrunt_output\.[^ \n]+`)

	_ = spew.Config
	_ = fmt.Sprint()
)

func parseContent(hclString *string) (*ast.File, error) {
	astf, err := hcl.Parse(*hclString)
	if err != nil {
		return nil, err
	}

	return astf, nil
}

func scanComments(astf *ast.File) (bool, []string) {
	comments := astf.Comments

	isDisabled := false
	keysFound := []string{}

	for _, commentList := range comments {
		for _, comment := range commentList.List {

			if strings.Contains(comment.Text, tfvarsDisableAnnotations) {
				isDisabled = true
			}

			allKeys := tfvarsTerragruntOutputRegexp.FindAllString(comment.Text, -1)

			keysFound = append(keysFound, allKeys...)
		}
	}

	sort.Strings(keysFound)

	keysFound = util.UniqueNonEmptyElementsOf(keysFound)

	log.Debugf("Found keys: %s", keysFound)

	return isDisabled, keysFound
}

func updateValuesInTfvarsFile(astf *ast.File, allKeyValues map[string]interface{}) (ast.File, []string) {

	var errors []string
	var hclContent string

	flag.Parse()

	if *debug == true {
		log.Level = logrus.TraceLevel
	} else {
		log.Level = logrus.InfoLevel
	}

	ast.Walk(astf.Node, func(n ast.Node) (ast.Node, bool) {
		if n == nil {
			return n, false
		}

		typeName := reflect.TypeOf(n).String()

		if typeName == "*ast.ObjectItem" {
			//log.Traceln("* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *")
			//log.Traceln("Node type=", typeName)

			//log.Traceln("Node=")
			//log.Traceln(spew.Sdump(n))

			leadCommentText := ""
			lineCommentText := ""

			//spew.Dump(n.(*ast.ObjectItem).Keys[0])

			leadComment := n.(*ast.ObjectItem).LeadComment
			if leadComment != nil {
				leadCommentText = leadComment.List[0].Text
			}

			lineComment := n.(*ast.ObjectItem).LineComment
			if lineComment != nil {
				lineCommentText = lineComment.List[0].Text
				log.Traceln("Found line comment:", lineCommentText)

				for key, value := range allKeyValues {

					if strings.Contains(lineCommentText, key) {

						//lineComment.List[0].Text = lineCommentText // + "!!!"

						currentVal := n.(*ast.ObjectItem).Val

						currentValPos := currentVal.Pos().Line
						//log.Traceln("Current line number of the value: ", currentValPos)

						valueType := reflect.TypeOf(value).String() // real value type
						//desiredValueType := reflect.TypeOf(value).String() // desired value type (to_list will set this to `list`)

						log.Tracef("Found line comment value to replace using value from key %s", key)

						log.Traceln(spew.Sdump(value))
						log.Traceln(spew.Sdump(valueType))

						hclBytes, err := hclencoder.Encode(value)

						if err != nil {
							log.Warnln("Error during hclencoder: ", err)
						}

						// Create HCL string with properly encoded values and with real number of newlines prefixed to be able to replace in current value
						prefixNewlines := strings.Repeat("\n", currentValPos-1)
						//hclString := strings.TrimSuffix(string(hclBytes), "\n")
						hclString := string(hclBytes)

						// @todo: Maps are not supported yet, because comments are placed in strange places
						if valueType == "map[string]interface {}" {
							//hclContent = prefixNewlines + `key = {` + hclString + `}`
							continue
						} else {

							split := strings.Split(key, ".")

							convertToType := ""

							if len(split) > 3 {
								convertToType = split[3]
							}

							if convertToType == "to_list" {
								hclContent = prefixNewlines + `key = [` + hclString + `]`
							} else {
								hclContent = prefixNewlines + `key = ` + hclString
							}

						}

						//log.Traceln("HCL content created from the new value to parse to AST: ", hclContent)
						astfNew, err2 := hcl.Parse(hclContent)

						if err2 != nil {
							log.Warnln(err2)
							continue
						}

						// Value from the first element item created earlier (new value)
						newVal := astfNew.Node.(*ast.ObjectList).Items[0].Val

						//log.Traceln("+ + + + + + + + + + + + + + + + + + + + + + + + + + + + + + + + +")
						//log.Traceln("astfNew.Node=", spew.Sdump(astfNew.Node))
						//log.Traceln("NEW VALUE:", spew.Sdump(newVal))
						//log.Traceln("NEW VALUE OFFSET:", spew.Sdump(newVal.Pos().Offset))
						//log.Traceln("OLD VALUE:", spew.Sdump(currentVal))
						//log.Traceln("OLD VALUE OFFSET:", spew.Sdump(currentVal.Pos().Offset))

						// Replacing old value
						n.(*ast.ObjectItem).Val = newVal
					}
				}

			}

			//spew.Dump(leadComment)
			//spew.Dump(lineComment)
			_ = leadCommentText
			_ = lineCommentText
		}

		return n, true
	})

	if false {
		log.Traceln("= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =")

		spew.Dump(astf.Comments)

		for i, commentGroup := range astf.Comments {
			for _, comment := range commentGroup.List {

				log.Traceln("ORIGINAL i=", i, "; offset=", commentGroup.Pos().Offset, "; commentOffset=", comment.Pos().Offset, spew.Sdump(comment))

				//lineNumber := comment.Pos().Line
				//var lineNumber, columnNumber, Offset int
				lineNumber := comment.Pos().Line
				columnNumber := commentGroup.Pos().Column
				Offset := commentGroup.Pos().Offset

				if i == 2 {
					lineNumber += 5
					//Offset += 10 // number of chars to shift?
				}

				if i == 4 {
					lineNumber += 5
					//Offset += 10 // number of chars to shift?
				}

				// @todo: Adjust Line to it bigger on the length of the previous block
				/*comment.Start = token.Pos{
					Filename: "",
					Offset:   Offset,
					Line:     lineNumber,
					Column:   columnNumber,
				}*/

				fmt.Println(lineNumber, columnNumber, Offset)

				log.Traceln("NEW i=", i, "; offset=", comment.Pos().Offset, spew.Sdump(comment))
			}
		}

		log.Traceln("= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =")
		spew.Dump(astf.Comments)

	}

	//log.Traceln("= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =")
	//log.Traceln("Complete astf after update:", spew.Sdump(astf))

	return *astf, errors
}

func fprintToFile(astf *ast.File, filename string) ([]byte, error) {

	var buf bytes.Buffer

	if err := printer.Fprint(&buf, astf); err != nil {
		return buf.Bytes(), err
	}

	// Add trailing newline to result to prevent from reformatting every time
	buf.WriteString("\n")
	if filename != "" {
		if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
			return buf.Bytes(), err
		}
	}

	return buf.Bytes(), nil
}
