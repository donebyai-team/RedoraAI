package utils

import "github.com/k3a/html2text"

func HTMLToText(in string) string {
	return html2text.HTML2Text(in)
}
