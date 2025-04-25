package mpv

import (
	"errors"
	"fmt"
	"os/exec"
)

type MpvWrapper struct {
	provider Provider
}

func NewMpvWrapper(provider Provider) *MpvWrapper {
	return &MpvWrapper{
		provider: provider,
	}
}

func (w MpvWrapper) Play(url, title string) error {
	switch w.provider {
	case ANIMEGG:
		cmdStr := fmt.Sprintf(`mpv --http-header-fields="Referrer: https://www.animegg.org/" '%s' --force-media-title="%s"`, url, title)
		cmd := exec.Command("sh", "-c", cmdStr)
		return cmd.Run()
	default:
		return errors.New("provider not supported")
	}
}
