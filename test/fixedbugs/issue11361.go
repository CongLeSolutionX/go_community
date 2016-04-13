// errorcheck

package a
import "fmt" // ERROR "imported and not used"
const n = fmt // ERROR "fmt without selector" "fmt is not a constant"
