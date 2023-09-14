/**
 * @Author: koulei
 * @Description:
 * @File: data
 * @Version: 1.0.0
 * @Date: 2023/9/12 23:12
 */

package pusher

import (
	"github.com/flash520/pusher/pkg/utils"
)

type Data interface {
	ID() string
	Raw() interface{}
	Metadata() Metadata
}

type data struct {
	id       string
	raw      interface{}
	metadata Metadata
}

type Metadata interface {
	Source() string
	SetSource(string)
}

func NewData(source string, msg interface{}) *data {
	return &data{
		id:       utils.RandString(20),
		raw:      msg,
		metadata: &metadata{source: source},
	}
}

func (d *data) ID() string {
	return d.id
}

func (d *data) Raw() interface{} {
	return d.raw
}

func (d *data) Metadata() Metadata {
	return d.metadata
}
