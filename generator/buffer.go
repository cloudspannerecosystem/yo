package generator

import (
	"bytes"
	"strings"
)

// TBuf is to hold the executed templates.
type TBuf struct {
	TemplateType TemplateType
	Name         string
	Subname      string
	Buf          *bytes.Buffer
}

// TBufSlice is a slice of TBuf compatible with sort.Interface.
type TBufSlice []TBuf

func (t TBufSlice) Len() int {
	return len(t)
}

func (t TBufSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TBufSlice) Less(i, j int) bool {
	if t[i].TemplateType < t[j].TemplateType {
		return true
	} else if t[j].TemplateType < t[i].TemplateType {
		return false
	}

	if strings.Compare(t[i].Name, t[j].Name) < 0 {
		return true
	} else if strings.Compare(t[j].Name, t[i].Name) < 0 {
		return false
	}

	return strings.Compare(t[i].Subname, t[j].Subname) < 0
}
