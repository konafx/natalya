package util

import "fmt"

func ToUser(id string) string {
	return fmt.Sprintf("<@%s>", id)
}

func ToChannel(id string) string {
	return fmt.Sprintf("<#%s>", id)
}

func ToRole(id string) string {
	return fmt.Sprintf("<@&%s>", id)
}
