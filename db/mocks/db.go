// Copyright (C) 2015 NTT Innovation Institute, Inc.
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

// Automatically generated by MockGen. DO NOT EDIT!
// Source: db.go

package mocks

import (
	transaction "github.com/cloudwan/gohan/db/transaction"
	schema "github.com/cloudwan/gohan/schema"
	gomock "github.com/golang/mock/gomock"
)

// Mock of DB interface
type MockDB struct {
	ctrl     *gomock.Controller
	recorder *_MockDBRecorder
}

// Recorder for MockDB (not exported)
type _MockDBRecorder struct {
	mock *MockDB
}

func NewMockDB(ctrl *gomock.Controller) *MockDB {
	mock := &MockDB{ctrl: ctrl}
	mock.recorder = &_MockDBRecorder{mock}
	return mock
}

func (_m *MockDB) EXPECT() *_MockDBRecorder {
	return _m.recorder
}

func (_m *MockDB) Connect(_param0 string, _param1 string, _param2 int) error {
	ret := _m.ctrl.Call(_m, "Connect", _param0, _param1, _param2)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockDBRecorder) Connect(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Connect", arg0, arg1, arg2)
}

func (_m *MockDB) Begin() (transaction.Transaction, error) {
	ret := _m.ctrl.Call(_m, "Begin")
	ret0, _ := ret[0].(transaction.Transaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockDBRecorder) Begin() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Begin")
}

func (_m *MockDB) RegisterTable(_param0 *schema.Schema, _param1 bool) error {
	ret := _m.ctrl.Call(_m, "RegisterTable", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockDBRecorder) RegisterTable(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RegisterTable", arg0, arg1)
}

func (_m *MockDB) DropTable(_param0 *schema.Schema) error {
	ret := _m.ctrl.Call(_m, "DropTable", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockDBRecorder) DropTable(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DropTable", arg0)
}

func (_m *MockDB) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockDBRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}
