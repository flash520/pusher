/**
 * @Author: koulei
 * @Description:
 * @File: data
 * @Version: 1.0.0
 * @Date: 2023/9/12 23:12
 */

package pusher

import (
	"strings"

	"github.com/flash520/pusher/pkg/utils"
)

type Data interface {
	ID() string
	Raw() interface{}
	MetaData() string
}

type data struct {
	Id       string
	raw      interface{}
	metaData string
}

func NewData(source string, msg interface{}) *data {
	return &data{
		Id:       utils.RandString(20),
		raw:      msg,
		metaData: source,
	}
}

func (d *data) ID() string {
	return d.Id
}

func (d *data) Raw() interface{} {
	return d.raw
}

func (d *data) MetaData() string {
	return strings.ToLower(d.metaData)
}
