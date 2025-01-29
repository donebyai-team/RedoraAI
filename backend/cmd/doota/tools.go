package main

import (
	. "github.com/streamingfast/cli"
)

var ToolsGroup = Group("tools", "Doota admin & developer tools",
	toolsPTSGroup,
	toolsPTSSyncCmd,
)
