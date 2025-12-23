// utils/banner.go
package utils

import "fmt"

func SuccessBanner(title, details string) string {
	return fmt.Sprintf(`===========================================================
%s
===========================================================
%s
===========================================================`, title, details)
}

func InfoBanner(title, details string) string {
	return fmt.Sprintf(`-----------------------------------------------------
%s
-----------------------------------------------------
%s
-----------------------------------------------------`, title, details)
}
