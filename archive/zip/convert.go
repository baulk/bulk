package zip

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func (e *Extractor) initializeEncoder(codepage string) error {
	switch codepage {
	case "GB18030":
		e.dec = simplifiedchinese.GB18030.NewDecoder()
	case "GBK", "936":
		e.dec = simplifiedchinese.GBK.NewDecoder()
	}
	return nil
}

func (e *Extractor) convertToUTF8(s string) string {
	if e.dec == nil {
		return s
	}
	i := bytes.NewReader([]byte(s))
	decoder := transform.NewReader(i, e.dec)
	content, _ := ioutil.ReadAll(decoder)
	return string(content)
}
