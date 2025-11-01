package mousemover

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/resousse/activity-tracker/pkg/activity"
	"github.com/resousse/activity-tracker/pkg/tracker"
)

type TestMover struct {
	suite.Suite
	activityTracker *tracker.Instance
	heartbeatCh     chan *tracker.Heartbeat
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(TestMover))
}

// Run once before all tests
func (suite *TestMover) SetupSuite() {
	heartbeatInterval := 60
	workerInterval := 10

	suite.activityTracker = &tracker.Instance{
		HeartbeatInterval: heartbeatInterval,
		WorkerInterval:    workerInterval,
	}

	suite.heartbeatCh = make(chan *tracker.Heartbeat)
}

// Run once before each test
func (suite *TestMover) SetupTest() {
	instance = nil
}
func (suite *TestMover) TestAppStart() {
	t := suite.T()
	mouseMover := GetInstance()
	mouseMover.run(suite.heartbeatCh, suite.activityTracker)
	time.Sleep(time.Millisecond * 500) //wait for app to start
	assert.True(t, mouseMover.state.isRunning(), "app should have started")
}
func (suite *TestMover) TestSingleton() {
	t := suite.T()

	mouseMover1 := GetInstance()
	mouseMover1.run(suite.heartbeatCh, suite.activityTracker)

	time.Sleep(time.Millisecond * 500)

	mouseMover2 := GetInstance()
	assert.True(t, mouseMover2.state.isRunning(), "instance should have started")
}

func (suite *TestMover) TestLogFile() {
	t := suite.T()
	mouseMover := GetInstance()
	logFileName := "test1"

	getLogger(mouseMover, true, logFileName)

	filePath := logDir + "/" + logFileName
	assert.FileExists(t, filePath, "log file should exist")
	os.RemoveAll(filePath)
}
func (suite *TestMover) TestSystemSleepAndWake() {
	t := suite.T()
	mouseMover := GetInstance()

	state := &state{
		override: &override{
			valueToReturn: true,
		},
	}
	mouseMover.state = state
	heartbeatCh := make(chan *tracker.Heartbeat)

	mouseMover.run(heartbeatCh, suite.activityTracker)
	time.Sleep(time.Millisecond * 500) //wait for app to start
	assert.True(t, mouseMover.state.isRunning(), "instance should have started")
	assert.False(t, mouseMover.state.isSystemSleeping(), "machine should not be sleeping")

	//fake a machine-sleep activity
	machineSleepActivityMap := make(map[activity.Type][]time.Time)
	var sleepTimeArray []time.Time
	sleepTimeArray = append(sleepTimeArray, time.Now())
	machineSleepActivityMap[activity.MachineSleep] = sleepTimeArray
	heartbeatCh <- &tracker.Heartbeat{
		WasAnyActivity: true,
		ActivityMap:    machineSleepActivityMap,
	}
	time.Sleep(time.Millisecond * 500) //wait for it to be registered
	assert.True(t, mouseMover.state.isSystemSleeping(), "machine should be sleeping now")

	//assert app is sleeping
	heartbeatCh <- &tracker.Heartbeat{
		WasAnyActivity: false,
	}

	time.Sleep(time.Millisecond * 500) //wait for it to be registered
	assert.True(t, time.Time.IsZero(state.getLastMouseMovedTime()), "should be default but is ", state.getLastMouseMovedTime())
	assert.Equal(t, state.getDidNotMoveCount(), 0, "should be 0")

	//fake a machine-wake activity
	machineWakeActivityMap := make(map[activity.Type][]time.Time)
	var wakeTimeArray []time.Time
	wakeTimeArray = append(wakeTimeArray, time.Now())
	machineWakeActivityMap[activity.MachineWake] = wakeTimeArray
	heartbeatCh <- &tracker.Heartbeat{
		WasAnyActivity: true,
		ActivityMap:    machineWakeActivityMap,
	}

	time.Sleep(time.Millisecond * 500) //wait for it to be registered
	assert.False(t, mouseMover.state.isSystemSleeping(), "machine should be awake now")
}

func (suite *TestMover) TestMouseMoveSuccess() {
	t := suite.T()
	mouseMover := GetInstance()

	state := &state{
		override: &override{
			valueToReturn: true,
		},
	}
	mouseMover.state = state
	heartbeatCh := make(chan *tracker.Heartbeat)

	mouseMover.run(heartbeatCh, suite.activityTracker)
	time.Sleep(time.Millisecond * 500) //wait for app to start
	assert.True(t, state.isRunning(), "instance should have started")
	assert.False(t, state.isSystemSleeping(), "machine should not be sleeping")
	assert.True(t, time.Time.IsZero(state.getLastMouseMovedTime()), "should be default")
	assert.Equal(t, state.getDidNotMoveCount(), 0, "should be 0")

	heartbeatCh <- &tracker.Heartbeat{
		WasAnyActivity: false,
	}

	time.Sleep(time.Millisecond * 500) //wait for it to be registered
	assert.False(t, time.Time.IsZero(state.getLastMouseMovedTime()), "should be default but is ", state.getLastMouseMovedTime())
}

func (suite *TestMover) TestMouseMoveFailure() {
	t := suite.T()
	mouseMover := GetInstance()

	state := &state{
		override: &override{
			valueToReturn: false,
		},
	}
	mouseMover.state = state
	heartbeatCh := make(chan *tracker.Heartbeat)

	mouseMover.run(heartbeatCh, suite.activityTracker)
	time.Sleep(time.Millisecond * 500) //wait for app to start
	assert.True(t, state.isRunning(), "instance should have started")
	assert.False(t, state.isSystemSleeping(), "machine should not be sleeping")
	assert.True(t, time.Time.IsZero(state.getLastMouseMovedTime()), "should be default")
	assert.Equal(t, state.getDidNotMoveCount(), 0, "should be 0")
	assert.True(t, state.getLastErrorTime().IsZero(), "should be default")

	heartbeatCh <- &tracker.Heartbeat{
		WasAnyActivity: false,
	}

	time.Sleep(time.Millisecond * 500) //wait for it to be registered
	assert.True(t, time.Time.IsZero(state.getLastMouseMovedTime()), "should be default but is ", state.getLastMouseMovedTime())
	assert.NotEqual(t, state.getDidNotMoveCount(), 0, "should not be 0")
	assert.NotEqual(t, state.getLastErrorTime(), 0, "should not be 0")
}

func (suite *TestMover) TestGetInstanceReset() {
	t := suite.T()
	// ensure a fresh singleton is created after reset
	instance = nil
	m1 := GetInstance()
	assert.NotNil(t, m1, "GetInstance should return an instance")
	m2 := GetInstance()
	assert.Equal(t, m1, m2, "GetInstance should return the same singleton")
}

func (suite *TestMover) TestQuitIdempotentBuffered() {
	t := suite.T()
	mouseMover := GetInstance()
	mouseMover.state.updateRunningStatus(true)
	mouseMover.quit = make(chan struct{}, 1)

	tmpFile, err := os.CreateTemp("", "amm-log-*")
	assert.NoError(t, err, "should create temp file")
	mouseMover.logFile = tmpFile

	mouseMover.Quit()
	assert.False(t, mouseMover.state.isRunning(), "state should not be running after Quit")
	mouseMover.Quit()
	tmpFile.Close()
	os.Remove(tmpFile.Name())
}

func (suite *TestMover) TestRunPreventsMultipleStarts() {
	t := suite.T()
	mouseMover := GetInstance()

	// simulate already running
	mouseMover.state.updateRunningStatus(true)

	// calling run when already running should return immediately and not change running state
	mouseMover.run(suite.heartbeatCh, suite.activityTracker)
	assert.True(t, mouseMover.state.isRunning(), "state should remain running after calling run again")
}
