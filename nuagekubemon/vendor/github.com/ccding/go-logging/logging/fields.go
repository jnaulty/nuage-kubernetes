// Copyright 2013, Cong Ding. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// author: Cong Ding <dinggnu@gmail.com>

package logging

import (
	"github.com/kardianos/osext"
	"os"
	"path"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

// The struct for each log record
type record struct {
	level    Level
	seqid    uint64
	pathname string
	filename string
	module   string
	lineno   int
	funcname string
	thread   int
	process  int
	message  string
	time     time.Time
}

// This variable maps fields in recordArgs to relavent function signatures
var fields = map[string]func(*Logger, *record) interface{}{
	"name":      (*Logger).lname,     // name of the logger
	"seqid":     (*Logger).nextSeqid, // sequence number
	"levelno":   (*Logger).levelno,   // level number
	"levelname": (*Logger).levelname, // level name
	"created":   (*Logger).created,   // starting time of the logger
	"nsecs":     (*Logger).nsecs,     // nanosecond of the starting time
	"time":      (*Logger).time,      // record created time
	"timestamp": (*Logger).timestamp, // timestamp of record
	"rtime":     (*Logger).rtime,     // relative time since started
	"filename":  (*Logger).filename,  // source filename of the caller
	"pathname":  (*Logger).pathname,  // filename with path
	"module":    (*Logger).module,    // executable filename
	"lineno":    (*Logger).lineno,    // line number in source code
	"funcname":  (*Logger).funcname,  // function name of the caller
	"process":   (*Logger).process,   // process id
	"message":   (*Logger).message,   // logger message
}

var runtimeFields = map[string]bool{
	"name":      false,
	"seqid":     false,
	"levelno":   false,
	"levelname": false,
	"created":   false,
	"nsecs":     false,
	"time":      false,
	"timestamp": false,
	"rtime":     false,
	"filename":  true,
	"pathname":  true,
	"module":    false,
	"lineno":    true,
	"funcname":  true,
	"thread":    true,
	"process":   false,
	"message":   false,
}

// If it fails to get some fields with string type, these fields are set to
// errString value.
const errString = "???"

// getShortFuncName generates short function name.
func getShortFuncName(fname string) string {
	fns := strings.Split(fname, ".")
	return fns[len(fns)-1]
}

// genRuntime generates the runtime information, including pathname, function
// name, filename, line number.
func (r *record) genRuntime() {
	calldepth := 4
	pc, file, line, ok := runtime.Caller(calldepth)
	if ok {
		fname := runtime.FuncForPC(pc).Name()
		r.pathname = file
		r.funcname = getShortFuncName(fname)
		r.filename = path.Base(file)
		r.lineno = line
	} else {
		r.pathname = errString
		r.funcname = errString
		r.filename = errString
		// Here we uses -1 rather than 0, because the default value in
		// golang is 0 and we should know the value is uninitialized
		// or failed to get
		r.lineno = -1
	}
}

// genNonRuntime generates the non-runtime information, including sequential
// id and time.
func (r *record) genNonRuntime(logger *Logger) {
	r.seqid = atomic.AddUint64(&(logger.seqid), 1)
	r.time = time.Now()
}

// Logger name
func (logger *Logger) lname(r *record) interface{} {
	return logger.name
}

// Next sequence number
func (logger *Logger) nextSeqid(r *record) interface{} {
	return r.seqid
}

// Log level number
func (logger *Logger) levelno(r *record) interface{} {
	return int32(r.level)
}

// Log level name
func (logger *Logger) levelname(r *record) interface{} {
	return levelNames[r.level]
}

// File name of calling logger, with whole path
func (logger *Logger) pathname(r *record) interface{} {
	return r.pathname
}

// File name of calling logger
func (logger *Logger) filename(r *record) interface{} {
	return r.filename
}

// module name
func (logger *Logger) module(r *record) interface{} {
	module, _ := osext.Executable()
	return path.Base(module)
}

// Line number
func (logger *Logger) lineno(r *record) interface{} {
	return r.lineno
}

// Function name
func (logger *Logger) funcname(r *record) interface{} {
	return r.funcname
}

// Timestamp of starting time
func (logger *Logger) created(r *record) interface{} {
	return logger.startTime.UnixNano()
}

// RFC3339Nano time
func (logger *Logger) time(r *record) interface{} {
	return r.time.Format(logger.timeFormat)
}

// Nanosecond of starting time
func (logger *Logger) nsecs(r *record) interface{} {
	return logger.startTime.Nanosecond()
}

// Nanosecond timestamp
func (logger *Logger) timestamp(r *record) interface{} {
	return r.time.UnixNano()
}

// Nanoseconds since logger created
func (logger *Logger) rtime(r *record) interface{} {
	return r.time.Sub(logger.startTime).Nanoseconds()
}

// Process ID
func (logger *Logger) process(r *record) interface{} {
	r.process = os.Getpid()
	return r.process
}

// The log message
func (logger *Logger) message(r *record) interface{} {
	return r.message
}
