/**
 * @Author: koulei
 * @Description:
 * @File: metadata
 * @Version: 1.0.0
 * @Date: 2023/9/14 09:42
 */

package pusher

type metadata struct {
	source string
}

func (m *metadata) Source() string {
	return m.source
}

func (m *metadata) SetSource(source string) {
	m.source = source
}
