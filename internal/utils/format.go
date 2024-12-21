package utils

import (
	"fmt"
	"os"
)

func FormatFileMode(mode os.FileMode) string {
	return fmt.Sprintf("%O (%s)", uint32(mode), mode.String())
}
