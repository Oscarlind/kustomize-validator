package validate

import "fmt"

type Carrier struct {
	Path   string
	Stdout string
	Stderr string
	Err    error
}

func (c Carrier) Msg(errorOnly bool, verbose bool) string {
	msg := ""
	if c.Err != nil {
		msg += Errorf("Error while executing kustomize in path: %s, %s", c.Path, c.Err)
		if verbose {
			msg += Unknownf("==> Stdout (%s):\n%s", c.Path, c.Stdout)
			msg += Unknownf("==> Stderr (%s):\n%s", c.Path, c.Stderr)
		}
	} else {
		if !errorOnly {
			msg += Okf("Successfully executed kustomize on %s", c.Path)
			if verbose {
				msg += Unknownf("==> Stdout (%s):\n%s", c.Path, c.Stdout)
				msg += Unknownf("==> Stderr (%s):\n%s", c.Path, c.Stderr)
			}
		}
	}
	return msg
}

// ------

type Color string

const (
	ColorGreen  Color = "\033[0;32m"
	ColorRed    Color = "\033[0;31m"
	ColorYellow Color = "\033[0;33m"
	ColorOrange Color = "\033[0;33m"
	ColorBlue   Color = "\033[0;34m"
	ColorNC     Color = "\033[0m"
	ColorNone   Color = ""
)

type reason string

const (
	ReasonOK      reason = "OK"
	ReasonInfo    reason = "INFO"
	ReasonWarning reason = "WARNING"
	ReasonError   reason = "ERROR"
	ReasonUnknown reason = "UNKNOWN"
)

// printfC is a helper function to print out colored messages
func printfC(r reason, c Color, msg string) string {
	return fmt.Sprintf("[%s%s%s]: %s\n", c, r, ColorNC, msg)
}

func printf(r reason, msg string) string {
	switch r {
	case ReasonOK:
		return printfC(r, ColorGreen, msg)
	case ReasonInfo:
		return printfC(r, ColorBlue, msg)
	case ReasonWarning:
		return printfC(r, ColorOrange, msg)
	case ReasonError:
		return printfC(r, ColorRed, msg)
	default:
		return fmt.Sprintf("%s\n", msg)
	}
}

func Errorf(format string, a ...interface{}) string {
	return printf(ReasonError, fmt.Sprintf(format, a...))
}

func Warningf(format string, a ...interface{}) string {
	return printf(ReasonWarning, fmt.Sprintf(format, a...))
}

func Okf(format string, a ...interface{}) string {
	return printf(ReasonOK, fmt.Sprintf(format, a...))
}

func Infof(format string, a ...interface{}) string {
	return printf(ReasonInfo, fmt.Sprintf(format, a...))
}

func Unknownf(format string, a ...interface{}) string {
	return printf(ReasonUnknown, fmt.Sprintf(format, a...))
}

func ColorF(color Color, format string, a ...any) string {
	return fmt.Sprintf("%s%s%s", color, fmt.Sprintf(format, a...), ColorNC)
}
