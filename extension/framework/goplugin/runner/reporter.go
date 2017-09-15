// Copyright (C) 2017 NTT Innovation Institute, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runner

import (
	"fmt"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters/stenographer"
	"github.com/onsi/ginkgo/types"
)

const defaultStyle = "\x1b[0m"
const boldStyle = "\x1b[1m"
const redColor = "\x1b[91m"
const greenColor = "\x1b[32m"
const yellowColor = "\x1b[33m"
const cyanColor = "\x1b[36m"
const grayColor = "\x1b[90m"
const lightGrayColor = "\x1b[37m"

type Reporter struct {
	suites []types.SuiteSummary
	specs  []types.SpecSummary
}

const (
	configSuccinct = false
	configFullTrace = true
	configNoisyPendings = false
	configSlowSpecThreshold = float64(0.5) // sec
)

func (reporter *Reporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (reporter *Reporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *Reporter) SpecWillRun(specSummary *types.SpecSummary) {
}

func (reporter *Reporter) SpecDidComplete(specSummary *types.SpecSummary) {
	reporter.specs = append(reporter.specs, deepcopy.Copy(*specSummary).(types.SpecSummary))
}

func (reporter *Reporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *Reporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	for i, _ := range reporter.suites {
		if reporter.suites[i].SuiteDescription == summary.SuiteDescription {
			reporter.suites[i] = deepcopy.Copy(*summary).(types.SuiteSummary)
			break
		}
	}
}

func (reporter *Reporter) Prepare(description string) {
	reporter.suites = append(reporter.suites, types.SuiteSummary{
		SuiteDescription: description,
		SuiteSucceeded:   true,
		SuiteID:          "undefined",
		NumberOfSpecsBeforeParallelization: 0,
		NumberOfTotalSpecs:                 0,
		NumberOfSpecsThatWillBeRun:         0,
		NumberOfPendingSpecs:               0,
		NumberOfSkippedSpecs:               0,
		NumberOfPassedSpecs:                0,
		NumberOfFailedSpecs:                0,
		RunTime:                            time.Duration(0),
	})
}

func (reporter *Reporter) Report() {
	fmt.Println("--------------------------------------------------------------------------------")

	fmt.Println("Failures:")
	fmt.Println()

	steno := stenographer.New(true)

	for _, spec := range reporter.specs {
		steno.AnnounceCapturedOutput(spec.CapturedOutput)

		switch spec.State {
		case types.SpecStatePassed:
			if spec.IsMeasurement {
				steno.AnnounceSuccesfulMeasurement(&spec, configSuccinct)
			} else if spec.RunTime.Seconds() >= configSlowSpecThreshold {
				steno.AnnounceSuccesfulSlowSpec(&spec, configSuccinct)
			} else {
				steno.AnnounceSuccesfulSpec(&spec)
			}

		case types.SpecStatePending:
			steno.AnnouncePendingSpec(&spec, configNoisyPendings && !configSuccinct)
		case types.SpecStateSkipped:
			steno.AnnounceSkippedSpec(&spec, configSuccinct, configFullTrace)
		case types.SpecStateTimedOut:
			steno.AnnounceSpecTimedOut(&spec, configSuccinct, configFullTrace)
		case types.SpecStatePanicked:
			steno.AnnounceSpecPanicked(&spec, configSuccinct, configFullTrace)
		case types.SpecStateFailed:
			steno.AnnounceSpecFailed(&spec, configSuccinct, configFullTrace)
		}
	}

	fmt.Println()
	fmt.Println("Report:")
	fmt.Println()

	// total
	totalNumberOfTotalSpecs := 0
	totalNumberOfPassedSpecs := 0
	totalNumberOfFailedSpecs := 0
	totalNumberOfSkippedSpecs := 0
	totalNumberOfPendingSpecs := 0
	totalRunTime := time.Duration(0) * time.Nanosecond

	// suites
	for index, suite := range reporter.suites {
		fmt.Printf("[%4d] ", index+1)
		if suite.NumberOfFailedSpecs == 0 {
			fmt.Printf(greenColor+"%-80s"+defaultStyle, suite.SuiteDescription)
		} else {
			fmt.Printf(redColor+"%-80s"+defaultStyle, suite.SuiteDescription)
		}
		fmt.Printf(cyanColor+"total: %-4d%8s"+defaultStyle, suite.NumberOfTotalSpecs, "")
		if suite.NumberOfFailedSpecs == 0 {
			fmt.Printf(greenColor+"passed: %-4d%8s"+defaultStyle, suite.NumberOfPassedSpecs, "")
		} else {
			fmt.Printf("passed: %-4d%8s", suite.NumberOfPassedSpecs, "")
		}
		if suite.NumberOfFailedSpecs > 0 {
			fmt.Printf(redColor+"failed: %-4d%8s"+defaultStyle, suite.NumberOfFailedSpecs, "")
		} else {
			fmt.Printf("failed: %-4d%8s", suite.NumberOfFailedSpecs, "")
		}
		if suite.NumberOfSkippedSpecs > 0 {
			fmt.Printf(yellowColor+"skipped: %-4d%8s"+defaultStyle, suite.NumberOfSkippedSpecs, "")
		} else {
			fmt.Printf("skipped: %-4d%8s", suite.NumberOfSkippedSpecs, "")
		}
		if suite.NumberOfPendingSpecs > 0 {
			fmt.Printf(yellowColor+"pending: %-4d%8s"+defaultStyle, suite.NumberOfPendingSpecs, "")
		} else {
			fmt.Printf("pending: %-4d%8s", suite.NumberOfPendingSpecs, "")
		}
		if suite.RunTime >= time.Duration(configSlowSpecThreshold * 1000) * time.Millisecond {
			fmt.Printf(redColor + "run time: %s\n" + defaultStyle, suite.RunTime)
		} else {
			fmt.Printf("run time: %s\n", suite.RunTime)
		}

		totalNumberOfTotalSpecs += suite.NumberOfTotalSpecs
		totalNumberOfPassedSpecs += suite.NumberOfPassedSpecs
		totalNumberOfFailedSpecs += suite.NumberOfFailedSpecs
		totalNumberOfSkippedSpecs += suite.NumberOfSkippedSpecs
		totalNumberOfPendingSpecs += suite.NumberOfPendingSpecs
		totalRunTime += suite.RunTime
	}

	fmt.Println()

	fmt.Printf("[----] ")
	fmt.Printf(yellowColor+"%-80s"+defaultStyle, "SUMMARY")
	fmt.Printf(cyanColor+"total: %-4d%8s"+defaultStyle, totalNumberOfTotalSpecs, "")
	fmt.Printf("passed: %-4d%8s", totalNumberOfPassedSpecs, "")
	fmt.Printf("failed: %-4d%8s", totalNumberOfFailedSpecs, "")
	fmt.Printf("skipped: %-4d%8s", totalNumberOfSkippedSpecs, "")
	fmt.Printf("pending: %-4d%8s", totalNumberOfPendingSpecs, "")
	fmt.Printf("run time: %s\n", totalRunTime)
}

func NewReporter() *Reporter {
	return &Reporter{
		suites: []types.SuiteSummary{},
	}
}
