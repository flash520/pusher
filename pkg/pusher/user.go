/**
 * @Author: koulei
 * @Description:
 * @File: user
 * @Version: 1.0.0
 * @Date: 2023/9/7 08:55
 */

package pusher

type User interface {
	User() interface{}
	First() bool
	Write(msg Message)
	SetUser(user interface{})
	SetFirst(first bool)
	Close()
}

type userInfo struct {
	user   interface{}
	first  bool
	closed bool
	msg    chan<- Message
}

func (u *userInfo) User() interface{} {
	return u.user
}

func (u *userInfo) SetUser(user interface{}) {
	u.user = user
}

func (u *userInfo) First() bool {
	return u.first
}

func (u *userInfo) SetFirst(first bool) {
	u.first = first
}

func (u *userInfo) Write(msg Message) {
	if u.closed {
		return
	}
	select {
	case u.msg <- msg:
	default:
		return
	}
}

func (u *userInfo) Close() {
	u.closed = true
}
