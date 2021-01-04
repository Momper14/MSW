package wrapper

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/looplab/fsm"
	"github.com/momper14/msw/wrapper/model"
	"github.com/momper14/viperfix"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var config struct {
	Heapsize   int
	Jar        string
	Workingdir string
	Eula       bool
	JvmArgs    []string
	ServerArgs []string
}

// inits viper
func init() {
	viper.SetDefault("mc.heapsize", 4096)
	viper.SetDefault("mc.jar", "server.jar")
	viper.SetDefault("mc.workingdir", "server")
	viper.SetDefault("mc.eula", false)

	if err := viperfix.UnmarshalKey("mc", &config); err != nil {
		logrus.Fatal(err)
	}
}

// Wrapper for the Minecraft Server
type Wrapper struct {
	console  *console
	machine  *fsm.FSM
	commands chan *model.Command
	subs     []chan *model.Message
}

// NewWrapper initialises a new Wrapper
func NewWrapper() *Wrapper {
	wrapper := &Wrapper{
		console:  nil,
		commands: make(chan *model.Command),
	}
	wrapper.machine = fsm.NewFSM(
		ServerOffline.String(),
		fsm.Events{
			fsm.EventDesc{
				Name: StopEvent.String(),
				Src:  []string{ServerOnline.String()},
				Dst:  ServerStopping.String(),
			},
			fsm.EventDesc{
				Name: StoppedEvent.String(),
				Src:  []string{ServerStopping.String(), ServerStarting.String(), ServerOnline.String()},
				Dst:  ServerOffline.String(),
			},
			fsm.EventDesc{
				Name: StartEvent.String(),
				Src:  []string{ServerOffline.String()},
				Dst:  ServerStarting.String(),
			},
			fsm.EventDesc{
				Name: StartedEvent.String(),
				Src:  []string{ServerStarting.String()},
				Dst:  ServerOnline.String(),
			},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { wrapper.enterState(e) },
		},
	)

	return wrapper
}

// enterState callpack for state change of the state machine
func (w *Wrapper) enterState(e *fsm.Event) {
	w.publish(&model.Message{
		Type:    model.TypeState,
		Payload: e.Dst,
	})
}

// publish publishes messages to all subscribers
func (w *Wrapper) publish(msg *model.Message) {
	for _, s := range w.subs {
		s <- msg
	}
}

// Subscribe subscribes to the MSW
func (w *Wrapper) Subscribe(sub chan *model.Message) chan *model.Command {
	w.subs = append(w.subs, sub)

	return w.commands
}

// processLogEvents processes log events from the Minecraft Server
func (w *Wrapper) processLogEvents() {
	for {
		line, err := w.console.ReadLine()
		if err == io.EOF {

			if err := w.updateState(StoppedEvent); err != nil {
				logrus.Warn(err)
			}
			break
		}

		if err != nil {
			logrus.Errorln(err)
		}

		w.publish(&model.Message{
			Type:    model.TypeLog,
			Payload: line,
		})

		ll, err := parseToLogLine(line)
		if err == nil {
			logToConsole(ll)
			if err := w.updateState(ll.toEvent()); err != nil {
				logrus.Error(err)
			}
		} else {
			logrus.Info(line)
		}
	}
}

// processErrEvents processes error events from the Minecraft Server
func (w *Wrapper) processErrEvents() {
	for {
		line, err := w.console.ReadErr()
		if err == io.EOF {
			break
		}

		if err != nil {
			logrus.Error(err)
		}
		w.publishErr(line)
	}
}

// logToConsole logs a log from the Minecraft Server to the console
func logToConsole(ll *logLine) {
	var fn func(...interface{})

	switch ll.level {
	case "INFO":
		fn = logrus.Info
	case "WARN":
		fn = logrus.Warn
	case "ERROR":
		fn = logrus.Error
	default:
		fn = logrus.Print
	}

	fn(ll.output)
}

// publishLog publishes a log
func (w *Wrapper) publishLog(line string) {
	logrus.Info(line)
	w.publish(&model.Message{
		Type:    model.TypeLog,
		Payload: line,
	})
}

// publishErr publishes a error
func (w *Wrapper) publishErr(line string) {
	logrus.Error(line)
	w.publish(&model.Message{
		Type:    model.TypeError,
		Payload: line,
	})
}

// updateState updates the Minecraft Server state
func (w *Wrapper) updateState(ev Event) error {
	if ev == EmptyEvent {
		return nil
	}
	return w.machine.Event(ev.String())
}

// CurrentState returns the current state of the Minecraft server
func (w *Wrapper) CurrentState() ServerState {
	return ServerStateFor(w.machine.Current())
}

// IsOffline returns if the Minecraft Server is offline
func (w *Wrapper) IsOffline() bool {
	return w.CurrentState() == ServerOffline
}

// WaitUntilOffline waits until the Minecraft Server is Offline
// if timeout is 0, its ignored
func (w *Wrapper) WaitUntilOffline(timeout time.Duration) error {

	var inTime = true
	var sleep = 100 * time.Millisecond

	if timeout != 0 {
		sleep = timeout / 100

		timer := time.NewTimer(timeout)
		go func() {
			<-timer.C
			inTime = false
		}()
	}

	for inTime {
		if w.IsOffline() {
			return nil
		}
		time.Sleep(sleep)
	}

	if w.IsOffline() {
		return nil
	}

	return fmt.Errorf("timeout")
}

// processCommands processes commands from the commands channel
func (w *Wrapper) processCommands() {
	for command := range w.commands {
		var err error
		target := command.Target
		payload := command.Payload

		if !target.Validate() {
			logrus.Warnf("invalid target %s", target)
		}

		logrus.Infof("recieved command \"%s\"\n", payload)
		w.publishLog(payload)

		switch target {
		case model.TargetServer:
			err = w.console.WriteCmd(payload)

		case model.TargetWrapper:
			switch payload {
			case "start":
				if w.CurrentState() != ServerOnline {
					err = w.Start()
				} else {
					w.publishLog("server already running!")
				}
			case "restart":
				err = w.Restart()
			case "stop":
				cs := w.CurrentState()
				if cs == ServerStarting || cs == ServerOnline {
					err = w.Stop()
				} else {
					w.publishLog("server not running!")
				}
			default:
				logrus.Warnf("unknown wrapper command: %s", payload)
			}
		}

		if err != nil {
			logrus.Error(err)
			w.publishLog(err.Error())
		}
	}
}

// eula sets the eula if mc.eula=true
func eula() {
	if config.Eula {
		f, err := os.OpenFile(config.Workingdir+"/eula.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			logrus.Error(err)
		}
		defer f.Close()

		if _, err := f.WriteString("eula=true"); err != nil {
			logrus.Warnf("Failed to write eula because of %s", err.Error())
		}
	}
}

// calculateArgs calculates the right memory values based on the heap size
func calculateArgs() string {
	heapSize := config.Heapsize
	jar := config.Jar
	nurseryMinimum := heapSize / 2
	nurseryMaximum := heapSize * 4 / 5
	args := fmt.Sprintf("-Xms%[1]dM -Xmx%[1]dM -Xmns%[2]dM -Xmnx%[3]dM -Xgc:concurrentScavenge -Xgc:dnssExpectedTimeRatioMaximum=3 -Xgc:scvNoAdaptiveTenure -Xdisableexplicitgc -Xtune:virtualized %[5]s -jar %[4]s nogui %[6]s",
		heapSize, nurseryMinimum, nurseryMaximum, jar, strings.Join(config.JvmArgs, " "), strings.Join(config.ServerArgs, " "))
	args = strings.Trim(args, " ")
	args = strings.ReplaceAll(args, "  ", " ")
	return args
}

// javaExecCmd setup the java exec
func javaExecCmd() *exec.Cmd {
	cmd := exec.Command("java", strings.Split(calculateArgs(), " ")...)
	cmd.Dir = config.Workingdir
	logrus.Infof("Executing %s in %s", cmd.Args, cmd.Dir)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
		Setpgid:   true,
	}

	return cmd
}

// Run starts the Minecraft Server Wrapper
func (w *Wrapper) Run() error {
	go w.processCommands()
	return w.Start()
}

// Start starts the Minecraft Server and the event processing
func (w *Wrapper) Start() error {
	w.console = newConsole(javaExecCmd())

	eula()

	go w.processLogEvents()
	go w.processErrEvents()
	return w.console.Start()
}

// Stop stops the Minecraft Server
func (w *Wrapper) Stop() error {
	return w.console.WriteCmd("stop")
}

// Restart restarts the Minecraft Server
func (w *Wrapper) Restart() error {
	if w.CurrentState() == ServerOnline {
		if err := w.Stop(); err != nil {
			return err
		}

		if err := w.WaitUntilOffline(30 * time.Second); err != nil {
			return err
		}
	}

	return w.Start()
}
