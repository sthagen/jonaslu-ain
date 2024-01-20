package parse

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

var executableExpressionRe = regexp.MustCompile(`(m?)\$\([^)]*\)?`)
var executableRe = regexp.MustCompile(`\$\(([^)]*)\)`)

type executableAndArgs struct {
	executable string
	args       []string
}

type executableOutput struct {
	output       string
	fatalMessage string
}

func getExecutableExpr(templateLine string) ([]string, string) {
	retVal := []string{}

	templateLineRunes := []rune(templateLine)

	idx := 0
	for idx < len(templateLineRunes) {
		substr := string(templateLineRunes[idx:])

		if !strings.HasPrefix(substr, "$(") {
			idx += 1
			continue
		}

		startIdx := idx
		idx += 2

		var quoteRune rune
		var quoteStartIdx int

		done := false

		for !done && idx < len(templateLineRunes) {
			currChar := templateLineRunes[idx]

			switch currChar {
			case '"', '\'':
				prevRune := templateLineRunes[idx-1]
				if prevRune != '\\' && quoteRune == currChar {
					quoteRune = 0
				} else if quoteRune == 0 {
					quoteRune = currChar
					quoteStartIdx = idx
				}

			case ')':
				if quoteRune == 0 {
					done = true
				}
			}

			idx += 1
		}

		if quoteRune > 0 {
			return retVal, fmt.Sprintf("Unterminated quote sequence: %s", utils.Ellipsize(quoteStartIdx-3, quoteStartIdx+4, templateLine))
		}

		if !done {
			return retVal, fmt.Sprintf("Missing end parenthesis on executable: %s", utils.Ellipsize(startIdx-3, startIdx+4, templateLine))
		}

		retVal = append(retVal, string(templateLineRunes[startIdx:idx]))
	}

	return retVal, ""
}

func (s *sectionedTemplate) captureExecutableAndArgs() []executableAndArgs {
	executables := []executableAndArgs{}

	for expandedTemplateLineIndex, expandedTemplateLine := range s.expandedTemplateLines {
		noCommentsLineContents, _, _ := strings.Cut(expandedTemplateLine.LineContents, "#")

		for _, executableWithParens := range executableExpressionRe.FindAllString(noCommentsLineContents, -1) {
			executableAndArgsCapture := executableRe.FindStringSubmatch(executableWithParens)

			if len(executableAndArgsCapture) != 2 {
				s.setFatalMessage("Malformed executable", expandedTemplateLineIndex)
				continue
			}

			executableAndArgsStr := executableAndArgsCapture[1]
			if executableAndArgsStr == "" {
				s.setFatalMessage("Empty executable", expandedTemplateLineIndex)
				continue
			}

			tokenizedExecutableLine, err := utils.TokenizeLine(executableAndArgsStr)
			if err != nil {
				s.setFatalMessage(err.Error(), expandedTemplateLineIndex)
				continue
			}

			executable := tokenizedExecutableLine[0]

			executables = append(executables, executableAndArgs{
				executable: executable,
				args:       tokenizedExecutableLine[1:],
			})
		}
	}

	return executables
}

func callExecutables(ctx context.Context, config data.Config, executables []executableAndArgs) []executableOutput {
	executableResults := make([]executableOutput, len(executables))

	wg := sync.WaitGroup{}
	for i, executable := range executables {
		go func(resultIndex int, executable executableAndArgs) {
			defer wg.Done()

			var stdout, stderr bytes.Buffer

			cmd := exec.CommandContext(ctx, executable.executable, executable.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if ctx.Err() == context.DeadlineExceeded {
				executableResults[resultIndex].fatalMessage = fmt.Sprintf("Executable %s timed out after %d seconds", cmd.String(), config.Timeout)
			}

			if ctx.Err() != nil {
				return
			}

			stdoutStr := stdout.String()

			if err != nil {
				stderrStr := stderr.String()

				executableOutput := ""
				if stdoutStr != "" || stderrStr != "" {
					executableOutput = "\n" + strings.TrimSpace(strings.Join([]string{
						strings.TrimSpace(stdoutStr),
						strings.TrimSpace(stderrStr),
					}, " "))
				}

				executableResults[resultIndex].fatalMessage = fmt.Sprintf("Executable %s error: %v%s", cmd.String(), err, executableOutput)
				return
			}

			if stdoutStr == "" {
				executableResults[resultIndex].fatalMessage = fmt.Sprintf("Executable %s\nCommand produced no stdout output", cmd.String())
				return
			}

			executableResults[resultIndex].output = stdoutStr
		}(i, executable)

		wg.Add(1)
	}

	wg.Wait()

	return executableResults
}

func (s *sectionedTemplate) insertExecutableOutput(executableResults *[]executableOutput) {
	newExpandedTemplateLines := []expandedSourceMarker{}

	for expandedTemplateLineIndex, expandedTemplateLine := range s.expandedTemplateLines {
		lineContents := expandedTemplateLine.LineContents
		noCommentsLineContents, _, _ := strings.Cut(lineContents, "#")

		anythingReplaced := false

		for _, executableWithParens := range executableExpressionRe.FindAllString(noCommentsLineContents, -1) {
			result := (*executableResults)[0]
			*executableResults = (*executableResults)[1:]
			if result.fatalMessage != "" {
				s.setFatalMessage(result.fatalMessage, expandedTemplateLineIndex)
				continue
			}

			lineContents = strings.Replace(lineContents, executableWithParens, result.output, 1)
			anythingReplaced = true
		}

		if !anythingReplaced {
			newExpandedTemplateLines = append(newExpandedTemplateLines, expandedTemplateLine)
			continue
		}

		splitExpandedLines := strings.Split(strings.ReplaceAll(lineContents, "\r\n", "\n"), "\n")
		for _, splitExpandedLine := range splitExpandedLines {
			newExpandedTemplateLines = append(newExpandedTemplateLines, expandedSourceMarker{
				sourceMarker: sourceMarker{
					LineContents:    splitExpandedLine,
					SourceLineIndex: expandedTemplateLine.SourceLineIndex,
				},
				expanded: true,
			})
		}
	}

	s.expandedTemplateLines = newExpandedTemplateLines
}
