// generated by stringer -type=Kind; DO NOT EDIT

package span

import "fmt"

const _Kind_name = "KeyTooLongInvalidStartPageInvalidEndPageEndPageBeforeStartPageInvalidURLSuspiciousPageCount"

var _Kind_index = [...]uint8{0, 10, 26, 40, 62, 72, 91}

func (i Kind) String() string {
	if i < 0 || i+1 >= Kind(len(_Kind_index)) {
		return fmt.Sprintf("Kind(%d)", i)
	}
	return _Kind_name[_Kind_index[i]:_Kind_index[i+1]]
}