package vcgencmd

import (
	"bytes"
	"os/exec"
	"pi_temperature_monitor/model"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Subcmd string

// supported commands
const (
	GET_THROTTLED = "get_throttled"
	MEASURE_TEMP  = "measure_temp"
)

type Vcgencmd struct {
	subcmd []Subcmd
	Result model.Value
}

func NewCmd() *Vcgencmd {
	return &Vcgencmd{
		subcmd: []Subcmd{},
	}
}

func (v *Vcgencmd) SetSubcmd(cmd Subcmd) *Vcgencmd {
	v.subcmd = append(v.subcmd, cmd)
	return v
}

func (v *Vcgencmd) Run() {

	for _, sc := range v.subcmd {
		var stdout, stderr bytes.Buffer

		cmd := exec.Command("vcgencmd", string(sc))
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			errors.Wrap(err, stderr.String())
			model.NewErrorEntry(err).Save()
			return
		}

		op := stdout.String()
		switch sc {
		case MEASURE_TEMP:
			{
				tempStr := strings.Split(op, "=")[1]
				temperature := strings.Split(tempStr, "'")[0]
				t, _ := strconv.ParseFloat(temperature, 64)
				model.NewValueEntry(&model.Value{
					Temperature: &t,
				}).Save()
			}
		case GET_THROTTLED:
			{
				throttleVal := strings.Split(op, "0x")[1]
				t, _ := strconv.ParseInt(throttleVal, 16, 64)
				model.NewValueEntry(&model.Value{
					Throttle: model.NewThrottleFromInt(int(t)),
				}).Save()
			}
		default:
			model.NewErrorEntry(errors.New("unkown cmd " + string(sc))).Save()
		}
	}
}
