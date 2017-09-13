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
	"path/filepath"
	"plugin"
	"regexp"
	"testing"

	gohan_db "github.com/cloudwan/gohan/db"
	"github.com/cloudwan/gohan/db/options"
	"github.com/cloudwan/gohan/extension/goext"
	"github.com/cloudwan/gohan/extension/goplugin"
	logger "github.com/cloudwan/gohan/log"
	"github.com/cloudwan/gohan/schema"
	"github.com/cloudwan/gohan/server/middleware"
	"github.com/cloudwan/gohan/sync/noop"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	testDBFile = "test.db"
)

var log = logger.NewLogger()

// TestRunner is a test runner for go (plugin) extensions
type TestRunner struct {
	pluginFileNames []string
	verboseLogs     bool
	fileNameFilter  *regexp.Regexp
	workerCount     int
}

// NewTestRunner allocates a new TestRunner
func NewTestRunner(pluginFileNames []string, printAllLogs bool, testFilter string, workers int) *TestRunner {
	return &TestRunner{
		pluginFileNames: pluginFileNames,
		verboseLogs:     printAllLogs,
		fileNameFilter:  regexp.MustCompile(testFilter),
		workerCount:     workers,
	}
}

// Run runs go (plugin) test runner
func (testRunner *TestRunner) Run() error {
	for _, pluginFileName := range testRunner.pluginFileNames {
		if !testRunner.fileNameFilter.MatchString(pluginFileName) {
			continue
		}

		if err := testRunner.runSingle(pluginFileName); err != nil {
			return err
		}
	}

	return nil
}

func dbConnString(name string) string {
	return fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
}

func dbConnect(connString string) (gohan_db.DB, error) {
	return gohan_db.ConnectDB("sqlite3", connString, gohan_db.DefaultMaxOpenConn, options.Default())
}

func readSchemas(p *plugin.Plugin) ([]string, error) {
	fnRaw, err := p.Lookup("Schemas")

	if err != nil {
		return nil, fmt.Errorf("missing 'Schemas' export: %s", err)
	}

	fn, ok := fnRaw.(func() []string)

	if !ok {
		return nil, fmt.Errorf("invalid signature of 'Schemas' export")
	}

	return fn(), nil
}

func readBinaries(p *plugin.Plugin) ([]string, error) {
	fnRaw, err := p.Lookup("Binaries")

	if err != nil {
		return nil, fmt.Errorf("missing 'Binaries' export: %s", err)
	}

	fn, ok := fnRaw.(func() []string)

	if !ok {
		return nil, fmt.Errorf("invalid signature of 'Binaries' export")
	}

	return fn(), nil
}

func readTest(p *plugin.Plugin) (func(goext.IEnvironment), error) {
	fnRaw, err := p.Lookup("Test")

	if err != nil {
		return nil, fmt.Errorf("missing 'Test' export: %s", err)
	}

	testFn, ok := fnRaw.(func(goext.IEnvironment))

	if !ok {
		return nil, fmt.Errorf("invalid signature of 'Test' export")
	}

	return testFn, nil
}

func (testRunner *TestRunner) runSingle(pluginFileName string) error {
	log.Notice("Running Go (plugin) extensions test: %s", pluginFileName)

	// load plugin
	p, err := plugin.Open(pluginFileName)

	if err != nil {
		return fmt.Errorf("failed to open plugin: %s", err)
	}

	// read schemas
	schemas, err := readSchemas(p)

	if err != nil {
		return fmt.Errorf("failed to read schemas from '%s': %s", pluginFileName, err)
	}

	// get state
	path := filepath.Dir(pluginFileName)
	manager := schema.GetManager()
	//extManager := extension.GetManager()

	// load schemas
	for _, schemaPath := range schemas {
		if err = manager.LoadSchemaFromFile(path + "/" + schemaPath); err != nil {
			return fmt.Errorf("failed to load schema: %s", err)
		}
	}

	// get binaries
	binaries, err := readBinaries(p)

	if err != nil {
		return fmt.Errorf("failed to read binaries from '%s': %s", pluginFileName, err)
	}

	// connect db
	db, err := dbConnect(dbConnString(testDBFile))

	if err != nil {
		return fmt.Errorf("failed to connect db: %s", err)
	}

	// create env
	env := goplugin.NewEnvironment("go environment test", db, &middleware.FakeIdentity{}, noop.NewSync())

	//if err := extManager.RegisterEnvironment(binary, env); err != nil {
	//	return return fmt.Errorf("failed to register environment: %s", err)
	//}

	// load binaries
	afterEach := func() error {
		// reset DB
		err = gohan_db.InitDBWithSchemas("sqlite3", dbConnString(testDBFile), true, false, false)

		if err != nil {
			return fmt.Errorf("failed to init db: %s", err)
		}

		return nil
	}

	for _, binary := range binaries {
		_, err := env.Load(path+"/"+binary, afterEach)

		if err != nil {
			return fmt.Errorf("failed to load test binary: %s", err)
		}
	}

	// start env
	err = env.Start()

	if err != nil {
		log.Error("failed to start extension test dependant plugin: %s; error: %s", pluginFileName, err)
		return err
	}

	// get test
	test, err := readTest(p)

	if err != nil {
		return fmt.Errorf("failed to read schemas from '%s': %s", pluginFileName, err)
	}

	// prepare test
	test(env)

	// run test
	gomega.RegisterFailHandler(ginkgo.Fail)

	t := &testing.T{}

	passed := ginkgo.RunSpecs(t, "Go Extensions Test Suite")
	fmt.Println()

	log.Notice("Go (plugin) extension test finished: %s", pluginFileName)

	if !passed {
		return fmt.Errorf("go extensions test failed: %s", pluginFileName)
	}

	// stop env
	env.Stop()

	// clear state
	manager.ClearExtensions()
	schema.ClearManager()

	return nil
}
